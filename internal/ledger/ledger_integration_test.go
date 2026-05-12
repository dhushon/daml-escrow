//go:build integration
// +build integration

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
	ledgerHost := os.Getenv("LEDGER_HOST")
	if ledgerHost == "" {
		ledgerHost = "127.0.0.1"
	}

	logger, _ := zap.NewDevelopment()
	
	// Create specific clients for each institutional node
	bankClient := NewJsonLedgerClient(logger, ledgerHost, 7575, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	depClient  := NewJsonLedgerClient(logger, ledgerHost, 7576, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	benClient  := NewJsonLedgerClient(logger, ledgerHost, 7577, "stablecoin-escrow", "stablecoin-escrow-interfaces")

	client := NewMultiLedgerClient(logger, map[string]Client{
		"bank":        bankClient,
		"depositor":   depClient,
		"beneficiary": benClient,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Perform discovery across all nodes
	err := client.Discover(ctx, true)
	require.NoError(t, err)

	// Run shared lifecycle logic (updated for institutional vocabulary)
	runFullEscrowLifecycle(t, ctx, client)
}
