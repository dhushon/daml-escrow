//go:build integration
// +build integration

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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

	t.Run("Escrow Refund Lifecycle (Buyer Initiated)", func(t *testing.T) {
		// 1. Create Escrow
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   150.0,
			Currency: "EUR",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)

		// 2. Automated Refund
		t.Log("Testing RefundBuyer (Automated Raise+Resolve)...")
		err = client.RefundBuyer(ctx, escrow.ID)
		require.NoError(t, err)
		t.Log("Successfully exercised RefundBuyer")

		// 3. Verify Original is gone
		_, err = client.GetEscrow(ctx, escrow.ID)
		require.Error(t, err)
		t.Log("Verified original escrow is archived after refund")
	})

	t.Run("Escrow Refund Lifecycle (Seller Initiated)", func(t *testing.T) {
		// 1. Create Escrow
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   300.0,
			Currency: "GBP",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)

		// 2. Seller Proactive Refund
		t.Log("Testing RefundBySeller...")
		err = client.RefundBySeller(ctx, escrow.ID)
		require.NoError(t, err)
		t.Log("Successfully exercised RefundBySeller")

		// 3. Verify Original is gone
		_, err = client.GetEscrow(ctx, escrow.ID)
		require.Error(t, err)
		t.Log("Verified original escrow is archived after seller refund")
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
		time.Sleep(2 * time.Second)
		contracts, err := client.ListEscrows(ctx, BuyerUser)
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

	t.Run("Full Settlement Lifecycle", func(t *testing.T) {
		// 1. Create Escrow
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   1000.0,
			Currency: "USD",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)

		// 2. Approve Milestone (Creates Settlement)
		t.Log("Approving milestone to create settlement...")
		err = client.ReleaseFunds(ctx, escrow.ID)
		require.NoError(t, err)

		// 3. List Settlements
		t.Log("Listing pending settlements...")
		time.Sleep(2 * time.Second)
		settlements, err := client.ListSettlements(ctx)
		require.NoError(t, err)
		
		var mySettlement *EscrowSettlement
		for _, s := range settlements {
			if s.Recipient == client.getParty(SellerUser) && s.Amount == 1000.0 {
				mySettlement = s
				break
			}
		}
		require.NotNil(t, mySettlement, "Could not find pending settlement")
		t.Logf("Found settlement ID: %s", mySettlement.ID)

		// 4. Settle Payment
		t.Log("Finalizing settlement (CentralBank)...")
		err = client.SettlePayment(ctx, mySettlement.ID)
		require.NoError(t, err)
		t.Log("Successfully finalized settlement")

		// 5. Verify Settlement is archived
		time.Sleep(1 * time.Second)
		settlements2, err := client.ListSettlements(ctx)
		require.NoError(t, err)
		found := false
		for _, s := range settlements2 {
			if s.ID == mySettlement.ID {
				found = true
				break
			}
		}
		require.False(t, found, "Settlement contract should be archived")
		t.Log("Verified settlement archived")
	})

	t.Run("Role-Based Visibility and Metrics", func(t *testing.T) {
		// 1. Setup - Create an escrow
		_, err := client.CreateEscrow(ctx, CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   777.0,
			Currency: "USD",
		})
		require.NoError(t, err)

		// 2. Verify visibility for Seller
		t.Log("Checking visibility for Seller...")
		sellerEscrows, err := client.ListEscrows(ctx, SellerUser)
		require.NoError(t, err)
		found := false
		for _, e := range sellerEscrows {
			if e.Amount == 777.0 {
				found = true
				break
			}
		}
		require.True(t, found, "Seller should see the escrow")

		// 3. Verify visibility for Bank
		t.Log("Checking visibility for Bank...")
		bankEscrows, err := client.ListEscrows(ctx, CentralBankUser)
		require.NoError(t, err)
		foundBank := false
		for _, e := range bankEscrows {
			if e.Amount == 777.0 {
				foundBank = true
				break
			}
		}
		require.True(t, foundBank, "Bank should see the escrow as signatory")

		// 4. Verify Metrics
		t.Log("Checking aggregated metrics for Bank...")
		metrics, err := client.GetMetrics(ctx, CentralBankUser)
		require.NoError(t, err)
		require.GreaterOrEqual(t, metrics.TotalActiveEscrows, 1)
		require.GreaterOrEqual(t, metrics.TotalValueInEscrow, 777.0)
		t.Logf("Bank Metrics: %+v", metrics)
	})

	t.Run("Metadata and Anonymous Counterparties", func(t *testing.T) {
		// 1. Create Escrow with rich metadata and placeholder parties
		metadata := map[string]interface{}{
			"opportunityId":    "O-GRANT-999",
			"trackingNumber":   "SHIP-12345",
			"carrier":          "FedEx",
			"customInternalId": "INT-777",
		}
		
		createReq := CreateEscrowRequest{
			Buyer:      "AnonymousBuyer", // Opaque ID
			Seller:     "AnonymousSeller",
			Amount:     100.0,
			Currency:   "USD",
			Metadata:   metadata,
		}
		
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		
		// 2. Fetch and Verify Metadata
		fetched, err := client.GetEscrow(ctx, escrow.ID)
		require.NoError(t, err)
		require.Equal(t, "O-GRANT-999", fetched.Metadata["opportunityId"])
		require.Equal(t, "FedEx", fetched.Metadata["carrier"])
		require.Equal(t, "INT-777", fetched.Metadata["customInternalId"])
		
		t.Logf("Verified rich metadata: %+v", fetched.Metadata)
	})
}
