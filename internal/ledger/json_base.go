package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
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
	mu                 sync.RWMutex     // Protects partyMap
	partyMap           map[string]string // Maps User ID -> Canton Party ID
	PackageID          string            // Instance-specific Package ID
	InterfacePackageID string            // Instance-specific Interface ID
	ImplName           string            // Logical name for implementation DAR
	InterfaceName      string            // Logical name for interface DAR
}

func NewJsonLedgerClient(logger *zap.Logger, host string, port int, implName, ifaceName string) *JsonLedgerClient {
	if host == "localhost" {
		host = "127.0.0.1"
	}
	
	c := &JsonLedgerClient{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
		baseURL:       fmt.Sprintf("http://%s:%d", host, port),
		partyMap:      make(map[string]string),
		ImplName:      implName,
		InterfaceName: ifaceName,
	}
	
	return c
}

func (c *JsonLedgerClient) Discover(ctx context.Context) error {
	c.logger.Info("performing ledger discovery...")

	// 1. Try to load from ledger-state.json (created by make sync)
	if data, err := os.ReadFile("ledger-state.json"); err == nil {
		var state struct {
			PackageID          string            `json:"packageId"`
			InterfacePackageID string            `json:"interfacePackageId"`
			Parties            map[string]string `json:"parties"`
		}
		if err := json.Unmarshal(data, &state); err == nil {
			c.PackageID = state.PackageID
			c.InterfacePackageID = state.InterfacePackageID
			c.mu.Lock()
			for k, v := range state.Parties {
				c.partyMap[k] = v
			}
			c.mu.Unlock()
			c.logger.Info("loaded ledger state from ledger-state.json", 
				zap.String("packageId", c.PackageID))
			return nil
		}
	}

	// 2. Fallback to active discovery if file missing or invalid
	c.logger.Info("ledger-state.json not found or invalid, performing active discovery...")

	// ... (Rest of existing discovery logic)

	var pids []string
	var listResponse struct {
		PackageIds []string `json:"packageIds"`
	}
	if err := json.Unmarshal(respBody, &listResponse); err == nil && len(listResponse.PackageIds) > 0 {
		pids = listResponse.PackageIds
	} else {
		var altResponse struct {
			PackageDetails []struct {
				PackageId string `json:"packageId"`
			} `json:"packageDetails"`
		}
		if err2 := json.Unmarshal(respBody, &altResponse); err2 == nil && len(altResponse.PackageDetails) > 0 {
			for _, d := range altResponse.PackageDetails {
				pids = append(pids, d.PackageId)
			}
		}
	}

	if len(pids) == 0 {
		return fmt.Errorf("no packages found on ledger")
	}

	// 2. Query each package for metadata to find logical names
	for _, pid := range pids {
		path := fmt.Sprintf("/v2/packages/%s", pid)
		pkgBody, err := c.doRawRequest(ctx, "GET", path, nil)
		if err != nil {
			continue
		}

		// Try to find packageName in the response
		var pkgMap map[string]interface{}
		if err := json.Unmarshal(pkgBody, &pkgMap); err == nil {
			if details, ok := pkgMap["packageDetails"].(map[string]interface{}); ok {
				if name, ok := details["packageName"].(string); ok {
					if name == c.ImplName {
						c.PackageID = pid
						c.logger.Info("discovered package", zap.String("name", name), zap.String("id", pid))
					} else if name == c.InterfaceName {
						c.InterfacePackageID = pid
						c.logger.Info("discovered interface package", zap.String("name", name), zap.String("id", pid))
					}
				}
			}
		}
	}

	if c.PackageID == "" || c.InterfacePackageID == "" {
		c.logger.Warn("dynamic discovery failed, using authoritative fallbacks", 
			zap.String("implName", c.ImplName), 
			zap.String("ifaceName", c.InterfaceName))
		if c.PackageID == "" {
			c.PackageID = "18d08e81f601958adc214a33970fa3a06729cad1f4283d873cf0799ab77ac878"
		}
		if c.InterfacePackageID == "" {
			c.InterfacePackageID = "c27305f41570a49eb794b3dcca9d5723334829719eba5752c6bca134df19f95b"
		}
	}

	// 3. Resolve Party IDs
	return c.refreshPartyMap(ctx)
}

func (c *JsonLedgerClient) GetPartyMap() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to ensure thread safety
	m := make(map[string]string)
	for k, v := range c.partyMap {
		m[k] = v
	}
	return m
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

// v2TransactionResponse is shared between escrow and settlement logic
type v2TransactionResponse struct {
	Transaction struct {
		Events []map[string]interface{} `json:"events"`
		Offset json.RawMessage          `json:"offset"`
	} `json:"transaction"`
}
