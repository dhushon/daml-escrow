package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"

	"daml-escrow/internal/ledger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestProcessOracleWebhook(t *testing.T) {
	secret := "test-secret"
	logger, _ := zap.NewDevelopment()
	compliance := NewMockCompliance()
	
	t.Run("Valid Signature", func(t *testing.T) {
		mockLedger := new(ledger.MockLedgerClient)
		stablecoin := ledger.NewJsonStablecoinProvider(logger, mockLedger)
		svc := NewEscrowService(logger, mockLedger, stablecoin, compliance, secret)
		
		req := ledger.OracleWebhookRequest{
			EscrowID:       "escrow-123",
			MilestoneIndex: 0,
			Event:          "DELIVERED",
			OracleProvider: "FedEx",
		}
		
		// Generate valid signature
		payload := fmt.Sprintf("%s|%d|%s|%s", req.EscrowID, req.MilestoneIndex, req.Event, req.OracleProvider)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(payload))
		req.Signature = hex.EncodeToString(h.Sum(nil))
		
		// Setup mock expectations - Oracle uses EscrowMediatorUser for lookups
		mockLedger.On("GetEscrow", mock.Anything, "escrow-123", ledger.EscrowMediatorUser).Return(&ledger.EscrowContract{
			ID:                    "escrow-123",
			CurrentMilestoneIndex: 0,
			State:                 "ACTIVE",
		}, nil)
		mockLedger.On("ConfirmConditions", mock.Anything, "escrow-123", ledger.EscrowMediatorUser).Return(nil)
		
		err := svc.ProcessOracleWebhook(context.Background(), req)
		assert.NoError(t, err)
		mockLedger.AssertExpectations(t)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		mockLedger := new(ledger.MockLedgerClient)
		stablecoin := ledger.NewJsonStablecoinProvider(logger, mockLedger)
		svc := NewEscrowService(logger, mockLedger, stablecoin, compliance, secret)
		
		req := ledger.OracleWebhookRequest{
			EscrowID:       "escrow-123",
			Signature:      "invalid-sig",
		}
		
		err := svc.ProcessOracleWebhook(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})

	t.Run("Mismatched Milestone Index", func(t *testing.T) {
		mockLedger := new(ledger.MockLedgerClient)
		stablecoin := ledger.NewJsonStablecoinProvider(logger, mockLedger)
		svc := NewEscrowService(logger, mockLedger, stablecoin, compliance, secret)
		
		req := ledger.OracleWebhookRequest{
			EscrowID:       "escrow-123",
			MilestoneIndex: 1, // Webhook says 1
			Event:          "DELIVERED",
			OracleProvider: "FedEx",
		}
		
		// Generate valid signature for the mismatching request
		payload := fmt.Sprintf("%s|%d|%s|%s", req.EscrowID, req.MilestoneIndex, req.Event, req.OracleProvider)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(payload))
		req.Signature = hex.EncodeToString(h.Sum(nil))
		
		// Contract is still at index 0
		mockLedger.On("GetEscrow", mock.Anything, "escrow-123", ledger.EscrowMediatorUser).Return(&ledger.EscrowContract{
			ID:                    "escrow-123",
			CurrentMilestoneIndex: 0,
			State:                 "ACTIVE",
		}, nil)
		
		err := svc.ProcessOracleWebhook(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match current escrow milestone index")
	})
}
