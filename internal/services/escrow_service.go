package services

import (
	"context"
	"daml-escrow/internal/crypto"
	"daml-escrow/internal/ledger"
	"fmt"
	"time"

	"go.uber.org/zap"
)

type EscrowService struct {
	logger        *zap.Logger
	ledger        ledger.Client
	stablecoin    ledger.StablecoinProvider
	compliance    ComplianceService
	webhookSecret string
	oracleSigner  crypto.HighAssuranceSigner
}

func NewEscrowService(
	logger *zap.Logger,
	ledger ledger.Client,
	stablecoin ledger.StablecoinProvider,
	compliance ComplianceService,
	webhookSecret string,
	oracleSigner crypto.HighAssuranceSigner,
) *EscrowService {
	return &EscrowService{
		logger:        logger,
		ledger:        ledger,
		stablecoin:    stablecoin,
		compliance:    compliance,
		webhookSecret: webhookSecret,
		oracleSigner:  oracleSigner,
	}
}

// ---------------------------------------------------------------------------
// Lifecycle Operations
// ---------------------------------------------------------------------------

func (s *EscrowService) ProposeEscrow(ctx context.Context, req ledger.CreateEscrowRequest) (*ledger.EscrowProposal, error) {
	s.logger.Info("proposing new escrow", zap.String("depositor", req.Depositor), zap.String("beneficiary", req.Beneficiary))
	return s.ledger.ProposeEscrow(ctx, req)
}

func (s *EscrowService) FundEscrow(ctx context.Context, id string, userID string, holdingID string) error {
	s.logger.Info("funding escrow", zap.String("id", id), zap.String("userID", userID))
	// Fund(ctx, id, custodyRef, holdingCid, userID)
	return s.ledger.Fund(ctx, id, "CUSTODY-REF-001", holdingID, userID)
}

func (s *EscrowService) ActivateEscrow(ctx context.Context, id string, userID string, actAs []string) (string, error) {
	s.logger.Info("activating escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.Activate(ctx, id, actAs)
}

func (s *EscrowService) DisburseEscrow(ctx context.Context, id string, userID string, actAs []string) error {
	s.logger.Info("disbursing escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.Disburse(ctx, id, actAs)
}

func (s *EscrowService) RaiseDispute(ctx context.Context, id string, userID string) error {
	s.logger.Info("raising dispute", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.RaiseDispute(ctx, id, userID)
}

func (s *EscrowService) ProposeSettlement(ctx context.Context, id string, userID string, amount float64) (string, error) {
	s.logger.Info("proposing settlement", zap.String("id", id), zap.Float64("amount", amount))
	proposal := ledger.SettlementTerms{
		SettlementType:     "PARTIAL",
		DepositorReturn:    amount,
		BeneficiaryPayment: 0.0, // Simplified for this call
		MediatorFee:        0.0,
	}
	return s.ledger.ProposeSettlement(ctx, id, proposal, userID)
}

func (s *EscrowService) CancelEscrow(ctx context.Context, id string, userID string) error {
	s.logger.Info("cancelling escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.Cancel(ctx, id, userID)
}

func (s *EscrowService) ExpireEscrow(ctx context.Context, id string, userID string) (string, error) {
	s.logger.Info("expiring escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.ExpireEscrow(ctx, id, userID)
}

// ---------------------------------------------------------------------------
// Utility and Query methods
// ---------------------------------------------------------------------------

func (s *EscrowService) ListEscrows(ctx context.Context, userID string) ([]*ledger.EscrowContract, error) {
	return s.ledger.ListEscrows(ctx, userID)
}

func (s *EscrowService) GetEscrow(ctx context.Context, id string, userID string) (*ledger.EscrowContract, error) {
	return s.ledger.GetEscrow(ctx, id, userID)
}

func (s *EscrowService) GetLedgerClient() ledger.Client {
	return s.ledger
}

func (s *EscrowService) GetOracleSecret() string {
	return s.webhookSecret
}

// Authoritative Settlement Triggers

func (s *EscrowService) SettleEscrow(ctx context.Context, settlementID string) error {
	s.logger.Info("finalizing settlement", zap.String("id", settlementID))
	return s.ledger.SettlePayment(ctx, settlementID)
}

func (s *EscrowService) OracleMilestoneTrigger(ctx context.Context, escrowID string, milestoneIndex int, event string, signature string, asymmetric bool) error {
	// High-Assurance: Verify signature against authoritative oracle signer
	var signer crypto.HighAssuranceSigner
	if asymmetric {
		signer = s.oracleSigner
	}
	
	if !s.compliance.VerifyOracleSignature(ctx, escrowID, milestoneIndex, event, signature, signer) {
		return fmt.Errorf("invalid oracle signature")
	}

	// Fetch current state
	escrow, err := s.ledger.GetEscrow(ctx, escrowID, "CentralBank") // Internal lookup
	if err != nil {
		return fmt.Errorf("escrow not found: %w", err)
	}

	if escrow.CurrentMilestoneIndex != milestoneIndex {
		return fmt.Errorf("milestone index mismatch: expected %d, got %d", escrow.CurrentMilestoneIndex, milestoneIndex)
	}

	// Execute authoritative transition (Internal admin action)
	_, err = s.ledger.Activate(ctx, escrowID, []string{"CentralBank"}) 
	return err
}

func (s *EscrowService) GetMetrics(ctx context.Context, userID string) (*ledger.LedgerMetrics, error) {
	return s.ledger.GetMetrics(ctx, userID)
}

func (s *EscrowService) PromoteDraft(ctx context.Context, draft *DraftEscrow, userID string) (string, error) {
	s.logger.Info("promoting draft to ledger", zap.String("rootId", draft.RootID), zap.String("userID", userID))

	// 1. Authoritatively determine if beneficiary is registered
	if draft.BeneficiaryID == "" {
		// No beneficiary ID yet, we must create an Invitation
		inv, err := s.ledger.CreateInvitation(ctx, draft.InitiatorID, draft.CounterpartyEmail, "Beneficiary", "Business", ledger.Asset{
			Amount:   draft.Amount,
			Currency: draft.Currency,
		}, ledger.EscrowTerms{
			ExpiryDate: time.Now().AddDate(0, 0, 30), // Default
		})
		if err != nil {
			return "", fmt.Errorf("failed to create escrow invitation: %w", err)
		}
		return inv.ID, nil
	}

	// 2. Beneficiary is registered, create a Proposal
	req := ledger.CreateEscrowRequest{
		Depositor:   draft.InitiatorID,
		Beneficiary: draft.BeneficiaryID,
		Mediator:    "EscrowMediator", // Default for now
		Asset: ledger.Asset{
			Amount:   draft.Amount,
			Currency: draft.Currency,
		},
		Terms: ledger.EscrowTerms{
			ExpiryDate: time.Now().AddDate(0, 0, 30),
		},
		Metadata: string(draft.Metadata),
	}

	proposal, err := s.ledger.ProposeEscrow(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to propose escrow: %w", err)
	}

	return proposal.ID, nil
}
