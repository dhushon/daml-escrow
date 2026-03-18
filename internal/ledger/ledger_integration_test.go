package ledger

import (
	"context"
	"os"
	"strings"
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
		t.Log("Testing RaiseDispute...")
		disputeId, err := client.RaiseDispute(ctx, escrow.ID)
		require.NoError(t, err)
		require.NotEmpty(t, disputeId)
		t.Logf("Successfully exercised RaiseDispute, new ID: %s", disputeId)

		// 3. Verify Original is gone
		_, err = client.GetEscrow(ctx, escrow.ID)
		require.Error(t, err)
		t.Log("Verified original escrow is archived after dispute")
	})

	t.Run("Escrow with Multi-Milestones", func(t *testing.T) {
		// 1. Create Escrow with 3 milestones
		t.Log("Testing CreateEscrow with milestones...")
		milestones := []Milestone{
			{Label: "Design", Amount: 100.0, Completed: false},
			{Label: "Implementation", Amount: 300.0, Completed: false},
			{Label: "Testing", Amount: 100.0, Completed: false},
		}
		createReq := CreateEscrowRequest{
			Buyer:       BuyerUser,
			Seller:      SellerUser,
			Amount:      500.0,
			Currency:    "USD",
			Description: "Software Project",
			Milestones:  milestones,
		}
		
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		
		// 2. Fetch and Verify
		fetched, err := client.GetEscrow(ctx, escrow.ID)
		require.NoError(t, err)
		require.Equal(t, 3, len(fetched.Milestones))
		require.Equal(t, "Design", fetched.Milestones[0].Label)
		require.Equal(t, 100.0, fetched.Milestones[0].Amount)
		require.Equal(t, 0, fetched.CurrentMilestoneIndex)
		t.Log("Multi-milestone escrow verified")

		// 3. Approve first milestone
		t.Log("Approving first milestone...")
		err = client.ReleaseFunds(ctx, escrow.ID)
		require.NoError(t, err)

		// 4. Verify progress
		// Use listEscrows directly to check for the update
		time.Sleep(2 * time.Second)
		contracts, err := client.listEscrows(ctx)
		require.NoError(t, err)
		
		var updated *EscrowContract
		for _, c := range contracts {
			if strings.Contains(c.Buyer, "Buyer") && strings.Contains(c.Seller, "Seller") && c.CurrentMilestoneIndex == 1 {
				updated = c
				break
			}
		}
		require.NotNil(t, updated, "Could not find updated escrow contract")
		require.True(t, updated.Milestones[0].Completed)
		t.Log("First milestone approval verified")
	})

	t.Run("Mediated Dispute Resolution", func(t *testing.T) {
		// 1. Create Escrow
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   200.0,
			Currency: "USD",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)

		// 2. Raise Dispute
		disputeId, err := client.RaiseDispute(ctx, escrow.ID)
		require.NoError(t, err)

		// 3. Resolve Dispute (50/50 split)
		t.Log("Testing ResolveDispute (Mediator path)...")
		err = client.ResolveDispute(ctx, disputeId, 100.0, 100.0)
		require.NoError(t, err)
		t.Log("Successfully exercised ResolveDispute")

		// 4. Verify Original is gone
		_, err = client.GetEscrow(ctx, escrow.ID)
		require.Error(t, err)
		t.Log("Verified original escrow is archived after resolution")
	})
}
