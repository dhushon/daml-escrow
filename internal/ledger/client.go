package ledger

import (
	"context"
)

type Milestone struct {
	Label     string  `json:"label"`
	Amount    float64 `json:"amount"`
	Completed bool    `json:"completed"`
}

type EscrowMetadata struct {
	SchemaURL  string                 `json:"schemaUrl"`
	Payload    map[string]interface{} `json:"payload"`
	Exclusions map[string]struct{}    `json:"-"` // Fields to exclude from ledger (privacy)
}

type CreateEscrowRequest struct {
	Buyer       string                `json:"buyer"`
	Seller      string                `json:"seller"`
	Amount      float64               `json:"amount"`
	Currency    string                `json:"currency"`
	Description string                `json:"description"`
	Milestones  []Milestone           `json:"milestones"`
	Metadata    EscrowMetadata        `json:"metadata"`
}

type EscrowContract struct {
	ID                    string         `json:"id"`
	Buyer                 string         `json:"buyer"`
	Seller                string         `json:"seller"`
	Issuer                string         `json:"issuer"`
	Mediator              string         `json:"mediator"`
	Amount                float64        `json:"amount"`
	Currency              string         `json:"currency"`
	State                 string         `json:"state"`
	Milestones            []Milestone    `json:"milestones"`
	CurrentMilestoneIndex int            `json:"currentMilestoneIndex"`
	Metadata              EscrowMetadata `json:"metadata"`
}

type EscrowProposal struct {
	ID          string         `json:"id"`
	Buyer       string         `json:"buyer"`
	Seller      string         `json:"seller"`
	Issuer      string         `json:"issuer"`
	Mediator    string         `json:"mediator"`
	Amount      float64        `json:"amount"`
	Currency    string         `json:"currency"`
	Description string         `json:"description"`
	Milestones  []Milestone    `json:"milestones"`
	Metadata    EscrowMetadata `json:"metadata"`
}

type EscrowSettlement struct {
	ID        string  `json:"id"`
	Issuer    string  `json:"issuer"`
	Recipient string  `json:"recipient"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Status    string  `json:"status"`
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

type Wallet struct {
	ID       string  `json:"id"`
	Owner    string  `json:"owner"`
	Currency string  `json:"currency"`
	Balance  float64 `json:"balance"`
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
	// Escrow Lifecycle
	ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error)
	AcceptProposal(ctx context.Context, id string, sellerID string) error
	CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error)
	ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error)
	ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error)
	GetEscrow(ctx context.Context, id string, userID string) (*EscrowContract, error)
	ReleaseFunds(ctx context.Context, id string) error
	RaiseDispute(ctx context.Context, id string) (string, error)
	ResolveDispute(ctx context.Context, id string, payoutToBuyer, payoutToSeller float64) error
	RefundBuyer(ctx context.Context, id string) error
	RefundBySeller(ctx context.Context, id string) error

	// Metrics & Observability
	GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error)

	// Settlements
	ListSettlements(ctx context.Context) ([]*EscrowSettlement, error)
	SettlePayment(ctx context.Context, settlementID string) error

	// Wallet Management (Mockable)
	ListWallets(ctx context.Context, userID string) ([]*Wallet, error)

	// Internal helper for tests
	getParty(user string) string
	getOffset() interface{}
}
