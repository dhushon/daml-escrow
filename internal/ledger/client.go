package ledger

import (
	"context"
	"time"
)

// User IDs mapped in init.canton
const (
	CentralBankUser    = "CentralBank"
	DepositorUser      = "Depositor"
	BeneficiaryUser    = "Beneficiary"
	EscrowMediatorUser = "EscrowMediator"
)

type Asset struct {
	AssetType  string  `json:"assetType"`
	AssetID    string  `json:"assetId"`
	Amount     float64 `json:"amount,string"`
	Currency   string  `json:"currency"`
	CustodyRef string  `json:"custodyRef,omitempty"`
	HoldingCid string  `json:"holdingCid,omitempty"`
}

type EscrowTerms struct {
	ConditionDescription string    `json:"conditionDescription"`
	ConditionType        string    `json:"conditionType"`
	EvidenceRequired     string    `json:"evidenceRequired"`
	ExpiryDate           time.Time `json:"expiryDate"`
	GracePeriodDays      int       `json:"gracePeriodDays,string"`
	DisputeWindowDays    int       `json:"disputeWindowDays,string"`
	PartialSchedule      []string  `json:"partialSchedule"` // Simplified for JSON
	MinAmount            float64   `json:"minAmount,string,omitempty"`
	MaxAmount            float64   `json:"maxAmount,string,omitempty"`
}

type EscrowMetadata struct {
	SchemaURL    string                 `json:"schemaUrl"`
	AgreementURI string                 `json:"agreementUri,omitempty"` // storage://bank-vault/ingest/abc.pdf
	ContentHash  string                 `json:"contentHash,omitempty"`  // SHA-256
	Payload      map[string]interface{} `json:"payload"`
}

type SettlementTerms struct {
	SettlementType     string  `json:"settlementType"`
	DepositorReturn    float64 `json:"depositorReturn,string"`
	BeneficiaryPayment float64 `json:"beneficiaryPayment,string"`
	MediatorFee        float64 `json:"mediatorFee,string"`
}

type CreateEscrowRequest struct {
	Depositor            string      `json:"depositor"` // Kept for backward compatibility
	Beneficiary          string      `json:"beneficiary"` // Kept for backward compatibility
	Depositors           []string    `json:"depositors"`
	DepositorThreshold   int         `json:"depositorThreshold"`
	Beneficiaries        []string    `json:"beneficiaries"`
	BeneficiaryThreshold int         `json:"beneficiaryThreshold"`
	Mediator             string      `json:"mediator"`
	ContractType         string      `json:"contractType"`
	Asset                Asset       `json:"asset"`
	Terms                EscrowTerms `json:"terms"`
	Metadata             string      `json:"metadata"`
}

type EscrowProposal struct {
	ID                     string      `json:"id"`
	Issuer                 string      `json:"issuer"`
	Depositor              string      `json:"depositor"` // Kept for backward compatibility
	Beneficiary            string      `json:"beneficiary"` // Kept for backward compatibility
	Depositors             []string    `json:"depositors"`
	DepositorThreshold     int         `json:"depositorThreshold"`
	Beneficiaries          []string    `json:"beneficiaries"`
	BeneficiaryThreshold   int         `json:"beneficiaryThreshold"`
	Mediator               string      `json:"mediator"`
	Asset                  Asset       `json:"asset"`
	Terms                  EscrowTerms `json:"terms"`
	Metadata               string      `json:"metadata"`
	DepositorAcceptances   []string    `json:"depositorAcceptances"`
	BeneficiaryAcceptances []string    `json:"beneficiaryAcceptances"`
}

