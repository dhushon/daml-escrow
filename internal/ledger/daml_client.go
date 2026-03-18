package ledger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type DamlClient struct {
	logger *zap.Logger
}

func NewDamlClient(logger *zap.Logger, host string, port int) *DamlClient {
	return &DamlClient{
		logger: logger,
	}
}

func (c *DamlClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	return nil, fmt.Errorf("gRPC CreateEscrow not implemented")
}

func (c *DamlClient) GetEscrow(ctx context.Context, id string) (*EscrowContract, error) {
	return nil, fmt.Errorf("gRPC GetEscrow not implemented")
}

func (c *DamlClient) ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error) {
	return nil, fmt.Errorf("gRPC ListEscrows not implemented")
}

func (c *DamlClient) ReleaseFunds(ctx context.Context, id string) error {
	return fmt.Errorf("gRPC ReleaseFunds not implemented")
}

func (c *DamlClient) RaiseDispute(ctx context.Context, id string) (string, error) {
	return "", fmt.Errorf("gRPC RaiseDispute not implemented")
}

func (c *DamlClient) ResolveDispute(ctx context.Context, id string, payoutToBuyer, payoutToSeller float64) error {
	return fmt.Errorf("gRPC ResolveDispute not implemented")
}

func (c *DamlClient) RefundBuyer(ctx context.Context, id string) error {
	return fmt.Errorf("gRPC RefundBuyer not implemented")
}

func (c *DamlClient) RefundBySeller(ctx context.Context, id string) error {
	return fmt.Errorf("gRPC RefundBySeller not implemented")
}

func (c *DamlClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	return nil, fmt.Errorf("gRPC ListSettlements not implemented")
}

func (c *DamlClient) SettlePayment(ctx context.Context, settlementID string) error {
	return fmt.Errorf("gRPC SettlePayment not implemented")
}

func (c *DamlClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	return nil, fmt.Errorf("gRPC GetMetrics not implemented")
}

func (c *DamlClient) getParty(user string) string {
	return user
}

func (c *DamlClient) getOffset() interface{} {
	return nil
}

// Ensure DamlClient implements Client interface
var _ Client = (*DamlClient)(nil)
