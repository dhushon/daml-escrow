package ledger

import (
	"context"
)

// Stablecoin defines the interface for interacting with stablecoin contracts on the ledger.
// This allows the escrow service to remain agnostic of the specific stablecoin implementation.
type Stablecoin interface {
	// Transfer initiates a transfer of funds from one party to another.
	Transfer(ctx context.Context, from, to string, amount float64, currency string) (string, error)

	// Lock reserves funds for an escrow contract.
	Lock(ctx context.Context, owner string, amount float64, currency string) (string, error)

	// Unlock releases previously locked funds.
	Unlock(ctx context.Context, lockID string, recipient string) (string, error)

	// GetBalance retrieves the stablecoin balance for a specific party.
	GetBalance(ctx context.Context, party string, currency string) (float64, error)
}
