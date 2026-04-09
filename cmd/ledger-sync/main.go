package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"daml-escrow/internal/ledger"
	"go.uber.org/zap"
)

type LedgerState struct {
	GeneratedAt        time.Time         `json:"generatedAt"`
	PackageID          string            `json:"packageId"`
	InterfacePackageID string            `json:"interfacePackageId"`
	Parties            map[string]string `json:"parties"`
}

func main() {
	host := flag.String("host", "localhost", "Ledger host")
	port := flag.Int("port", 7575, "Ledger JSON API port")
	implName := flag.String("impl", "stablecoin-escrow", "Logical name for implementation package")
	ifaceName := flag.String("iface", "stablecoin-escrow-interfaces", "Logical name for interface package")
	output := flag.String("out", "ledger-state.json", "Output file path")
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	client := ledger.NewJsonLedgerClient(logger, *host, *port, *implName, *ifaceName)

	// High-Assurance: Give the synchronization process a long timeout and internal retries
	// to account for ledger startup and indexing lag.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Printf("Discovering ledger state on %s:%d (timeout: 5m)...\n", *host, *port)
	
	var lastErr error
	success := false
	for i := 0; i < 10; i++ {
		if err := client.Discover(ctx, false); err != nil {
			lastErr = err
			fmt.Printf("  [Retry %d/10] Ledger not ready: %v. Waiting 10s...\n", i+1, err)
			select {
			case <-ctx.Done():
				fmt.Printf("Sync timed out: %v\n", ctx.Err())
				os.Exit(1)
			case <-time.After(10 * time.Second):
				continue
			}
		}
		success = true
		break
	}

	if !success {
		fmt.Printf("Discovery failed after all retries: %v\n", lastErr)
		os.Exit(1)
	}

	state := LedgerState{
		GeneratedAt:        time.Now(),
		PackageID:          client.PackageID,
		InterfacePackageID: client.InterfacePackageID,
		Parties:            client.GetPartyMap(), // I'll need to add this getter
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshal state: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(*output, data, 0644); err != nil {
		fmt.Printf("Failed to write to %s: %v\n", *output, err)
		os.Exit(1)
	}

	fmt.Printf("Successfully saved ledger state to %s\n", *output)
	fmt.Printf("Package ID: %s\n", state.PackageID)
}
