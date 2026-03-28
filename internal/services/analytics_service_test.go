package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAnalyticsService(t *testing.T) {
	logger := zap.NewNop()
	svc := NewAnalyticsService(logger)
	ctx := context.Background()

	t.Run("ConfirmDeposit - Success", func(t *testing.T) {
		ok, err := svc.ConfirmDeposit(ctx, "0xhash", 1000.0, "USD")
		assert.NoError(t, err)
		assert.True(t, ok)
	})

	t.Run("GetEscrowLifecycle - Transitions", func(t *testing.T) {
		// Test logic for DRAFT state
		meta, err := svc.GetEscrowLifecycle(ctx, "escrow-1", "DRAFT")
		assert.NoError(t, err)
		assert.Equal(t, "CURRENT", meta.Steps[0].Status)
		assert.Equal(t, "PENDING", meta.Steps[1].Status)

		// Test logic for ACTIVE state
		meta2, err := svc.GetEscrowLifecycle(ctx, "escrow-2", "ACTIVE")
		assert.NoError(t, err)
		assert.Equal(t, "COMPLETED", meta2.Steps[0].Status)
		assert.Equal(t, "COMPLETED", meta2.Steps[1].Status)
		assert.Equal(t, "CURRENT", meta2.Steps[2].Status)
		assert.Equal(t, "PENDING", meta2.Steps[3].Status)
	})

	t.Run("Wallet History", func(t *testing.T) {
		history, err := svc.GetWalletHistory(ctx, "0xaddress")
		assert.NoError(t, err)
		assert.NotEmpty(t, history)
		assert.Equal(t, "DEPOSIT", history[0]["type"])
	})
}
