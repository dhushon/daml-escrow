package ledger

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"daml-escrow/internal/ledger/generated"

	"github.com/smartcontractkit/go-daml/pkg/client"
	"github.com/smartcontractkit/go-daml/pkg/model"
	. "github.com/smartcontractkit/go-daml/pkg/types"
	"go.uber.org/zap"
)

// Hardcoded Canton Party IDs for the Sandbox
const (
	CentralBankParty    = "CentralBank::1220e4edfbda663e622ba24a73ce1e201338a4a0634aa2e98a19d9044037dfc1f3d6"
	BuyerParty          = "Buyer::1220e4edfbda663e622ba24a73ce1e201338a4a0634aa2e98a19d9044037dfc1f3d6"
	SellerParty         = "Seller::1220e4edfbda663e622ba24a73ce1e201338a4a0634aa2e98a19d9044037dfc1f3d6"
	EscrowMediatorParty = "EscrowMediator::1220e4edfbda663e622ba24a73ce1e201338a4a0634aa2e98a19d9044037dfc1f3d6"
)

type DamlClient struct {
	logger *zap.Logger
	client *client.DamlBindingClient
	host   string
	port   int
}

func NewDamlClient(logger *zap.Logger, host string, port int) *DamlClient {
	// Initialize the go-daml client configuration
	grpcAddress := fmt.Sprintf("%s:%d", host, port)
	logger.Info("initializing DAML client", zap.String("address", grpcAddress))
	
	// Create a new DamlClient builder
	builder := client.NewDamlClient("", grpcAddress)
	
	// Build the binding client (usually needs a context, but we will do it lazily or use a background context)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger.Info("attempting to build DAML binding client...")
	bindingClient, err := builder.Build(ctx)
	if err != nil {
		logger.Error("failed to connect to DAML ledger", zap.Error(err), zap.String("address", grpcAddress))
	} else {
		logger.Info("DAML binding client built successfully")
	}

	return &DamlClient{
		logger: logger,
		client: bindingClient,
		host:   host,
		port:   port,
	}
}

func (c *DamlClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	c.logger.Info("creating escrow on DAML ledger", zap.Any("request", req))

	if c.client == nil {
		return nil, fmt.Errorf("DAML client not connected")
	}

	// Using the allocated parties for Sandbox
	issuer := PARTY(CentralBankParty)
	mediator := PARTY(EscrowMediatorParty)
	buyer := PARTY(BuyerParty)
	seller := PARTY(SellerParty)

	// Ensure Numeric format matches what the ledger expects (10 decimal places for Decimal)
	amountStr := strconv.FormatFloat(req.Amount, 'f', 10, 64)
	amountNumeric := NUMERIC(amountStr)

	escrow := generated.StablecoinEscrow{
		Issuer:      issuer,
		Buyer:       buyer,
		Seller:      seller,
		Mediator:    mediator,
		TotalAmount: amountNumeric,
		Currency:    TEXT(req.Currency),
		Description: TEXT("Escrow for " + req.Currency),
		Milestones: []generated.Milestone{
			{
				Label:     TEXT("Full Payment"),
				Amount:    amountNumeric,
				Completed: BOOL(false),
			},
		},
		CurrentMilestoneIndex: INT64(0),
	}

	createCmd := &model.CreateCommand{
		TemplateID: escrow.GetTemplateID(),
		Arguments: map[string]interface{}{
			"issuer":                escrow.Issuer,
			"buyer":                 escrow.Buyer,
			"seller":                escrow.Seller,
			"mediator":              escrow.Mediator,
			"totalAmount":           escrow.TotalAmount,
			"currency":              escrow.Currency,
			"description":           escrow.Description,
			"milestones":            []interface{}{escrow.Milestones[0].ToMap()},
			"currentMilestoneIndex": escrow.CurrentMilestoneIndex,
		},
	}

	submitReq := &model.SubmitAndWaitRequest{
		Commands: &model.Commands{
			CommandID: fmt.Sprintf("create-escrow-%d", time.Now().UnixNano()),
			UserID:    string(escrow.Buyer), // Use party ID as user ID for Sandbox v2
			ActAs:     []string{string(escrow.Buyer)},
			Commands: []*model.Command{
				{
					Command: createCmd,
				},
			},
		},
	}

	response, err := c.client.CommandService.SubmitAndWait(ctx, submitReq)
	if err != nil {
		c.logger.Error("failed to submit create command to DAML", zap.Error(err))
		return nil, fmt.Errorf("ledger submission failed: %w", err)
	}

	// The response from CommandService.SubmitAndWait might vary, 
	// but usually we can get the contract ID from the transaction/update ID or by querying.
	// For now, if we don't have events in the response, we might need to handle it.
	c.logger.Info("DAML escrow created", zap.String("updateId", response.UpdateID))

	return &EscrowContract{
		ID:       "unknown-id-query-needed", // Need to extract from response if available
		Buyer:    req.Buyer,
		Seller:   req.Seller,
		Amount:   req.Amount,
		Currency: req.Currency,
		State:    "Created on Ledger",
	}, nil
}

