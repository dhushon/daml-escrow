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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Printf("Discovering ledger state on %s:%d...\n", *host, *port)
	if err := client.Discover(ctx); err != nil {
		fmt.Printf("Discovery failed: %v\n", err)
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
