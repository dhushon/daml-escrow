package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// User IDs mapped in init.canton
const (
	CentralBankUser    = "CentralBank"
	BuyerUser          = "Buyer"
	SellerUser         = "Seller"
	EscrowMediatorUser = "EscrowMediator"
)

type JsonLedgerClient struct {
	logger             *zap.Logger
	httpClient         *http.Client
	baseURL            string
	partyMap           map[string]string // Maps User ID -> Canton Party ID
	PackageID          string            // Instance-specific Package ID
	InterfacePackageID string            // Instance-specific Interface ID
}

func (c *JsonLedgerClient) Discover(ctx context.Context) error {
	c.logger.Info("performing ledger discovery...")

	// 1. Resolve Package IDs by name for this specific instance
	if c.PackageID == "" {
		c.PackageID = "d209d27f09adfc9883015b5f23e89f28df6d507c31846cd09e4f2e2bb8b0726b"
	}
	if c.InterfacePackageID == "" {
		c.InterfacePackageID = "75da980e1b67864b12ca7d4d0f5530faaa20a7361ac44b737e640de70cc84bdb"
	}

	// 2. Resolve Party IDs
	return c.refreshPartyMap(ctx)
}

// v2TransactionResponse is shared between escrow and settlement logic
type v2TransactionResponse struct {
	Transaction struct {
		Events []map[string]interface{} `json:"events"`
		Offset json.RawMessage          `json:"offset"`
	} `json:"transaction"`
}

func NewJsonLedgerClient(logger *zap.Logger, host string, port int) *JsonLedgerClient {
	if host == "localhost" {
		host = "127.0.0.1"
	}
	
	c := &JsonLedgerClient{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL:  fmt.Sprintf("http://%s:%d", host, port),
		partyMap: make(map[string]string),
	}
	
	return c
}

func (c *JsonLedgerClient) getOffset() interface{} {
	return nil
}

func (c *JsonLedgerClient) ListWallets(ctx context.Context, userID string) ([]*Wallet, error) {
	// Mock implementation for Phase 4
	return []*Wallet{
		{ID: "mock-wallet-usd", Owner: userID, Currency: "USD", Balance: 1000.0},
	}, nil
}

func (c *JsonLedgerClient) doRawRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	c.logger.Debug("ledger response", zap.String("path", path), zap.String("body", string(respBody)))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JSON API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
