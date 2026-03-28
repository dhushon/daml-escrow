//go:build distributed
// +build distributed

package ledger

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMultiNodeIntegration(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// 1. Initialize Clients for all 3 nodes
	bankClient := NewJsonLedgerClient(logger, "127.0.0.1", 7575, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	buyerClient := NewJsonLedgerClient(logger, "127.0.0.1", 7576, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	sellerClient := NewJsonLedgerClient(logger, "127.0.0.1", 7577, "stablecoin-escrow", "stablecoin-escrow-interfaces")

	clients := map[string]Client{
		"bank":   bankClient,
		"buyer":  buyerClient,
		"seller": sellerClient,
	}

	multiClient := NewMultiLedgerClient(logger, clients)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// 2. Discover on all nodes
	err := multiClient.Discover(ctx)
	require.NoError(t, err, "Multi-node discovery failed. Ensure docker-compose.distributed.yml is running.")

	waitForEscrowState := func(ctx context.Context, client Client, userID, assetID, expectedState string) *EscrowContract {
		for retry := 0; retry < 30; retry++ {
			escrows, _ := client.ListEscrows(ctx, userID)
			for _, e := range escrows {
				if e.Asset.AssetID == assetID && e.State == expectedState {
					return e
				}
			}
			time.Sleep(1 * time.Second)
		}
		return nil
	}

	t.Run("Cross-Node Happy Path (Bank -> Buyer -> Seller)", func(t *testing.T) {
		assetID := "DIST-USD-001"
		
		// 1. Bank Proposes Escrow (Node 7575)
		t.Log("Step 1: Bank Proposes Escrow...")
		req := CreateEscrowRequest{
			Buyer:  BuyerUser,
			Seller: SellerUser,
			Asset: Asset{
				AssetType: "stablecoin",
				AssetID:   assetID,
				Amount:    5000.0,
				Currency:  "USD",
			},
			Terms: EscrowTerms{
				ConditionDescription: "Distributed escrow test",
				ConditionType:        "Binary",
				ExpiryDate:           time.Now().Add(30 * 24 * time.Hour),
			},
			Metadata: "{\"test\": \"multi-node\"}",
		}

		// Note: We route through multiClient using "CentralBank" as userID
		proposal, err := multiClient.ProposeEscrow(ctx, req) 
		require.NoError(t, err)
		require.NotNil(t, proposal)
		t.Logf("Escrow proposed on Bank node: %s", proposal.ID)

		// 2. Create a Mock Holding for the Buyer (Find test package ID dynamically)
		testPkgID, err := multiClient.SearchPackageID(ctx, "stablecoin-escrow-tests")
		require.NoError(t, err)
		holdingCid, err := multiClient.CreateContract(ctx, BuyerUser, fmt.Sprintf("%s:%s:%s", testPkgID, "Test.StablecoinEscrowTest", "MockHolding"), map[string]interface{}{
			"owner":   multiClient.getParty(BuyerUser),
			"amount":  fmt.Sprintf("%.10f", req.Asset.Amount),
			"issuer":  multiClient.getParty(CentralBankUser),
			"assetId": req.Asset.AssetID,
		})
		require.NoError(t, err)

		// 3. Buyer Funds Escrow (Node 7576)
		// Buyer must see the proposal even though it was created on Bank node
		// (In Canton, proposals are visible to all signatories/observers on their respective nodes)
		t.Log("Step 2: Buyer Funds Escrow (on Buyer Node)...")
		err = multiClient.Fund(ctx, proposal.ID, "DIST-CUST-001", holdingCid, BuyerUser)
		require.NoError(t, err)

		escrow := waitForEscrowState(ctx, multiClient, BuyerUser, assetID, "FUNDED")
		require.NotNil(t, escrow)
		require.Equal(t, "FUNDED", escrow.State)

		// 3. Bank Activates (Node 7575)
		t.Log("Step 3: Bank Activates Escrow (on Bank Node)...")
		err = multiClient.Activate(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)

		escrow = waitForEscrowState(ctx, multiClient, BuyerUser, assetID, "ACTIVE")
		require.NotNil(t, escrow)

		// 4. Mediator Confirms Conditions (Node 7575)
		t.Log("Step 4: Mediator Confirms Conditions (on Bank Node)...")
		err = multiClient.ConfirmConditions(ctx, escrow.ID, EscrowMediatorUser)
		require.NoError(t, err)

		escrow = waitForEscrowState(ctx, multiClient, SellerUser, assetID, "SETTLED")
		require.NotNil(t, escrow)

		// 5. Bank Disburses (Node 7575)
		t.Log("Step 5: Bank Disburses (on Bank Node)...")
		err = multiClient.Disburse(ctx, escrow.ID, CentralBankUser)
		require.NoError(t, err)

		t.Log("Cross-node happy path completed successfully")
	})
}
