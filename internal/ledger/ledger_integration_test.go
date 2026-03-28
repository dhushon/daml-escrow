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

	logger, _ := zap.NewDevelopment()
	
	// Distributed topology nodes
	bankClient := NewJsonLedgerClient(logger, ledgerHost, 7575, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	buyerClient := NewJsonLedgerClient(logger, ledgerHost, 7576, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	sellerClient := NewJsonLedgerClient(logger, ledgerHost, 7577, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	
	bankClient.Verbose = false
	buyerClient.Verbose = false
	sellerClient.Verbose = false
	
	client := NewMultiLedgerClient(logger, map[string]Client{
		"bank":   bankClient,
		"buyer":  buyerClient,
		"seller": sellerClient,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Perform discovery to resolve Package and Party IDs
	err := client.Discover(ctx)
	require.NoError(t, err)

	waitForEscrowState := func(ctx context.Context, userID, assetID, expectedState string) *EscrowContract {
		for retry := 0; retry < 30; retry++ {
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
		assetID := fmt.Sprintf("HAPPY-%d", time.Now().UnixNano())
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
			Metadata: "{\"schemaUrl\": \"https://stablecoin-escrow.io/schemas/grants.v1.json\", \"payload\": {\"grantId\": \"G-123\"}}",
		}

		proposal, err := client.ProposeEscrow(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, proposal)
		t.Logf("Proposed escrow ID: %s", proposal.ID)

		// 1.5 Seller Accepts Proposal (ID updates)
		t.Log("Step 1.5: SellerAccept...")
		acceptedID, err := client.SellerAccept(ctx, proposal.ID, SellerUser)
		require.NoError(t, err)
		require.NotEmpty(t, acceptedID)
		proposal.ID = acceptedID

		// 2. Create a Mock Holding for the Buyer (Find test package ID dynamically)
		testPkgID, err := client.SearchPackageID(ctx, "stablecoin-escrow-tests")
		require.NoError(t, err)
		holdingCid, err := client.CreateContract(ctx, BuyerUser, fmt.Sprintf("%s:%s:%s", testPkgID, "Test.StablecoinEscrowTest", "MockHolding"), map[string]interface{}{
			"owner":   client.getParty(BuyerUser),
			"amount":  fmt.Sprintf("%.10f", req.Asset.Amount),
			"issuer":  client.getParty(CentralBankUser),
			"assetId": req.Asset.AssetID,
		})
		require.NoError(t, err)
		t.Logf("Created mock holding: %s", holdingCid)

		// 3. Fund (DRAFT -> FUNDED)
		t.Log("Step 2: Fund...")
		err = client.Fund(ctx, proposal.ID, "CUST-REF-999", holdingCid, BuyerUser)
		require.NoError(t, err)

		// Fetch updated state
		escrow := waitForEscrowState(ctx, BuyerUser, assetID, "FUNDED")
		require.NotNil(t, escrow)
		require.Equal(t, "FUNDED", escrow.State)
		require.Equal(t, "CUST-REF-999", escrow.Asset.CustodyRef)

		// 3. Activate (FUNDED -> ACTIVE)
		t.Log("Step 3: Activate...")
		activeID, err := client.Activate(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)
		require.NotEmpty(t, activeID)
		escrow.ID = activeID // Update ID for subsequent steps

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "ACTIVE")
		require.NotNil(t, escrow)
		require.Equal(t, "ACTIVE", escrow.State)

		// 4. Confirm Conditions (ACTIVE -> SETTLED)
		t.Log("Step 4: ConfirmConditions...")
		err = client.ConfirmConditions(ctx, escrow.ID, EscrowMediatorUser)
		require.NoError(t, err)

		// 5. Wait for SETTLED (DisbursementOrder created)
		t.Log("Step 5: Waiting for SETTLED state...")
		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "SETTLED")
		require.NotNil(t, escrow)
		require.Equal(t, "SETTLED", escrow.State)
		require.NotNil(t, escrow)
		require.Equal(t, "SETTLED", escrow.State)

		// 6. Disburse (SETTLED -> Terminal/Archived)
		t.Log("Step 6: Disburse...")
		err = client.Disburse(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)

		// Verify archive
		time.Sleep(2 * time.Second) // wait for indexer
		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "SETTLED")
		require.Nil(t, escrow, "escrow should be archived after disbursement")
		t.Log("Happy path completed successfully")
	})

	t.Run("High-Assurance Dispute & Ratification", func(t *testing.T) {
		assetID := fmt.Sprintf("DISP-%d", time.Now().UnixNano())
		// 1. Setup Active Escrow
		req := CreateEscrowRequest{
			Seller: SellerUser,
			Asset: Asset{AssetType: "stablecoin", AssetID: assetID, Amount: 1000.0, Currency: "USD"},
			Terms: EscrowTerms{ConditionDescription: "Disputed project", ConditionType: "Binary", ExpiryDate: time.Now().Add(30 * 24 * time.Hour)},
		}
		proposal, err := client.ProposeEscrow(ctx, req)
		require.NoError(t, err)
		
		t.Log("Step 1.5: SellerAccept...")
		acceptedID, err := client.SellerAccept(ctx, proposal.ID, SellerUser)
		require.NoError(t, err)
		proposal.ID = acceptedID

		testPkgID, err := client.SearchPackageID(ctx, "stablecoin-escrow-tests")
		require.NoError(t, err)
		holdingCid, err := client.CreateContract(ctx, BuyerUser, fmt.Sprintf("%s:%s:%s", testPkgID, "Test.StablecoinEscrowTest", "MockHolding"), map[string]interface{}{
			"owner":   client.getParty(BuyerUser),
			"amount":  fmt.Sprintf("%.10f", req.Asset.Amount),
			"issuer":  client.getParty(CentralBankUser),
			"assetId": req.Asset.AssetID,
		})
		require.NoError(t, err)
		
		err = client.Fund(ctx, proposal.ID, "REF-DISP", holdingCid, BuyerUser)
		require.NoError(t, err)
		
		escrow := waitForEscrowState(ctx, BuyerUser, assetID, "FUNDED")
		require.NotNil(t, escrow)
		
		activeID, err := client.Activate(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)
		escrow.ID = activeID
		
		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "ACTIVE")
		require.NotNil(t, escrow)

		// 2. Raise Dispute (ACTIVE -> DISPUTED)
		t.Log("Step 1: RaiseDispute...")
		err = client.RaiseDispute(ctx, escrow.ID, BuyerUser)
		require.NoError(t, err, "RaiseDispute failed")

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "DISPUTED")
		require.NotNil(t, escrow, "Failed to reach DISPUTED state")
		require.Equal(t, "DISPUTED", escrow.State)

		// 2. Propose Settlement (DISPUTED -> PROPOSED)
		t.Log("Step 2: ProposeSettlement...")
		settlement := SettlementTerms{
			SettlementType: "PartialSplit",
			BuyerReturn:    400.0,
			SellerPayment:  500.0,
			MediatorFee:    100.0,
		}
		proposedID, err := client.ProposeSettlement(ctx, escrow.ID, settlement, EscrowMediatorUser)
		require.NoError(t, err, "ProposeSettlement failed")
		require.NotEmpty(t, proposedID)
		escrow.ID = proposedID

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "PROPOSED")
		require.NotNil(t, escrow, "Failed to reach PROPOSED state")
		require.Equal(t, "PROPOSED", escrow.State)

		// 4. Ratify Settlement (Buyer)
		t.Log("Step 3: RatifySettlement (Buyer)...")
		ratifiedID, err := client.RatifySettlement(ctx, escrow.ID, BuyerUser)
		require.NoError(t, err, "RatifySettlement (Buyer) failed")
		require.NotEmpty(t, ratifiedID)
		escrow.ID = ratifiedID

		// Wait for BuyerAccepted to be true
		escrow = waitForEscrowCondition(ctx, BuyerUser, assetID, "PROPOSED", func(e *EscrowContract) bool {
			return e.BuyerAccepted
		})
		require.NotNil(t, escrow, "Failed to find PROPOSED escrow with BuyerAccepted=true")
		require.True(t, escrow.BuyerAccepted)
		require.False(t, escrow.SellerAccepted)

		// 5. Ratify Settlement (Seller)
		t.Log("Step 4: RatifySettlement (Seller)...")
		ratifiedID2, err := client.RatifySettlement(ctx, escrow.ID, SellerUser)
		require.NoError(t, err, "RatifySettlement (Seller) failed")
		require.NotEmpty(t, ratifiedID2)
		escrow.ID = ratifiedID2

		// Wait for SellerAccepted to be true
		escrow = waitForEscrowCondition(ctx, BuyerUser, assetID, "PROPOSED", func(e *EscrowContract) bool {
			return e.SellerAccepted
		})
		require.NotNil(t, escrow, "Failed to find PROPOSED escrow with SellerAccepted=true")
		require.True(t, escrow.SellerAccepted)

		// 6. Finalize (PROPOSED -> SETTLED)
		t.Log("Step 5: FinalizeSettlement...")
		settledID, err := client.FinalizeSettlement(ctx, escrow.ID, EscrowMediatorUser)
		require.NoError(t, err, "FinalizeSettlement failed")
		require.NotEmpty(t, settledID)
		escrow.ID = settledID

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
		proposal, err := client.ProposeEscrow(ctx, req)
		require.NoError(t, err)

		t.Log("Step 1.5: SellerAccept...")
		acceptedID, err := client.SellerAccept(ctx, proposal.ID, SellerUser)
		require.NoError(t, err)
		proposal.ID = acceptedID
		
		testPkgID, err := client.SearchPackageID(ctx, "stablecoin-escrow-tests")
		holdingCid, _ := client.CreateContract(ctx, BuyerUser, fmt.Sprintf("%s:%s:%s", testPkgID, "Test.StablecoinEscrowTest", "MockHolding"), map[string]interface{}{
			"owner":   client.getParty(BuyerUser),
			"amount":  fmt.Sprintf("%.10f", req.Asset.Amount),
			"issuer":  client.getParty(CentralBankUser),
			"assetId": req.Asset.AssetID,
		})
		
		client.Fund(ctx, proposal.ID, "REF-TIMEOUT", holdingCid, BuyerUser)
		
		escrow := waitForEscrowState(ctx, BuyerUser, assetID, "FUNDED")
		require.NotNil(t, escrow)
		
		activeID, err := client.Activate(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)
		escrow.ID = activeID
		
		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "ACTIVE")
		require.NotNil(t, escrow)

		t.Log("Exercising ExpireEscrow (Issuer path)...")
		settledID, err := client.ExpireEscrow(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)
		require.NotEmpty(t, settledID)
		escrow.ID = settledID

		escrow = waitForEscrowState(ctx, BuyerUser, assetID, "SETTLED")
		require.NotNil(t, escrow)
		require.Equal(t, "SETTLED", escrow.State)
		
		err = client.Disburse(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)
		t.Log("Timeout path completed successfully")
	})
}
