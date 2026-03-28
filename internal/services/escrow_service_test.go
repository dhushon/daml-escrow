package services

import (
	"context"
	"testing"
	"time"

	"daml-escrow/internal/ledger"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestEscrowService_Unit(t *testing.T) {
	logger := zap.NewNop()
	mockLedger := new(MockLedgerClient)
	mockStable := new(MockStablecoinProvider)
	mockCompliance := new(MockCompliance)
	secret := "test-secret"

	svc := NewEscrowService(logger, mockLedger, mockStable, mockCompliance, secret)
	ctx := context.Background()

	t.Run("ProposeEscrow", func(t *testing.T) {
		req := ledger.CreateEscrowRequest{
			Seller: "seller1",
			Asset:  ledger.Asset{Amount: 100},
		}
		expected := &ledger.EscrowProposal{ID: "prop1"}
		mockLedger.On("ProposeEscrow", ctx, req).Return(expected, nil)

		resp, err := svc.ProposeEscrow(ctx, req)
		assert.NoError(t, err)
		assert.Equal(t, expected, resp)
		mockLedger.AssertExpectations(t)
	})

	t.Run("Fund", func(t *testing.T) {
		mockLedger.On("Fund", ctx, "esc1", "ref1", "cid1", "user1").Return(nil)
		err := svc.Fund(ctx, "esc1", "ref1", "cid1", "user1")
		assert.NoError(t, err)
		mockLedger.AssertExpectations(t)
	})

	t.Run("Activate", func(t *testing.T) {
		mockLedger.On("Activate", ctx, "esc1", "user1").Return("new-id", nil)
		id, err := svc.Activate(ctx, "esc1", "user1")
		assert.NoError(t, err)
		assert.Equal(t, "new-id", id)
		mockLedger.AssertExpectations(t)
	})

	t.Run("ConfirmConditions", func(t *testing.T) {
		mockLedger.On("ConfirmConditions", ctx, "esc1", "user1").Return(nil)
		err := svc.ConfirmConditions(ctx, "esc1", "user1")
		assert.NoError(t, err)
		mockLedger.AssertExpectations(t)
	})

	t.Run("RaiseDispute", func(t *testing.T) {
		mockLedger.On("RaiseDispute", ctx, "esc1", "user1").Return(nil)
		err := svc.RaiseDispute(ctx, "esc1", "user1")
		assert.NoError(t, err)
		mockLedger.AssertExpectations(t)
	})

	t.Run("Invitation Flow", func(t *testing.T) {
		asset := ledger.Asset{Amount: 500}
		terms := ledger.EscrowTerms{ExpiryDate: time.Now()}
		expected := &ledger.EscrowInvitation{ID: "inv1"}
		
		mockLedger.On("CreateInvitation", ctx, "inviter", "invitee@test.com", "Buyer", "Individual", asset, terms).Return(expected, nil)
		
		inv, err := svc.CreateInvitation(ctx, "inviter", "invitee@test.com", "Buyer", "Individual", asset, terms)
		assert.NoError(t, err)
		assert.Equal(t, expected, inv)
		mockLedger.AssertExpectations(t)
	})
}
