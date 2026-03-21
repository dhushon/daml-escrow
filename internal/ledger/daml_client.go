package ledger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// DamlClient implements the Client interface using gRPC.
// (Simplified for this prototype, mostly stubbed)
type DamlClient struct {
	logger *zap.Logger
}

func NewDamlClient(logger *zap.Logger, host string, port int) *DamlClient {
	return &DamlClient{
		logger: logger,
	}
}

func (c *DamlClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	return nil, fmt.Errorf("gRPC not implemented in this prototype")
}

func (c *DamlClient) ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error) {
	return nil, fmt.Errorf("gRPC not implemented in this prototype")
}

func (c *DamlClient) AcceptProposal(ctx context.Context, id string, sellerID string) error {
	return fmt.Errorf("gRPC not implemented in this prototype")
}

func (c *DamlClient) ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error) {
	return nil, nil
}

func (c *DamlClient) ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error) {
	return nil, nil
}

func (c *DamlClient) GetEscrow(ctx context.Context, id string) (*EscrowContract, error) {
	return nil, fmt.Errorf("not found")
}

func (c *DamlClient) ReleaseFunds(ctx context.Context, id string) error {
	return nil
}

func (c *DamlClient) RaiseDispute(ctx context.Context, id string) (string, error) {
	return "", nil
}

func (c *DamlClient) ResolveDispute(ctx context.Context, id string, payoutToBuyer, payoutToSeller float64) error {
	return nil
}

func (c *DamlClient) RefundBuyer(ctx context.Context, id string) error {
	return nil
}

func (c *DamlClient) RefundBySeller(ctx context.Context, id string) error {
	return nil
}

func (c *DamlClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	return &LedgerMetrics{}, nil
}

func (c *DamlClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	return nil, nil
}

func (c *DamlClient) SettlePayment(ctx context.Context, settlementID string) error {
	return nil
}

func (c *DamlClient) ListWallets(ctx context.Context, userID string) ([]*Wallet, error) {
	return nil, nil
}

func (c *DamlClient) getParty(user string) string {
	return user
}

func (c *DamlClient) getOffset() interface{} {
	return nil
}

var _ Client = (*DamlClient)(nil)
