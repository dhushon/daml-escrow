package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"daml-escrow/internal/crypto"
)

type WebhookRequest struct {
	EscrowID       string                 `json:"escrowId"`
	MilestoneIndex int                    `json:"milestoneIndex"`
	Event          string                 `json:"event"`
	OracleProvider string                 `json:"oracleProvider"`
	Evidence       string                 `json:"evidence"`
	Metadata       map[string]interface{} `json:"metadata"`
	Signature      string                 `json:"signature"`
	Asymmetric     bool                   `json:"asymmetric"`
}

func main() {
	escrowID := flag.String("escrow", "", "Daml Contract ID of the escrow")
	index := flag.Int("milestone", 0, "Index of the milestone to approve")
	event := flag.String("event", "DELIVERY_CONFIRMED", "Event type")
	provider := flag.String("provider", "OracleSim-V1", "Oracle provider name")
	secret := flag.String("secret", "development-secret-key", "HMAC shared secret")
	kmsKey := flag.String("kms-key", "", "GCP KMS Key ID (if set, uses asymmetric signing)")
	url := flag.String("url", "http://localhost:8080/api/v1/webhooks/milestone", "Webhook URL")
	
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
	digest := sha256.Sum256([]byte(payload))

	if *kmsKey != "" {
		// High-Assurance Path: Cloud KMS HSM Signing
		fmt.Printf("Generating asymmetric HSM signature using key: %s\n", *kmsKey)
		ctx := context.Background()
		signer, err := crypto.NewCloudKMSSigner(ctx, *kmsKey)
		if err != nil {
			fmt.Printf("Error initializing KMS signer: %v\n", err)
			os.Exit(1)
		}
		defer func() { _ = signer.Close() }()

		sig, err := signer.Sign(ctx, digest[:])
		if err != nil {
			fmt.Printf("Error signing with KMS: %v\n", err)
			os.Exit(1)
		}
		req.Signature = hex.EncodeToString(sig)
		req.Asymmetric = true
	} else {
		// Standalone Path: HMAC (Legacy/Dev)
		fmt.Println("Generating legacy HMAC signature (Local Dev)")
		h := hmac.New(sha256.New, []byte(*secret))
		h.Write([]byte(payload))
		req.Signature = hex.EncodeToString(h.Sum(nil))
		req.Asymmetric = false
	}

	// 3. Send Webhook
	jsonBody, _ := json.Marshal(req)
	resp, err := http.Post(*url, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		fmt.Printf("Error sending webhook: %v\n", err)
		os.Exit(1)
	}
	defer func() { _ = resp.Body.Close() }()

	fmt.Printf("Webhook Sent! Status: %s\n", resp.Status)
	if resp.StatusCode != http.StatusOK {
		os.Exit(1)
	}
}
