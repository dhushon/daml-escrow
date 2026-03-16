package ledger

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"
)

type DamlClient struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	escrows map[string]*EscrowContract
}

func NewDamlClient(logger *zap.Logger) *DamlClient {
	return &DamlClient{
		logger:  logger,
		escrows: make(map[string]*EscrowContract),
	}
}

func (c *DamlClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := fmt.Sprintf("escrow-%d", len(c.escrows)+1)
	escrow := &EscrowContract{
		ID:       id,
		Buyer:    req.Buyer,
		Seller:   req.Seller,
		Amount:   req.Amount,
		Currency: req.Currency,
		State:    "Created",
	}

	c.escrows[id] = escrow
	c.logger.Info("created escrow in memory", zap.String("id", id))
	return escrow, nil
}

func (c *DamlClient) GetEscrow(ctx context.Context, id string) (*EscrowContract, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	escrow, ok := c.escrows[id]
	if !ok {
		return nil, fmt.Errorf("escrow %s not found", id)
	}

	return escrow, nil
}

func (c *DamlClient) ReleaseFunds(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	escrow, ok := c.escrows[id]
	if !ok {
		return fmt.Errorf("escrow %s not found", id)
	}

	escrow.State = "Released"
	c.logger.Info("released funds in memory", zap.String("id", id))
	return nil
}

func (c *DamlClient) RefundBuyer(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	escrow, ok := c.escrows[id]
	if !ok {
		return fmt.Errorf("escrow %s not found", id)
	}

	escrow.State = "Refunded"
	c.logger.Info("refunded buyer in memory", zap.String("id", id))
	return nil
}
