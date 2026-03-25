package ledger

import (
	"context"
	"time"
)

// Directive 02
type Asset struct {
	AssetType  string  `json:"assetType"` // stablecoin | tokenized_reserve | digital_asset
	AssetID    string  `json:"assetId"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	CustodyRef string  `json:"custodyRef,omitempty"`
}

// Directive 04
type EscrowTerms struct {
	ConditionDescription string      `json:"conditionDescription"`
	ConditionType        string      `json:"conditionType"` // Binary | Partial | Milestone
	EvidenceRequired     string      `json:"evidenceRequired"`
	ExpiryDate           time.Time   `json:"expiryDate"`
	GracePeriodDays      int         `json:"gracePeriodDays"`
	DisputeWindowDays    int         `json:"disputeWindowDays"`
	PartialSchedule      []Milestone `json:"partialSchedule"`
	Metadata             string      `json:"metadata"`
}

type Milestone struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Completed bool    `json:"completed"`
}

// Directive 06
type SettlementTerms struct {
	SettlementType string  `json:"settlementType"` // FullRelease | FullReturn | PartialSplit
	BuyerReturn    float64 `json:"buyerReturn"`
	SellerPayment  float64 `json:"sellerPayment"`
	MediatorFee    float64 `json:"mediatorFee"`
}

type EscrowMetadata struct {
	SchemaURL  string                 `json:"schemaUrl"`
	Payload    map[string]interface{} `json:"payload"`
	Exclusions map[string]struct{}    `json:"-"`
}

type CreateEscrowRequest struct {
	Buyer       string         `json:"buyer"`
	Seller      string         `json:"seller"`
	Asset       Asset          `json:"asset"`
	Terms       EscrowTerms    `json:"terms"`
	Metadata    EscrowMetadata `json:"metadata"`
	Description string         `json:"description"` // Legacy support
	Amount      float64        `json:"amount"`      // Legacy support
	Currency    string         `json:"currency"`    // Legacy support
	Milestones  []Milestone    `json:"milestones"`  // Legacy support
}

type EscrowContract struct {
	ID                    string          `json:"id"`
	Buyer                 string          `json:"buyer"`
	Seller                string          `json:"seller"`
	Issuer                string          `json:"issuer"`
	Mediator              string          `json:"mediator"`
	Asset                 Asset           `json:"asset"`
	Terms                 EscrowTerms     `json:"terms"`
	State                 string          `json:"state"` // DRAFT, FUNDED, ACTIVE, DISPUTED, PROPOSED, SETTLED
	Metadata              EscrowMetadata  `json:"metadata"`
	Settlement            SettlementTerms `json:"settlement,omitempty"`
	BuyerAccepted         bool            `json:"buyerAccepted,omitempty"`
	SellerAccepted        bool            `json:"sellerAccepted,omitempty"`
	Amount                float64         `json:"amount"`                // Legacy
	Currency              string          `json:"currency"`              // Legacy
	CurrentMilestoneIndex int             `json:"currentMilestoneIndex"` // Legacy
}

type EscrowProposal struct {
	ID       string         `json:"id"`
	Buyer    string         `json:"buyer"`
	Seller   string         `json:"seller"`
	Issuer   string         `json:"issuer"`
	Mediator string         `json:"mediator"`
	Asset    Asset          `json:"asset"`
	Terms    EscrowTerms    `json:"terms"`
	Metadata EscrowMetadata `json:"metadata"`
	Amount   float64        `json:"amount"`      // Legacy
	Currency string         `json:"currency"`    // Legacy
}

type EscrowInvitation struct {
	ID           string      `json:"id"`
	Inviter      string      `json:"inviter"`
	Mediator     string      `json:"mediator"`
	Issuer       string      `json:"issuer"`
	InviteeEmail string      `json:"inviteeEmail"`
	InviteeRole  string      `json:"inviteeRole"`
	InviteeType  string      `json:"inviteeType"`
	TokenHash    string      `json:"tokenHash"`
	Asset        Asset       `json:"asset"`
	Terms        EscrowTerms `json:"terms"`
	Expiry       time.Time   `json:"expiry"`
}

type EscrowSettlement struct {
	ID        string  `json:"id"`
	Issuer    string  `json:"issuer"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
}

