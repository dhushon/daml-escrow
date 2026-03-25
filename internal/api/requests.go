package api

import (
	"errors"
	"strings"
	"time"

	"daml-escrow/internal/ledger"
)

// ProposeEscrowRequest is the API DTO for proposing a new escrow.
type ProposeEscrowRequest struct {
	Seller               string                 `json:"seller"`
	AssetType            string                 `json:"assetType"`
	AssetID              string                 `json:"assetId"`
	Amount               float64                `json:"amount"`
	Currency             string                 `json:"currency"`
	ConditionDescription string                 `json:"conditionDescription"`
	ConditionType        string                 `json:"conditionType"`
	EvidenceRequired     string                 `json:"evidenceRequired"`
	ExpiryDate           time.Time              `json:"expiryDate"`
	GracePeriodDays      int                    `json:"gracePeriodDays"`
	DisputeWindowDays    int                    `json:"disputeWindowDays"`
	SchemaURL            string                 `json:"schemaUrl"`
	Payload              map[string]interface{} `json:"payload"`
}

func (r *ProposeEscrowRequest) Validate() error {
	if strings.TrimSpace(r.Seller) == "" {
		return errors.New("seller is required")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if strings.TrimSpace(r.Currency) == "" {
		return errors.New("currency is required")
	}
	if r.ExpiryDate.Before(time.Now()) {
		return errors.New("expiry date must be in the future")
	}
	if r.GracePeriodDays < 0 {
		return errors.New("grace period days cannot be negative")
	}
	if r.DisputeWindowDays < 0 {
		return errors.New("dispute window days cannot be negative")
	}
	return nil
}

func (r *ProposeEscrowRequest) ToLedgerRequest() ledger.CreateEscrowRequest {
	return ledger.CreateEscrowRequest{
		Seller: r.Seller,
		Asset: ledger.Asset{
			AssetType: r.AssetType,
			AssetID:   r.AssetID,
			Amount:    r.Amount,
			Currency:  r.Currency,
		},
		Terms: ledger.EscrowTerms{
			ConditionDescription: r.ConditionDescription,
			ConditionType:        r.ConditionType,
			EvidenceRequired:     r.EvidenceRequired,
			ExpiryDate:           r.ExpiryDate,
			GracePeriodDays:      r.GracePeriodDays,
			DisputeWindowDays:    r.DisputeWindowDays,
			PartialSchedule:      []ledger.Milestone{},
		},
		Metadata: ledger.EscrowMetadata{
			SchemaURL: r.SchemaURL,
			Payload:   r.Payload,
		},
	}
}

// FundEscrowRequest is the API DTO for funding an escrow.
type FundEscrowRequest struct {
	CustodyRef string `json:"custodyRef"`
}

func (r *FundEscrowRequest) Validate() error {
	if strings.TrimSpace(r.CustodyRef) == "" {
		return errors.New("custodyRef is required")
	}
	return nil
}

// ProposeSettlementRequest is the API DTO for proposing a settlement during a dispute.
type ProposeSettlementRequest struct {
	SettlementType string  `json:"settlementType"`
	BuyerReturn    float64 `json:"buyerReturn"`
	SellerPayment  float64 `json:"sellerPayment"`
	MediatorFee    float64 `json:"mediatorFee"`
}

func (r *ProposeSettlementRequest) Validate() error {
	if strings.TrimSpace(r.SettlementType) == "" {
		return errors.New("settlementType is required")
	}
	if r.BuyerReturn < 0 || r.SellerPayment < 0 || r.MediatorFee < 0 {
		return errors.New("settlement amounts cannot be negative")
	}
	return nil
}

func (r *ProposeSettlementRequest) ToLedgerTerms() ledger.SettlementTerms {
	return ledger.SettlementTerms{
		SettlementType: r.SettlementType,
		BuyerReturn:    r.BuyerReturn,
		SellerPayment:  r.SellerPayment,
		MediatorFee:    r.MediatorFee,
	}
}

// CreateInvitationRequest is the API DTO for creating an invitation.
type CreateInvitationRequest struct {
	InviteeEmail         string    `json:"inviteeEmail"`
	InviteeRole          string    `json:"inviteeRole"`
	InviteeType          string    `json:"inviteeType"`
	AssetType            string    `json:"assetType"`
	AssetID              string    `json:"assetId"`
	Amount               float64   `json:"amount"`
	Currency             string    `json:"currency"`
	ConditionDescription string    `json:"conditionDescription"`
	ConditionType        string    `json:"conditionType"`
	EvidenceRequired     string    `json:"evidenceRequired"`
	ExpiryDate           time.Time `json:"expiryDate"`
	GracePeriodDays      int       `json:"gracePeriodDays"`
	DisputeWindowDays    int       `json:"disputeWindowDays"`
}

func (r *CreateInvitationRequest) Validate() error {
	if strings.TrimSpace(r.InviteeEmail) == "" {
		return errors.New("inviteeEmail is required")
	}
	if r.InviteeRole != "Buyer" && r.InviteeRole != "Seller" {
		return errors.New("inviteeRole must be 'Buyer' or 'Seller'")
	}
	if r.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if r.ExpiryDate.Before(time.Now()) {
		return errors.New("expiry date must be in the future")
	}
	return nil
}

func (r *CreateInvitationRequest) ToLedgerAssetAndTerms() (ledger.Asset, ledger.EscrowTerms) {
	asset := ledger.Asset{
		AssetType: r.AssetType,
		AssetID:   r.AssetID,
		Amount:    r.Amount,
		Currency:  r.Currency,
	}
	terms := ledger.EscrowTerms{
		ConditionDescription: r.ConditionDescription,
		ConditionType:        r.ConditionType,
		EvidenceRequired:     r.EvidenceRequired,
		ExpiryDate:           r.ExpiryDate,
		GracePeriodDays:      r.GracePeriodDays,
		DisputeWindowDays:    r.DisputeWindowDays,
		PartialSchedule:      []ledger.Milestone{},
	}
	return asset, terms
}

// OracleWebhookRequest is the API DTO for oracle webhooks.
type OracleWebhookRequest struct {
	EscrowID       string `json:"escrowId"`
	MilestoneIndex int    `json:"milestoneIndex"`
	Event          string `json:"event"`
	OracleProvider string `json:"oracleProvider"`
	Signature      string `json:"signature"`
}

func (r *OracleWebhookRequest) Validate() error {
	if strings.TrimSpace(r.EscrowID) == "" {
		return errors.New("escrowId is required")
	}
	if strings.TrimSpace(r.Event) == "" {
		return errors.New("event is required")
	}
	if strings.TrimSpace(r.Signature) == "" {
		return errors.New("signature is required")
	}
	return nil
}

func (r *OracleWebhookRequest) ToLedgerRequest() ledger.OracleWebhookRequest {
	return ledger.OracleWebhookRequest{
		EscrowID:       r.EscrowID,
		MilestoneIndex: r.MilestoneIndex,
		Event:          r.Event,
		OracleProvider: r.OracleProvider,
		Signature:      r.Signature,
	}
}
