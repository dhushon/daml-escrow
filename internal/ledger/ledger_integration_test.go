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

func TestLedgerIntegration_Sandbox(t *testing.T) {
	// Sandbox typically runs on port 7575
	ledgerHost := os.Getenv("LEDGER_HOST")
	if ledgerHost == "" {
		ledgerHost = "127.0.0.1"
	}

	logger, _ := zap.NewDevelopment()
	
	// In Sandbox (single node), all logical participants are hosted on the same node.
	sandboxClient := NewJsonLedgerClient(logger, ledgerHost, 7575, "stablecoin-escrow", "stablecoin-escrow-interfaces")
	sandboxClient.Verbose = false
	
	// We use MultiLedgerClient even for Sandbox to maintain interface consistency,
	// but all node keys point to the same physical sandbox node.
	client := NewMultiLedgerClient(logger, map[string]Client{
		"bank":   sandboxClient,
		"buyer":  sandboxClient,
		"seller": sandboxClient,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	// Perform discovery
	err := client.Discover(ctx, true)
	require.NoError(t, err)

	// Run shared lifecycle logic
	runFullEscrowLifecycle(t, ctx, client)
}
