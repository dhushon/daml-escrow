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

type MockLedgerClient struct {
	mock.Mock
	ledger.Client
}

func (m *MockLedgerClient) ReleaseFunds(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestProcessOracleWebhook(t *testing.T) {
	secret := "test-secret"
	logger, _ := zap.NewDevelopment()
	
	t.Run("Valid Signature", func(t *testing.T) {
		mockLedger := new(MockLedgerClient)
		svc := NewEscrowService(logger, mockLedger, secret)
		
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
		
		mockLedger.On("ReleaseFunds", mock.Anything, "escrow-123").Return(nil)
		
		err := svc.ProcessOracleWebhook(context.Background(), req)
		assert.NoError(t, err)
		mockLedger.AssertExpectations(t)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		mockLedger := new(MockLedgerClient)
		svc := NewEscrowService(logger, mockLedger, secret)
		
		req := ledger.OracleWebhookRequest{
			EscrowID:       "escrow-123",
			Signature:      "invalid-sig",
		}
		
		err := svc.ProcessOracleWebhook(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature")
	})
}