type EscrowContract struct {
	ID                     string      `json:"id"`
	Issuer                 string      `json:"issuer"`
	Depositor              string      `json:"depositor"` // Kept for backward compatibility
	Beneficiary            string      `json:"beneficiary"` // Kept for backward compatibility
	Depositors             []string    `json:"depositors"`
	DepositorThreshold     int         `json:"depositorThreshold"`
	Beneficiaries          []string    `json:"beneficiaries"`
	BeneficiaryThreshold   int         `json:"beneficiaryThreshold"`
	Mediator               string      `json:"mediator"`
	Asset                  Asset       `json:"asset"`
	Terms                  EscrowTerms `json:"terms"`
	Metadata               string      `json:"metadata"`
	AgreementURL           string      `json:"agreementUrl,omitempty"`
	State                  string      `json:"state"`
	CurrentMilestoneIndex  int         `json:"currentMilestoneIndex"`
	DepositorAccepted      bool        `json:"depositorAccepted"`
	BeneficiaryAccepted    bool        `json:"beneficiaryAccepted"`
	DepositorAcceptances   []string    `json:"depositorAcceptances"`
	BeneficiaryAcceptances []string    `json:"beneficiaryAcceptances"`
}

type EscrowInvitation struct {
	ID           string      `json:"id"`
	Inviter      string      `json:"inviter"`
	Mediator     string      `json:"mediator"`
	Issuer       string      `json:"issuer"`
	InviteeEmail string      `json:"inviteeEmail"`
	InviteeRole  string      `json:"inviteeRole"` // Depositor or Beneficiary
	ContractType string      `json:"contractType"`
	TokenHash    string      `json:"tokenHash"`
	Asset        Asset       `json:"asset"`
	Terms        EscrowTerms `json:"terms"`
}

type EscrowSettlement struct {
	ID         string          `json:"id"`
	EscrowID   string          `json:"escrowId"`
	Issuer     string          `json:"issuer"`
	Recipient  string          `json:"recipient"`
	Amount     float64         `json:"amount"`
	Currency   string          `json:"currency"`
	Settlement SettlementTerms `json:"settlement"`
	State      string          `json:"state"`
	Status     string          `json:"status"`
}

type LedgerMetrics struct {
	TotalEscrows       int               `json:"totalEscrows"`
	TotalActiveEscrows int               `json:"totalActiveEscrows"`
	TotalValueInEscrow float64           `json:"totalValueInEscrow"`
	ActiveEscrows      int               `json:"activeEscrows"`
	DisputedEscrows    int               `json:"disputedEscrows"`
	SettledVolume      float64           `json:"settledVolume"`
	ActivityHistory    []ActivityPoint   `json:"activityHistory"`
	LedgerHealth       LedgerHealth      `json:"ledgerHealth"`
	TPSHistory         []ActivityPoint   `json:"tpsHistory"`
	SystemPerformance  SystemPerformance `json:"systemPerformance"`
	AvgTimeToSettle    string            `json:"avgTimeToSettle"`
	BottleneckStage    string            `json:"bottleneckStage"`
	StageLatencies     map[string]int    `json:"stageLatencies"` // ms
	SuccessRate        float64           `json:"successRate"`    // percentage
}

type ActivityPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type LedgerHealth struct {
	TPS                 float64 `json:"tps"`
	CommandSuccessRate  float64 `json:"commandSuccessRate"`
	ActiveContracts     int     `json:"activeContracts"`
	IdentitiesAllocated int     `json:"identitiesAllocated"`
}

type SystemPerformance struct {
	ApiLatencyMs int     `json:"apiLatencyMs"`
	P95LatencyMs int     `json:"p95LatencyMs"`
	P99LatencyMs int     `json:"p99LatencyMs"`
	CpuUsage     float64 `json:"cpuUsage"`
	MemoryUsage  float64 `json:"memoryUsage"`
	Uptime       string  `json:"uptime"`
}

