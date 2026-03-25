//go:build integration
// +build integration

package ledger

import (
	"context"
	"fmt"
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
	client := NewJsonLedgerClient(logger, ledgerHost, ledgerPort, "stablecoin-escrow", "stablecoin-escrow-interfaces")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Perform discovery to resolve Package and Party IDs
	err := client.Discover(ctx)
	require.NoError(t, err)

	waitForEscrowState := func(ctx context.Context, userID, assetID, expectedState string) *EscrowContract {
		for retry := 0; retry < 20; retry++ {
			escrows, _ := client.ListEscrows(ctx, userID)
			for _, e := range escrows {
				if e.Asset.AssetID == assetID && e.State == expectedState {
					return e
				}
			}
			time.Sleep(2 * time.Second)
		}
		return nil
	}

	waitForEscrowCondition := func(ctx context.Context, userID, assetID, expectedState string, cond func(*EscrowContract) bool) *EscrowContract {
		for retry := 0; retry < 20; retry++ {
			escrows, _ := client.ListEscrows(ctx, userID)
			for _, e := range escrows {
				if e.Asset.AssetID == assetID && e.State == expectedState {
					if cond(e) {
						return e
					}
				}
			}
			time.Sleep(2 * time.Second)
		}
		return nil
	}

	t.Run("Identity JIT Provisioning", func(t *testing.T) {
		t.Log("Testing ProvisionUser for external identity...")
		uniqueID := time.Now().UnixNano()
		googleSub := fmt.Sprintf("google-oauth2|%d", uniqueID)
		email := fmt.Sprintf("tester-%d@datacloud.com", uniqueID)

		// 1. Provision User
		identity, err := client.ProvisionUser(ctx, googleSub, email)
		require.NoError(t, err)
		require.NotNil(t, identity)
		require.Contains(t, identity.DamlUserID, "google-oauth2")
		require.NotEmpty(t, identity.DamlPartyID)

		// 2. Fetch Identity
		fetched, err := client.GetIdentity(ctx, googleSub)
		require.NoError(t, err)
		require.NotNil(t, fetched)
		require.Equal(t, identity.DamlPartyID, fetched.DamlPartyID)
		
		t.Logf("Successfully provisioned identity: %s", identity.DamlPartyID)
	})

	t.Run("High-Assurance Happy Path Lifecycle", func(t *testing.T) {
		assetID := "USD-STABLE-001"
		// 1. Propose Escrow (DRAFT)
		t.Log("Step 1: ProposeEscrow...")
		req := CreateEscrowRequest{
			Seller: SellerUser,
			Asset: Asset{
				AssetType: "stablecoin",
				AssetID:   assetID,
				Amount:    1000.0,
				Currency:  "USD",
			},
			Terms: EscrowTerms{
				ConditionDescription: "Deliver high-assurance software",
				ConditionType:        "Binary",
				EvidenceRequired:     "MediatorAttestation",
				ExpiryDate:           time.Now().Add(30 * 24 * time.Hour),
			},
			Metadata: EscrowMetadata{
				SchemaURL: "https://stablecoin-escrow.io/schemas/grants.v1.json",
				Payload:   map[string]interface{}{"grantId": "G-123"},
			},
		}

		proposal, err := client.ProposeEscrow(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, proposal)
		t.Logf("Proposed escrow ID: %s", proposal.ID)

		// 2. Fund (DRAFT -> FUNDED)
		t.Log("Step 2: Fund...")
		err = client.Fund(ctx, proposal.ID, "CUST-REF-999", BuyerUser)
		require.NoError(t, err)

		// Fetch updated state
		escrow := waitForEscrowState(ctx, BuyerUser, assetID, "FUNDED")
		require.NotNil(t, escrow)
		require.Equal(t, "FUNDED", escrow.State)
		require.Equal(t, "CUST-REF-999", escrow.Asset.CustodyRef)

		// 3. Activate (FUNDED -> ACTIVE)
		t.Log("Step 3: Activate...")
		err = client.Activate(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "ACTIVE")
		require.NotNil(t, escrow)
		require.Equal(t, "ACTIVE", escrow.State)

		// 4. Confirm Conditions (ACTIVE -> SETTLED)
		t.Log("Step 4: ConfirmConditions...")
		err = client.ConfirmConditions(ctx, escrow.ID, EscrowMediatorUser)
		require.NoError(t, err)

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "SETTLED")
		require.NotNil(t, escrow)
		require.Equal(t, "SETTLED", escrow.State)

		// 5. Disburse (SETTLED -> Terminal)
		t.Log("Step 5: Disburse...")
		err = client.Disburse(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)

		// Verify archive
		time.Sleep(2 * time.Second) // wait for indexer
		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "SETTLED") // This will timeout and return nil if archived
		require.Nil(t, escrow, "escrow should be archived")
		t.Log("Happy path completed successfully")
	})

	t.Run("High-Assurance Dispute & Ratification", func(t *testing.T) {
		assetID := "USD-STABLE-002"
		// 1. Setup Active Escrow
		req := CreateEscrowRequest{
			Seller: SellerUser,
			Asset: Asset{AssetType: "stablecoin", AssetID: assetID, Amount: 500.0, Currency: "USD"},
			Terms: EscrowTerms{ConditionDescription: "Disputed project", ConditionType: "Binary", ExpiryDate: time.Now().Add(30 * 24 * time.Hour)},
		}
		proposal, _ := client.ProposeEscrow(ctx, req)
		client.Fund(ctx, proposal.ID, "REF-DISP", BuyerUser)
		
		escrow := waitForEscrowState(ctx, BuyerUser, assetID, "FUNDED")
		require.NotNil(t, escrow)
		
		client.Activate(ctx, escrow.ID, CentralBankUser)
		
		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "ACTIVE")
		require.NotNil(t, escrow)

		// 2. Raise Dispute (ACTIVE -> DISPUTED)
		t.Log("Step 1: RaiseDispute...")
		err = client.RaiseDispute(ctx, escrow.ID, BuyerUser)
		require.NoError(t, err, "RaiseDispute failed")

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "DISPUTED")
		require.NotNil(t, escrow, "Failed to reach DISPUTED state")
		require.Equal(t, "DISPUTED", escrow.State)

		// 3. Propose Settlement (DISPUTED -> PROPOSED)
		t.Log("Step 2: ProposeSettlement...")
		settlement := SettlementTerms{
			SettlementType: "PartialSplit",
			BuyerReturn:    200.0,
			SellerPayment:  300.0,
			MediatorFee:    0.0,
		}
		err = client.ProposeSettlement(ctx, escrow.ID, settlement, EscrowMediatorUser)
		require.NoError(t, err, "ProposeSettlement failed")

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "PROPOSED")
		require.NotNil(t, escrow, "Failed to reach PROPOSED state")
		require.Equal(t, "PROPOSED", escrow.State)

		// 4. Ratify Settlement (Buyer)
		t.Log("Step 3: RatifySettlement (Buyer)...")
		err = client.RatifySettlement(ctx, escrow.ID, BuyerUser)
		require.NoError(t, err, "RatifySettlement (Buyer) failed")

		// Wait for BuyerAccepted to be true
		escrow = waitForEscrowCondition(ctx, BuyerUser, assetID, "PROPOSED", func(e *EscrowContract) bool {
			return e.BuyerAccepted
		})
		require.NotNil(t, escrow, "Failed to find PROPOSED escrow with BuyerAccepted=true")
		require.True(t, escrow.BuyerAccepted)
		require.False(t, escrow.SellerAccepted)

		// 5. Ratify Settlement (Seller)
		t.Log("Step 4: RatifySettlement (Seller)...")
		err = client.RatifySettlement(ctx, escrow.ID, SellerUser)
		require.NoError(t, err, "RatifySettlement (Seller) failed")

		// Wait for SellerAccepted to be true
		escrow = waitForEscrowCondition(ctx, BuyerUser, assetID, "PROPOSED", func(e *EscrowContract) bool {
			return e.SellerAccepted
		})
		require.NotNil(t, escrow, "Failed to find PROPOSED escrow with SellerAccepted=true")
		require.True(t, escrow.SellerAccepted)

		// 6. Finalize (PROPOSED -> SETTLED)
		t.Log("Step 5: FinalizeSettlement...")
		err = client.FinalizeSettlement(ctx, escrow.ID, EscrowMediatorUser) // Acts as orchestrator with multi-party rights
		require.NoError(t, err, "FinalizeSettlement failed")

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "SETTLED")
		require.NotNil(t, escrow, "Failed to reach SETTLED state")
		require.Equal(t, "SETTLED", escrow.State)

		// 7. Disburse
		err = client.Disburse(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err, "Disburse failed")
		t.Log("Dispute path completed successfully")
	})

	t.Run("Timeout Refund Path", func(t *testing.T) {
		assetID := "USD-TIMEOUT"
		req := CreateEscrowRequest{
			Seller: SellerUser,
			Asset: Asset{AssetType: "stablecoin", AssetID: assetID, Amount: 100.0, Currency: "USD"},
			Terms: EscrowTerms{ConditionDescription: "Timeout test", ConditionType: "Binary", ExpiryDate: time.Now().Add(30 * 24 * time.Hour)},
		}
		proposal, _ := client.ProposeEscrow(ctx, req)
		client.Fund(ctx, proposal.ID, "REF-TIMEOUT", BuyerUser)
		
		escrow := waitForEscrowState(ctx, BuyerUser, assetID, "FUNDED")
		require.NotNil(t, escrow)
		
		client.Activate(ctx, escrow.ID, CentralBankUser)
		
		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "ACTIVE")
		require.NotNil(t, escrow)

		t.Log("Exercising ExpireEscrow (Issuer path)...")
		err = client.ExpireEscrow(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "SETTLED")
		require.NotNil(t, escrow)
		require.Equal(t, "SETTLED", escrow.State)
		
		err = client.Disburse(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)
		t.Log("Timeout path completed successfully")
	})
}
