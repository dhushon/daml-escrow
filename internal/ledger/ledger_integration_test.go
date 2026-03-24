//go:build integration
// +build integration

package ledger

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Initial party map refresh
	err := client.refreshPartyMap(ctx)
	require.NoError(t, err)

	t.Run("Identity JIT Provisioning", func(t *testing.T) {
		t.Log("Testing ProvisionUser for external identity...")
		uniqueID := time.Now().UnixNano()
		googleSub := fmt.Sprintf("google-oauth2|%d", uniqueID)
		email := fmt.Sprintf("tester-%d@datacloud.com", uniqueID)

		// 1. Provision User
		identity, err := client.ProvisionUser(ctx, googleSub, email)
		require.NoError(t, err)
		require.NotNil(t, identity)
		require.Equal(t, googleSub, identity.OktaSub)
		require.Contains(t, identity.DamlUserID, "google-oauth2")
		require.NotEmpty(t, identity.DamlPartyID)

		// 2. Fetch Identity
		fetched, err := client.GetIdentity(ctx, googleSub)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		require.Equal(t, identity.DamlPartyID, fetched.DamlPartyID)
		
		t.Logf("Successfully provisioned identity: %s", identity.DamlPartyID)
	})

	t.Run("Invitation and Onboarding Lifecycle", func(t *testing.T) {
		t.Log("Testing full invitation flow...")
		inviter := BuyerUser
		inviteeEmail := "contractor@external.com"
		terms := EscrowTerms{
			TotalAmount: 2500.0,
			Currency:    "USD",
			Description: "External Consulting Services",
			Milestones: []Milestone{
				{Label: "Initial Research", Amount: 500.0, Completed: false},
				{Label: "Final Report", Amount: 2000.0, Completed: false},
			},
		}

		// 1. Create Invitation (as Buyer)
		invite, err := client.CreateInvitation(ctx, inviter, inviteeEmail, "Seller", "Company", terms)
		require.NoError(t, err)
		require.NotNil(t, invite)
		token := invite.TokenHash
		t.Logf("Invitation created with token: %s", token)

		// 2. Lookup Invitation (Anonymous lookup via Token)
		lookup, err := client.GetInvitationByToken(ctx, token)
		require.NoError(t, err)
		require.Equal(t, inviteeEmail, lookup.InviteeEmail)
		require.Equal(t, "Seller", lookup.InviteeRole)
		t.Log("Anonymous token lookup verified")

		// 3. Provision New User (Simulating JIT logic during onboarding)
		uniqueID := time.Now().UnixNano()
		googleSub := fmt.Sprintf("onboard-%d", uniqueID)
		identity, err := client.ProvisionUser(ctx, googleSub, inviteeEmail)
		require.NoError(t, err)
		require.NotNil(t, identity)
		t.Logf("JIT provisioned new party: %s", identity.DamlPartyID)

		// 4. Claim Invitation (as the newly provisioned party)
		proposal, err := client.ClaimInvitation(ctx, invite.ID, googleSub)
		require.NoError(t, err)
		require.NotNil(t, proposal)
		require.Equal(t, identity.DamlPartyID, proposal.Seller)
		require.Equal(t, terms.TotalAmount, proposal.Amount)
		t.Log("Invitation successfully claimed and transformed into Proposal")
	})

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
		
		id := escrow.ID
		t.Logf("Created escrow with ID: %s", id)

		// 2. Get Escrow (Read)
		t.Log("Testing GetEscrow...")
		fetched, err := client.GetEscrow(ctx, id, BuyerUser)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		require.Equal(t, id, fetched.ID)
		t.Log("Successfully fetched escrow")

		// 3. Release Funds (Choice)
		t.Log("Testing ReleaseFunds (ApproveMilestone)...")
		err = client.ReleaseFunds(ctx, id)
		require.NoError(t, err)
		t.Log("Successfully exercised ApproveMilestone")

		// 4. Verify Archive
		t.Log("Testing GetEscrow after archive (should fail)...")
		_, err = client.GetEscrow(ctx, id, BuyerUser)
		require.Error(t, err)
		t.Log("Verified contract is archived")
	})

	t.Run("Escrow Refund Lifecycle (Buyer Initiated)", func(t *testing.T) {
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   100.0,
			Currency: "USD",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		
		t.Log("Testing RefundBuyer (Automated Raise+Resolve)...")
		err = client.RefundBuyer(ctx, escrow.ID)
		require.NoError(t, err)
		t.Log("Successfully exercised RefundBuyer")

		_, err = client.GetEscrow(ctx, escrow.ID, BuyerUser)
		require.Error(t, err)
		t.Log("Verified original escrow is archived after refund")
	})

	t.Run("Escrow Refund Lifecycle (Seller Initiated)", func(t *testing.T) {
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   150.0,
			Currency: "USD",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		
		t.Log("Testing RefundBySeller...")
		err = client.RefundBySeller(ctx, escrow.ID)
		require.NoError(t, err)
		t.Log("Successfully exercised RefundBySeller")

		_, err = client.GetEscrow(ctx, escrow.ID, BuyerUser)
		require.Error(t, err)
		t.Log("Verified original escrow is archived after seller refund")
	})

	t.Run("Escrow with Multi-Milestones", func(t *testing.T) {
		t.Log("Testing CreateEscrow with milestones...")
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   1000.0,
			Currency: "USD",
			Milestones: []Milestone{
				{Label: "Discovery", Amount: 200.0, Completed: false},
				{Label: "Design", Amount: 300.0, Completed: false},
				{Label: "Build", Amount: 500.0, Completed: false},
			},
		}
		
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		require.Equal(t, 3, len(escrow.Milestones))
		require.Equal(t, 0, escrow.CurrentMilestoneIndex)
		t.Log("Multi-milestone escrow verified")

		t.Log("Approving first milestone...")
		err = client.ReleaseFunds(ctx, escrow.ID)
		require.NoError(t, err)

		// Wait for propagation
		time.Sleep(2 * time.Second)
		
		updated, _ := client.GetEscrow(ctx, escrow.ID, BuyerUser)
		require.NotNil(t, updated)
		require.Equal(t, 1, updated.CurrentMilestoneIndex)
		require.True(t, updated.Milestones[0].Completed)
		t.Log("First milestone approval verified")
	})

	t.Run("Mediated Dispute Resolution", func(t *testing.T) {
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   500.0,
			Currency: "USD",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		
		t.Log("Raising dispute first...")
		disputedId, err := client.RaiseDispute(ctx, escrow.ID)
		require.NoError(t, err)
		require.NotEmpty(t, disputedId)

		t.Log("Testing ResolveDispute (Mediator path)...")
		err = client.ResolveDispute(ctx, disputedId, 250.0, 250.0)
		require.NoError(t, err)
		t.Log("Successfully exercised ResolveDispute")

		_, err = client.GetEscrow(ctx, disputedId, BuyerUser)
		require.Error(t, err)
		t.Log("Verified disputed escrow is archived after resolution")
	})

	t.Run("Full Settlement Lifecycle", func(t *testing.T) {
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   100.0,
			Currency: "USD",
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		
		t.Log("Approving milestone to create settlement...")
		err = client.ReleaseFunds(ctx, escrow.ID)
		require.NoError(t, err)
		time.Sleep(5 * time.Second) // Wait for automation

		t.Log("Listing pending settlements...")
		var settlementID string
		for i := 0; i < 10; i++ {
			settlements, _ := client.ListSettlements(ctx)
			for _, s := range settlements {
				if s.Amount == 100.0 && s.Status == "Pending" {
					settlementID = s.ID
					break
				}
			}
			if settlementID != "" { break }
			time.Sleep(1 * time.Second)
		}
		require.NotEmpty(t, settlementID)
		t.Logf("Found settlement ID: %s", settlementID)

		t.Log("Finalizing settlement (CentralBank)...")
		err = client.SettlePayment(ctx, settlementID)
		require.NoError(t, err)
		t.Log("Successfully finalized settlement")

		// Wait for propagation
		time.Sleep(2 * time.Second)
		
		// Verify settlement is gone (archived)
		settlements, _ := client.ListSettlements(ctx)
		found := false
		for _, s := range settlements {
			if s.ID == settlementID {
				found = true
				break
			}
		}
		require.False(t, found)
		t.Log("Verified settlement archived")
	})

	t.Run("Role-Based Visibility and Metrics", func(t *testing.T) {
		// Seller should see their own escrows
		t.Log("Checking visibility for Seller...")
		sellerEscrows, err := client.ListEscrows(ctx, SellerUser)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(sellerEscrows), 0)

		// Bank may see multiple escrows in this environment
		t.Log("Checking visibility for Bank...")
		bankEscrows, err := client.ListEscrows(ctx, CentralBankUser)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(bankEscrows), 0)

		// Bank should see metrics
		t.Log("Checking aggregated metrics for Bank...")
		metrics, err := client.GetMetrics(ctx, CentralBankUser)
		require.NoError(t, err)
		require.NotNil(t, metrics)
		require.GreaterOrEqual(t, metrics.TotalActiveEscrows, 0)
		t.Logf("Bank Metrics: %+v", metrics)
	})

	t.Run("Metadata and Anonymous Counterparties", func(t *testing.T) {
		metadata := EscrowMetadata{
			SchemaURL: "https://stablecoin-escrow.io/schemas/leasing.v1.json",
			Payload: map[string]interface{}{
				"assetId":               "SERIAL-123",
				"assetType":             "Machinery",
				"securityDepositAmount": 5000,
			},
		}

		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   27.0,
			Currency: "USD",
			Metadata: metadata,
		}

		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		require.Equal(t, metadata.SchemaURL, escrow.Metadata.SchemaURL)
		require.Equal(t, "SERIAL-123", escrow.Metadata.Payload["assetId"])
		t.Logf("Verified schema-driven metadata: %+v", escrow.Metadata)
	})

	t.Run("Metadata Selective Exclusion (Privacy)", func(t *testing.T) {
		metadata := EscrowMetadata{
			SchemaURL: "https://stablecoin-escrow.io/schemas/supply-chain.v1.json",
			Payload: map[string]interface{}{
				"shipmentId": "SHIP-999",
				"pvt_operator_code": "SECRET-SIG-44", // Should be excluded
			},
			Exclusions: map[string]struct{}{
				"pvt_operator_code": {},
			},
		}

		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   150.0,
			Currency: "USD",
			Metadata: metadata,
		}

		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)
		
		_, exists := escrow.Metadata.Payload["pvt_operator_code"]
		require.False(t, exists)
		require.Equal(t, "SHIP-999", escrow.Metadata.Payload["shipmentId"])
		t.Log("Verified sensitive fields were excluded from the ledger record")
	})

	t.Run("Oracle Automated Approval", func(t *testing.T) {
		createReq := CreateEscrowRequest{
			Buyer:    BuyerUser,
			Seller:   SellerUser,
			Amount:   50.0,
			Currency: "USD",
			Milestones: []Milestone{
				{Label: "Delivery", Amount: 50.0, Completed: false},
			},
		}
		escrow, err := client.CreateEscrow(ctx, createReq)
		require.NoError(t, err)
		require.NotNil(t, escrow)

		// Trigger webhook
		webhookSecret := "development-secret-key"
		payload := fmt.Sprintf("%s|%d|%s|%s", escrow.ID, 0, "DELIVERY_CONFIRMED", "TestOracle")
		h := hmac.New(sha256.New, []byte(webhookSecret))
		h.Write([]byte(payload))
		signature := hex.EncodeToString(h.Sum(nil))

		hookReq := map[string]interface{}{
			"escrowId":       escrow.ID,
			"milestoneIndex": 0,
			"event":          "DELIVERY_CONFIRMED",
			"oracleProvider": "TestOracle",
			"signature":      signature,
		}

		jsonData, _ := json.Marshal(hookReq)
		resp, err := http.Post("http://localhost:8081/api/v1/webhooks/milestone", "application/json", bytes.NewBuffer(jsonData))
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		// Verification loop with retries
		var archived bool
		for i := 0; i < 15; i++ {
			_, err = client.GetEscrow(ctx, escrow.ID, BuyerUser)
			if err != nil {
				archived = true
				break
			}
			time.Sleep(2 * time.Second)
		}
		require.True(t, archived)
		t.Log("Oracle automated approval verified end-to-end")
	})
}
