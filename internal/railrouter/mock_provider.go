package railrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type MockFiatProvider struct {
	apiURL string
	client *http.Client
}

func NewMockFiatProvider(apiURL string) *MockFiatProvider {
	return &MockFiatProvider{
		apiURL: apiURL,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (m *MockFiatProvider) InitiateTransfer(ctx context.Context, req TransferRequest) (TransferRef, error) {
	ref := TransferRef(fmt.Sprintf("MOCK-TX-%d", time.Now().UnixNano()))
	
	// Simulate asynchronous off-ledger settlement in a background goroutine.
	// After a short delay, this invokes the platform webhook callback to confirm.
	go func() {
		time.Sleep(2 * time.Second)
		
		payload := map[string]interface{}{
			"escrowId":   req.EscrowID,
			"paymentRef": string(ref),
			"status":     "COMPLETED",
			"signature":  "mock-authorized-signature",
		}
		
		bodyBytes, err := json.Marshal(payload)
		if err != nil {
			return
		}
		
		webhookURL := fmt.Sprintf("%s/api/v1/webhooks/fiat-settlement", m.apiURL)
		httpReq, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(bodyBytes))
		if err != nil {
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		
		resp, err := m.client.Do(httpReq)
		if err != nil {
			return
		}
		defer resp.Body.Close()
	}()
	
	return ref, nil
}

func (m *MockFiatProvider) GetStatus(ctx context.Context, ref TransferRef) (TransferStatus, error) {
	return StatusCompleted, nil
}