type Wallet struct {
	ID       string  `json:"id"`
	Owner    string  `json:"owner"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
}

type UserIdentity struct {
	OktaSub         string `json:"oktaSub"`
	DamlUserID      string `json:"damlUserId"`
	DamlPartyID     string `json:"damlPartyId"`
	Email           string `json:"email"`
	DisplayName     string `json:"displayName"`
	Role            string `json:"role"` // Depositor, Beneficiary, Mediator
	Title           string `json:"title,omitempty"`
	Affiliation     string `json:"affiliation,omitempty"`
	Organization    string `json:"organization,omitempty"`
	PhysicalAddress string `json:"physicalAddress,omitempty"`
	KYCStatus       string `json:"kycStatus,omitempty"` // PENDING, VERIFIED, REJECTED
}

type OracleWebhookRequest struct {
	EscrowID       string `json:"escrowId"`
	MilestoneIndex int    `json:"milestoneIndex"`
	Event          string `json:"event"`
	OracleProvider string `json:"oracleProvider"`
	Signature      string `json:"signature"`
}

type HealthResponse struct {
	Status      string                   `json:"status"`
	Version     string                   `json:"version"`
	Uptime      string                   `json:"uptime"`
	StartTime   string                   `json:"startTime"`
	CPUUsage    float64                  `json:"cpuUsage"`
	MemoryUsage float64                  `json:"memoryUsage"`
	Goroutines  int                      `json:"goroutines"`
	Services    map[string]ServiceHealth `json:"services"`
}

type ServiceHealth struct {
	Status      string `json:"status"` // UP, DOWN, DEGRADED
	Message     string `json:"message,omitempty"`
	LatencyMs   int64  `json:"latencyMs,omitempty"`
}

type Client interface {
	Discover(ctx context.Context, wait bool) error

	// Lifecycle (Directive 05)
	ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error)
	BeneficiaryAccept(ctx context.Context, id string, userID string) (string, error)
	Fund(ctx context.Context, id string, custodyRef string, holdingCid string, userID string) error
	Activate(ctx context.Context, id string, actAs []string) (string, error)
	ConfirmConditions(ctx context.Context, id string, userID string) error
	RaiseDispute(ctx context.Context, id string, userID string) error
	ProposeSettlement(ctx context.Context, id string, proposal SettlementTerms, userID string) (string, error)
	RatifySettlement(ctx context.Context, id string, userID string) (string, error)
	FinalizeSettlement(ctx context.Context, id string, userID string) (string, error)
	Disburse(ctx context.Context, id string, actAs []string) error
	Cancel(ctx context.Context, id string, userID string) error
	WithdrawProposal(ctx context.Context, id string, userID string) error
	ExpireEscrow(ctx context.Context, id string, userID string) (string, error)

	// Queries
	ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error)
	ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error)
	GetEscrow(ctx context.Context, id string, userID string) (*EscrowContract, error)
	GetPartyMap() map[string]string
	SetPartyMap(m map[string]string)

	// Invitation (Phase 5)
	CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, contractType string, asset Asset, terms EscrowTerms) (*EscrowInvitation, error)
	ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*EscrowProposal, error)
	ListInvitations(ctx context.Context, userID string) ([]*EscrowInvitation, error)
	GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error)

	// System
	GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error)
	ListSettlements(ctx context.Context) ([]*EscrowSettlement, error)
	SettlePayment(ctx context.Context, settlementID string) error
	ListWallets(ctx context.Context, userID string) ([]*Wallet, error)
	GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error)
	ProvisionUser(ctx context.Context, oktaSub string, email string, role string, scopes []string) (*UserIdentity, error)
	ListIdentities(ctx context.Context) ([]*UserIdentity, error)

	// Utilities
	CreateContract(ctx context.Context, userID string, templateID string, payload map[string]interface{}) (string, error)
	SearchPackageID(ctx context.Context, name string) (string, error)
	GetParty(user string) string
	GetOffset() interface{}
	GetInterfacePackageID() string
	DoRawRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error)

	// LEGACY
	CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error)
	ReleaseFunds(ctx context.Context, id string, userID string) error
	ResolveDispute(ctx context.Context, id string, b, s float64, userID string) error
	RefundDepositor(ctx context.Context, id string) error
	RefundByBeneficiary(ctx context.Context, id string) error
}

// StablecoinProvider defines the interface for interacting with tokenized reserves.
type StablecoinProvider interface {
	// EnsureVault ensures a user has a vault/party allocated for holdings.
	EnsureVault(ctx context.Context, userID string) (string, error)

	// GetBalance retrieves the current balance for a vault.
	GetBalance(ctx context.Context, vaultID string, currency string) (float64, error)

	// Transfer initiates an authoritative move of funds.
	Transfer(ctx context.Context, fromID, toID string, amount float64, currency string, reference string) (string, error)

	// VerifyTransfer checks the status of a specific transaction.
	VerifyTransfer(ctx context.Context, transferID string) (bool, error)
}
