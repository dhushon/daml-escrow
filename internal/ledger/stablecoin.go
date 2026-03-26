package ledger

import (
	"context"
	"fmt"
)

// StablecoinProvider defines the interface for interacting with tokenized reserves.
type StablecoinProvider interface {
	// CreateWallet associates a user with a new digital wallet.
	CreateWallet(ctx context.Context, userID string) (string, error)

	// GetBalance retrieves the current balance for a wallet.
	GetBalance(ctx context.Context, walletID string, currency string) (float64, error)

	// Transfer initiates a move of funds between wallets.
	Transfer(ctx context.Context, fromID, toID string, amount float64, currency string, reference string) (string, error)

	// VerifyTransfer checks the status of a specific transaction.
	VerifyTransfer(ctx context.Context, transferID string) (bool, error)
}

// MockStablecoinProvider is used for Phase 6 development and testing.
type MockStablecoinProvider struct {
	wallets map[string]float64
}

func NewMockStablecoinProvider() *MockStablecoinProvider {
	return &MockStablecoinProvider{
		wallets: make(map[string]float64),
	}
}

func (m *MockStablecoinProvider) CreateWallet(ctx context.Context, userID string) (string, error) {
	walletID := fmt.Sprintf("mock-wallet-%s", userID)
	m.wallets[walletID] = 0.0
	return walletID, nil
}

func (m *MockStablecoinProvider) GetBalance(ctx context.Context, walletID string, currency string) (float64, error) {
	balance, ok := m.wallets[walletID]
	if !ok {
		return 0.0, nil // Default for mock
	}
	return balance, nil
}

func (m *MockStablecoinProvider) Transfer(ctx context.Context, fromID, toID string, amount float64, currency string, reference string) (string, error) {
	// Simulate success for mock
	return fmt.Sprintf("mock-tx-%s", reference), nil
}

func (m *MockStablecoinProvider) VerifyTransfer(ctx context.Context, transferID string) (bool, error) {
	return true, nil
}