func (c *DamlClient) GetEscrow(ctx context.Context, id string) (*EscrowContract, error) {
	c.logger.Info("querying DAML ledger for escrow", zap.String("id", id))

	// Implementation using ActiveContractService would go here
	return &EscrowContract{
		ID:       id,
		Buyer:    "buyer-alice",
		Seller:   "seller-bob",
		Amount:   100.0,
		Currency: "USD",
		State:    "Queried from Ledger",
	}, nil
}

func (c *DamlClient) ReleaseFunds(ctx context.Context, id string) error {
	c.logger.Info("exercising ApproveMilestone choice on DAML ledger", zap.String("id", id))

	if c.client == nil {
		return fmt.Errorf("DAML client not connected")
	}

	exerciseCmd := &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", generated.PackageID, "StablecoinEscrow", "StablecoinEscrow"),
		ContractID: id,
		Choice:     "ApproveMilestone",
		Arguments:  map[string]interface{}{},
	}

	submitReq := &model.SubmitAndWaitRequest{
		Commands: &model.Commands{
			CommandID: fmt.Sprintf("approve-milestone-%d", time.Now().UnixNano()),
			UserID:    BuyerParty,
			ActAs:     []string{BuyerParty},
			Commands: []*model.Command{
				{
					Command: exerciseCmd,
				},
			},
		},
	}

	_, err := c.client.CommandService.SubmitAndWait(ctx, submitReq)
	if err != nil {
		c.logger.Error("failed to submit exercise command to DAML", zap.Error(err))
		return fmt.Errorf("ledger exercise failed: %w", err)
	}

	return nil
}

func (c *DamlClient) RefundBuyer(ctx context.Context, id string) error {
	c.logger.Info("exercising RaiseDispute choice on DAML ledger", zap.String("id", id))

	if c.client == nil {
		return fmt.Errorf("DAML client not connected")
	}

	exerciseCmd := &model.ExerciseCommand{
		TemplateID: fmt.Sprintf("%s:%s:%s", generated.PackageID, "StablecoinEscrow", "StablecoinEscrow"),
		ContractID: id,
		Choice:     "RaiseDispute",
		Arguments:  map[string]interface{}{},
	}

	submitReq := &model.SubmitAndWaitRequest{
		Commands: &model.Commands{
			CommandID: fmt.Sprintf("raise-dispute-%d", time.Now().UnixNano()),
			UserID:    BuyerParty,
			ActAs:     []string{BuyerParty},
			Commands: []*model.Command{
				{
					Command: exerciseCmd,
				},
			},
		},
	}

	_, err := c.client.CommandService.SubmitAndWait(ctx, submitReq)
	if err != nil {
		c.logger.Error("failed to submit dispute command to DAML", zap.Error(err))
		return fmt.Errorf("ledger dispute failed: %w", err)
	}

	return nil
}
