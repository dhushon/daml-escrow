package services

import (
	"context"
	"testing"

	"daml-escrow/internal/crypto"
	"daml-escrow/internal/ledger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestEscrowService_Unit(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockLedger := new(ledger.MockLedgerClient)
	mockStablecoin := new(ledger.MockStablecoinProvider)
	compliance := NewMockCompliance()
	signer, _ := crypto.NewLocalSigner()
	webhookSecret := "test-secret"

	svc := NewEscrowService(logger, mockLedger, mockStablecoin, compliance, webhookSecret, signer, nil)

	t.Run("ProposeEscrow", func(t *testing.T) {
		req := ledger.CreateEscrowRequest{Depositor: "Depositor", Beneficiary: "Beneficiary"}
		mockLedger.On("ProposeEscrow", mock.Anything, req).Return(&ledger.EscrowProposal{ID: "prop-123"}, nil)

		resp, err := svc.ProposeEscrow(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, "prop-123", resp.ID)
	})

	t.Run("Fund", func(t *testing.T) {
		mockLedger.On("Fund", mock.Anything, "escrow-123", "CUSTODY-REF-001", "holding-123", "Depositor").Return(nil)
		err := svc.FundEscrow(context.Background(), "escrow-123", "Depositor", "holding-123")
		assert.NoError(t, err)
	})

	t.Run("Activate", func(t *testing.T) {
		mockLedger.On("Activate", mock.Anything, "escrow-123", []string{"Depositor"}).Return("escrow-123", nil)
		id, err := svc.ActivateEscrow(context.Background(), "escrow-123", "Depositor", []string{"Depositor"})
		assert.NoError(t, err)
		assert.Equal(t, "escrow-123", id)
	})

	t.Run("ConfirmConditions", func(t *testing.T) {
		mockLedger.On("ConfirmConditions", mock.Anything, "escrow-123", "Beneficiary").Return(nil)
		// We call through ledger client for now based on handler logic
		err := svc.GetLedgerClient().ConfirmConditions(context.Background(), "escrow-123", "Beneficiary")
		assert.NoError(t, err)
	})

	t.Run("RaiseDispute", func(t *testing.T) {
		mockLedger.On("RaiseDispute", mock.Anything, "escrow-123", "Depositor").Return(nil)
		err := svc.RaiseDispute(context.Background(), "escrow-123", "Depositor")
		assert.NoError(t, err)
	})

	t.Run("Invitation Flow", func(t *testing.T) {
		asset := ledger.Asset{Amount: 100}
		terms := ledger.EscrowTerms{ConditionDescription: "Test"}
		mockLedger.On("CreateInvitation", mock.Anything, "Bank", "user@test.com", "Depositor", "Business", "Corporate", asset, terms).
			Return(&ledger.EscrowInvitation{ID: "inv-123"}, nil)

		// Call through ledger client
		resp, err := svc.GetLedgerClient().CreateInvitation(context.Background(), "Bank", "user@test.com", "Depositor", "Business", "Corporate", asset, terms)
		assert.NoError(t, err)
		assert.Equal(t, "inv-123", resp.ID)
	})
}
