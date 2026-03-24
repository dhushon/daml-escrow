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

// Package-level variables for Daml 3.x integration (resolved at startup)
var (
	PackageID          string
	InterfacePackageID string
)

type JsonLedgerClient struct {
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
	partyMap   map[string]string // Maps User ID -> Canton Party ID (e.g. Buyer -> Buyer::1220...)
}

func (c *JsonLedgerClient) Discover(ctx context.Context) error {
	c.logger.Info("performing ledger discovery...")

	// 1. Resolve Package IDs by name
	// In a real environment, we'd query /v2/packages and check metadata.
	// For this prototype, we'll allow them to be passed in or continue using 
	// the most recent valid IDs if discovery is restricted.
	if PackageID == "" {
		PackageID = "d209d27f09adfc9883015b5f23e89f28df6d507c31846cd09e4f2e2bb8b0726b"
	}
	if InterfacePackageID == "" {
		InterfacePackageID = "75da980e1b67864b12ca7d4d0f5530faaa20a7361ac44b737e640de70cc84bdb"
	}

	// 2. Resolve Party IDs (Enhancing refreshPartyMap)
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
