package ledger

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLedgerIntegration(t *testing.T) {
	// Skip if no ledger is available
	ledgerHost := os.Getenv("LEDGER_HOST")
	if ledgerHost == "" {
		ledgerHost = "127.0.0.1"
	}
	ledgerPort := 7575 // JSON API port

	logger, _ := zap.NewDevelopment()
	client := NewJsonLedgerClient(logger, ledgerHost, ledgerPort)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	t.Run("Standard Escrow Lifecycle", func(t *testing.T) {
		// 1. Create Escrow (Write)
		t.Log("Testing CreateEscrow...")
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   500.0,
			Currency: "USD",
		}
		
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		require.NotEmpty(t, escrow.ID)
		t.Logf("Created escrow with ID: %s", escrow.ID)

		// 2. Get Escrow (Read)
		t.Log("Testing GetEscrow...")
		fetched, err := client.GetEscrow(ctx, escrow.ID)
		require.NoError(t, err)
		require.Equal(t, escrow.ID, fetched.ID)
		require.Equal(t, 500.0, fetched.Amount)
		t.Log("Successfully fetched escrow")

		// 3. Exercise Choice (Transition)
		t.Log("Testing ReleaseFunds (ApproveMilestone)...")
		err = client.ReleaseFunds(ctx, escrow.ID)
		require.NoError(t, err)
		t.Log("Successfully exercised ApproveMilestone")

		// 4. Verify completion (Final Read)
		// Our DAML code archives the contract if the last milestone is approved
		t.Log("Testing GetEscrow after archive (should fail)...")
		_, err = client.GetEscrow(ctx, escrow.ID)
		require.Error(t, err)
		t.Log("Verified contract is archived")
	})

	t.Run("Escrow Refund Lifecycle", func(t *testing.T) {
		// 1. Create Escrow
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   150.0,
			Currency: "EUR",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)

		// 2. Raise Dispute (Refund Path)
		t.Log("Testing RefundBuyer (RaiseDispute)...")
		err = client.RefundBuyer(ctx, escrow.ID)
		require.NoError(t, err)
		t.Log("Successfully exercised RaiseDispute")

		// 3. Verify Original is gone
		_, err = client.GetEscrow(ctx, escrow.ID)
		require.Error(t, err)
		t.Log("Verified original escrow is archived after dispute")
	})
}
