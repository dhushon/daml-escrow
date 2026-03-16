package ledger

import (
	"context"
)

type CreateEscrowRequest struct {
	Buyer    string  `json:"buyer"`
	Seller   string  `json:"seller"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type EscrowContract struct {
	ID       string  `json:"id"`
	Buyer    string  `json:"buyer"`
	Seller   string  `json:"seller"`
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
	State    string  `json:"state"`
}

type Client interface {
	CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error)
	GetEscrow(ctx context.Context, id string) (*EscrowContract, error)
	ReleaseFunds(ctx context.Context, id string) error
	RefundBuyer(ctx context.Context, id string) error
}
