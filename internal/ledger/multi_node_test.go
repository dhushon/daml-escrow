//go:build distributed
// +build distributed

package ledger

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestLedgerIntegration_Distributed(t *testing.T) {
	// Distributed topology nodes (Bank: 7575, Buyer: 7576, Seller: 7577)
	ledgerHost := os.Getenv("LEDGER_HOST")
	if ledgerHost == "" {
		ledgerHost = "127.0.0.1"
	}

	logger, _ := zap.NewDevelopment()
	
	bankClient := NewJsonLedgerClient(logger, ledgerHost, 7575, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	buyerClient := NewJsonLedgerClient(logger, ledgerHost, 7576, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	sellerClient := NewJsonLedgerClient(logger, ledgerHost, 7577, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	
	client := NewMultiLedgerClient(logger, map[string]Client{
		"bank":   bankClient,
		"buyer":  buyerClient,
		"seller": sellerClient,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Perform discovery
	err := client.Discover(ctx, true)
	require.NoError(t, err)

	// Run shared lifecycle logic
	runFullEscrowLifecycle(t, ctx, client)
}
