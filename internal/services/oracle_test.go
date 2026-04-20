package services

import (
	"context"
	"testing"

	"daml-escrow/internal/ledger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestOracleMilestoneTrigger(t *testing.T) {
	secret := "test-secret"
	logger, _ := zap.NewDevelopment()
	compliance := NewMockCompliance()
	
	t.Run("Valid Signature", func(t *testing.T) {
		mockLedger := new(ledger.MockLedgerClient)
		stablecoin := ledger.NewJsonStablecoinProvider(logger, mockLedger)
		svc := NewEscrowService(logger, mockLedger, stablecoin, compliance, secret)
		
		escrowID := "escrow-123"
		milestoneIndex := 0
		event := "DELIVERED"
		
		// Generate valid signature (must match Oracle implementation's logic)
		// For tests, we use the simple mock verification logic
		signature := "valid-mock-sig"
		
		// Setup mock expectations - Oracle uses CentralBank for lookups in this refactor
		mockLedger.On("GetEscrow", mock.Anything, escrowID, "CentralBank").Return(&ledger.EscrowContract{
			ID:                    escrowID,
			CurrentMilestoneIndex: 0,
			State:                 "ACTIVE",
		}, nil)
		mockLedger.On("Activate", mock.Anything, escrowID, []string{"CentralBank"}).Return("new-id", nil)
		
		err := svc.OracleMilestoneTrigger(context.Background(), escrowID, milestoneIndex, event, signature)
		assert.NoError(t, err)
		mockLedger.AssertExpectations(t)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		mockLedger := new(ledger.MockLedgerClient)
		stablecoin := ledger.NewJsonStablecoinProvider(logger, mockLedger)
		svc := NewEscrowService(logger, mockLedger, stablecoin, compliance, secret)
		
		err := svc.OracleMilestoneTrigger(context.Background(), "escrow-123", 0, "EVENT", "invalid-sig")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid oracle signature")
	})

	t.Run("Mismatched Milestone Index", func(t *testing.T) {
		mockLedger := new(ledger.MockLedgerClient)
		stablecoin := ledger.NewJsonStablecoinProvider(logger, mockLedger)
		svc := NewEscrowService(logger, mockLedger, stablecoin, compliance, secret)
		
		escrowID := "escrow-123"
		
		// Contract is still at index 0
		mockLedger.On("GetEscrow", mock.Anything, escrowID, "CentralBank").Return(&ledger.EscrowContract{
			ID:                    escrowID,
			CurrentMilestoneIndex: 0,
			State:                 "ACTIVE",
		}, nil)
		
		err := svc.OracleMilestoneTrigger(context.Background(), escrowID, 1, "DELIVERED", "valid-sig")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "milestone index mismatch")
	})
}