type LedgerMetrics struct {
	TotalActiveEscrows     int               `json:"totalActiveEscrows"`
	TotalValueInEscrow     float64           `json:"totalValueInEscrow"`
	PendingSettlements     int               `json:"pendingSettlements"`
	PendingSettlementValue float64           `json:"pendingSettlementValue"`
	ActivityHistory        []ActivityPoint   `json:"activityHistory"`
	TPSHistory             []ActivityPoint   `json:"tpsHistory"`
	SystemPerformance      SystemPerformance `json:"systemPerformance"`
	LedgerHealth           LedgerHealth      `json:"ledgerHealth"`
}

type ActivityPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type SystemPerformance struct {
	APILatencyMS      int     `json:"apiLatencyMs"`
	P95LatencyMS      int     `json:"p95LatencyMs"`
	P99LatencyMS      int     `json:"p99LatencyMs"`
	ErrorRate         float64 `json:"errorRate"`
	RequestCount      int     `json:"requestCount"`
	SuccessRate       float64 `json:"successRate"`
	Uptime            string  `json:"uptime"`
	CPUUsage          float64 `json:"cpuUsage"`
	MemoryUsage       float64 `json:"memoryUsage"`
	DiskUsage         float64 `json:"diskUsage"`
	ActiveConnections int     `json:"activeConnections"`
}

type LedgerHealth struct {
	TPS                float64 `json:"tps"`
	CommandSuccessRate float64 `json:"commandSuccessRate"`
	ActiveContracts    int     `json:"activeContracts"`
	ParticipantUptime  string  `json:"participantUptime"`
}

type Wallet struct {
	ID       string  `json:"id"`
	Owner    string  `json:"owner"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

type UserIdentity struct {
	OktaSub     string `json:"oktaSub"`
	DamlUserID  string `json:"damlUserId"`
	DamlPartyID string `json:"damlPartyId"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type OracleWebhookRequest struct {
	EscrowID       string `json:"escrowId"`
	MilestoneIndex int    `json:"milestoneIndex"`
	Event          string `json:"event"`
	OracleProvider string `json:"oracleProvider"`
	Signature      string `json:"signature"`
}

type HealthResponse struct {
	Status      string  `json:"status"`
	Version     string  `json:"version"`
	Uptime      string  `json:"uptime"`
	StartTime   string  `json:"startTime"`
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryUsage float64 `json:"memoryUsage"`
	Goroutines  int     `json:"goroutines"`
}

type Client interface {
	Discover(ctx context.Context) error

	// Lifecycle (Directive 05)
	ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error)
	Fund(ctx context.Context, id string, custodyRef string, userID string) error
	Activate(ctx context.Context, id string, userID string) error
	ConfirmConditions(ctx context.Context, id string, userID string) error
	RaiseDispute(ctx context.Context, id string, userID string) error
	ProposeSettlement(ctx context.Context, id string, proposal SettlementTerms, userID string) error
	RatifySettlement(ctx context.Context, id string, userID string) error
	FinalizeSettlement(ctx context.Context, id string, userID string) error
	Disburse(ctx context.Context, id string, userID string) error
	Cancel(ctx context.Context, id string, userID string) error
	ExpireEscrow(ctx context.Context, id string, userID string) error

	// Queries
	ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error)
	ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error)
	GetEscrow(ctx context.Context, id string, userID string) (*EscrowContract, error)

	// Invitation (Phase 5)
	CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, asset Asset, terms EscrowTerms) (*EscrowInvitation, error)
	ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*EscrowProposal, error)
	ListInvitations(ctx context.Context, userID string) ([]*EscrowInvitation, error)
	GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error)

	// System
	GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error)
	ListSettlements(ctx context.Context) ([]*EscrowSettlement, error)
	SettlePayment(ctx context.Context, settlementID string) error
	ListWallets(ctx context.Context, userID string) ([]*Wallet, error)
	GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error)
	ProvisionUser(ctx context.Context, oktaSub string, email string) (*UserIdentity, error)

	getParty(user string) string
	getOffset() interface{}

	// LEGACY
	CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error)
	ReleaseFunds(ctx context.Context, id string, userID string) error
	ResolveDispute(ctx context.Context, id string, b, s float64, userID string) error
	RefundBuyer(ctx context.Context, id string) error
	RefundBySeller(ctx context.Context, id string) error
}
