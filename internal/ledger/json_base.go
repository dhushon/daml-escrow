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

// Package-level constants for Daml 3.x integration
const (
	PackageID          = "6c43548c421c1e66eb3911379a64d57de18dfe320b679ccb3f84bc7c4028e541"
	InterfacePackageID = "eeada456377e4287fabfe089057b419d54159c87f98da712fd543122fc7c39f3"
)

type JsonLedgerClient struct {
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
	partyMap   map[string]string // Maps User ID -> Canton Party ID (e.g. Buyer -> Buyer::1220...)
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
			Timeout: 30 * time.Second,
		},
		baseURL:  fmt.Sprintf("http://%s:%d", host, port),
		partyMap: make(map[string]string),
	}
	
	// Initial population of the party map
	_ = c.refreshPartyMap(context.Background())
	
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
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JSON API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
