package ledger

import (
	"context"
)

type Milestone struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Completed bool    `json:"completed"`
}

type CreateEscrowRequest struct {
	Buyer       string                 `json:"buyer"`
	Seller      string                 `json:"seller"`
	Amount      float64                `json:"amount"`
	Currency    string                 `json:"currency"`
	Description string                 `json:"description"`
	Milestones  []Milestone            `json:"milestones,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

type EscrowContract struct {
	ID                    string                 `json:"id"`
	Buyer                 string                 `json:"buyer"`
	Seller                string                 `json:"seller"`
	Issuer                string                 `json:"issuer"`
	Mediator              string                 `json:"mediator"`
	Amount                float64                `json:"amount"`
	Currency              string                 `json:"currency"`
	State                 string                 `json:"state"` // "Active" or "Disputed"
	Milestones            []Milestone            `json:"milestones"`
	CurrentMilestoneIndex int                    `json:"currentMilestoneIndex"`
	Metadata              map[string]interface{} `json:"metadata"`
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
	TotalActiveEscrows    int     `json:"totalActiveEscrows"`
	TotalValueInEscrow    float64 `json:"totalValueInEscrow"`
	PendingSettlements    int     `json:"pendingSettlements"`
	PendingSettlementValue float64 `json:"pendingSettlementValue"`
}

type OracleWebhookRequest struct {
	EscrowID       string                 `json:"escrowId"`
	MilestoneIndex int                    `json:"milestoneIndex"`
	Event          string                 `json:"event"`
	OracleProvider string                 `json:"oracleProvider"`
	Evidence       string                 `json:"evidence"`
	Metadata       map[string]interface{} `json:"metadata"`
	Signature      string                 `json:"signature"`
}

type Client interface {
	CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error)
	GetEscrow(ctx context.Context, id string) (*EscrowContract, error)
	
	// ListEscrows returns escrows visible to the given User ID
	ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error)
	
	ReleaseFunds(ctx context.Context, id string) error
	RaiseDispute(ctx context.Context, id string) (string, error)
	ResolveDispute(ctx context.Context, id string, payoutToBuyer, payoutToSeller float64) error
	RefundBuyer(ctx context.Context, id string) error
	RefundBySeller(ctx context.Context, id string) error
	
	// Aggregated Views
	GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error)
	
	// Settlement interactions
	ListSettlements(ctx context.Context) ([]*EscrowSettlement, error)
	SettlePayment(ctx context.Context, settlementID string) error

	// Internal helper for tests
	getParty(user string) string
	getOffset() interface{}
}
