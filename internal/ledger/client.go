package ledger

import (
	"context"
)

type CreateEscrowRequest struct {
	Buyer    string
	Seller   string
	Amount   float64
	Currency string
}

type EscrowContract struct {
	ID       string
	Buyer    string
	Seller   string
	Amount   float64
	Currency string
	State    string
}

type Client interface {
	CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error)
	GetEscrow(ctx context.Context, id string) (*EscrowContract, error)
	ReleaseFunds(ctx context.Context, id string) error
	RefundBuyer(ctx context.Context, id string) error
}
