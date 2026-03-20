package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
)

type WebhookRequest struct {
	EscrowID       string                 `json:"escrowId"`
	MilestoneIndex int                    `json:"milestoneIndex"`
	Event          string                 `json:"event"`
	OracleProvider string                 `json:"oracleProvider"`
	Evidence       string                 `json:"evidence"`
	Metadata       map[string]interface{} `json:"metadata"`
	Signature      string                 `json:"signature"`
}

func main() {
	escrowID := flag.String("escrow", "", "Daml Contract ID of the escrow")
	index := flag.Int("milestone", 0, "Index of the milestone to approve")
	event := flag.String("event", "DELIVERY_CONFIRMED", "Event type")
	provider := flag.String("provider", "OracleSim-V1", "Oracle provider name")
	secret := flag.String("secret", "development-secret-key", "HMAC shared secret")
	url := flag.String("url", "http://localhost:8080/webhooks/milestone", "Webhook URL")
	
	flag.Parse()

	if *escrowID == "" {
		fmt.Println("Error: -escrow ID is required")
		flag.Usage()
		os.Exit(1)
	}

	// 1. Prepare Payload
	req := WebhookRequest{
		EscrowID:       *escrowID,
		MilestoneIndex: *index,
		Event:          *event,
		OracleProvider: *provider,
		Evidence:       "ipfs://QmProofOfWork123",
		Metadata: map[string]interface{}{
			"simulated": true,
			"timestamp": "2026-03-19T12:00:00Z",
		},
	}

	// 2. Generate Signature
	// Format: escrowId|milestoneIndex|event|oracleProvider
	payload := fmt.Sprintf("%s|%d|%s|%s", req.EscrowID, req.MilestoneIndex, req.Event, req.OracleProvider)
	h := hmac.New(sha256.New, []byte(*secret))
	h.Write([]byte(payload))
	req.Signature = hex.EncodeToString(h.Sum(nil))

	// 3. Send Webhook
	jsonBody, _ := json.Marshal(req)
	resp, err := http.Post(*url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Error sending webhook: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	fmt.Printf("Webhook Sent! Status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
