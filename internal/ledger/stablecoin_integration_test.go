//go:build integration
// +build integration

package ledger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStablecoinProvider_MockIntegration(t *testing.T) {
	// MANDATORY: Test the MockStablecoinProvider during integration runs
	// to ensure interface stability and service-layer compatibility.
	ctx := context.Background()
	m := &MockStablecoinProvider{}

	t.Run("EnsureVault", func(t *testing.T) {
		m.On("EnsureVault", ctx, "user-123").Return("vault-abc", nil).Once()
		
		vaultID, err := m.EnsureVault(ctx, "user-123")
		require.NoError(t, err)
		require.Equal(t, "vault-abc", vaultID)
		m.AssertExpectations(t)
	})

	t.Run("GetBalance", func(t *testing.T) {
		m.On("GetBalance", ctx, "vault-abc", "USD").Return(1000.0, nil).Once()
		
		balance, err := m.GetBalance(ctx, "vault-abc", "USD")
		require.NoError(t, err)
		require.Equal(t, 1000.0, balance)
		m.AssertExpectations(t)
	})

	t.Run("Transfer", func(t *testing.T) {
		m.On("Transfer", ctx, "v1", "v2", 500.0, "USD", "ref-1").Return("tx-999", nil).Once()
		
		txID, err := m.Transfer(ctx, "v1", "v2", 500.0, "USD", "ref-1")
		require.NoError(t, err)
		require.Equal(t, "tx-999", txID)
		m.AssertExpectations(t)
	})
}

func TestStablecoinProvider_RealIntegration(t *testing.T) {
	// This test only runs if specific provider tags are present (future state)
	// For now, we test the logic of JsonStablecoinProvider against the sandbox if it's up.
	
	// TODO: Implement when CIP-0056 templates are fully deployed in sandbox_init.canton
	t.Skip("Skipping real integration until CIP-0056 templates are deployed")
}
