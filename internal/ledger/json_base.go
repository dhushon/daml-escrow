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
	Verbose            bool              // If true, logs raw JSON bodies (heavy token usage)
	lastOffset         json.RawMessage   // The most recent offset seen from the ledger
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
		Verbose:       os.Getenv("LEDGER_VERBOSE") == "true",
	}
	
	return c
}

func (c *JsonLedgerClient) Discover(ctx context.Context) error {
	c.logger.Info("performing ledger discovery...")

	forceDiscovery := os.Getenv("FORCE_DISCOVERY") == "true"

	// 1. Try to load from ledger-state.json (created by make sync)
	if !forceDiscovery {
		paths := []string{
			"ledger-state.json",
			"../../ledger-state.json",
			"../../../ledger-state.json",
			"../ledger-state.json",
			"internal/ledger/ledger-state.json",
		}
		var data []byte
		var err error
		var foundPath string
		for _, p := range paths {
			data, err = os.ReadFile(p)
			if err == nil {
				foundPath = p
				break
			}
		}

		if err == nil {
			c.logger.Debug("loading ledger state", zap.String("path", foundPath))
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
				c.logger.Info("loaded ledger state",
					zap.String("path", foundPath),
					zap.String("packageId", c.PackageID),
					zap.Int("partyCount", len(state.Parties)))
				
				// Only return early if we also have parties
				if len(state.Parties) > 0 {
					return nil
				}
			}
		}
	}

	// 2. Fallback to active discovery if file missing or invalid
	c.logger.Info("ledger-state.json not found or invalid, performing active discovery...")

	// In JSON API V2, we can't easily get package names via the public API.
	// We'll look for the IDs we know from the DAR build if discovery fails,
	// or we'll assume the latest ones uploaded are ours if we're in a clean env.
	
	respBody, err := c.doRawRequest(ctx, "GET", "/v2/packages", nil)
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	var pids []string
	var listResponse struct {
		PackageIds []string `json:"packageIds"`
	}
	if err := json.Unmarshal(respBody, &listResponse); err == nil && len(listResponse.PackageIds) > 0 {
		pids = listResponse.PackageIds
	}

	if len(pids) == 0 {
		return fmt.Errorf("no packages found on ledger")
	}

	// Strategy: If we can't get names, and we have a very small number of packages,
	// we might be able to guess. But better to rely on ledger-state.json.
	// For now, if discovery is forced, we'll just log that active name discovery is limited.
	c.logger.Warn("active package name discovery is limited in JSON API V2; relying on ledger-state.json is recommended")

	if c.PackageID == "" || c.InterfacePackageID == "" {
		// Last resort: check if we have hardcoded hints for this environment
		if c.PackageID == "" {
			c.PackageID = "07a6865290a4d9ec5879c4b084806f0ebe5f8429902d49e00a631ef24b54b379"
		}
		if c.InterfacePackageID == "" {
			c.InterfacePackageID = "487bf8c3df07f56647eb17dd991de8090d77dd0fc8a20f3e98549fdd3cd45519"
		}
		c.logger.Info("using hardcoded package ID fallbacks", zap.String("packageId", c.PackageID))
	}

	// 3. Resolve Party IDs
	return c.refreshPartyMap(ctx)
}

func (c *JsonLedgerClient) SearchPackageID(ctx context.Context, name string) (string, error) {
	respBody, err := c.doRawRequest(ctx, "GET", "/v2/packages", nil)
	if err != nil {
		return "", err
	}

	var listResponse struct {
		PackageIds []string `json:"packageIds"`
	}
	if err := json.Unmarshal(respBody, &listResponse); err != nil {
		return "", err
	}

	for _, pid := range listResponse.PackageIds {
		pkgBody, err := c.doRawRequest(ctx, "GET", "/v2/packages/"+pid, nil)
		if err == nil {
			// Note: This relies on the JSON API returning binary for /v2/packages/$pid 
			// We check the binary for the package name string as a hack for DevNet
			if bytes.Contains(pkgBody, []byte(name)) {
				return pid, nil
			}
		}
	}
	return "", fmt.Errorf("package not found: %s", name)
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

func (c *JsonLedgerClient) SetPartyMap(m map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range m {
		c.partyMap[k] = v
	}
}

func (c *JsonLedgerClient) getOffset() interface{} {
	return nil
}

func (c *JsonLedgerClient) GetInterfacePackageID() string {
	return c.InterfacePackageID
}

func (c *JsonLedgerClient) DoRawRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	return c.doRawRequest(ctx, method, path, body)
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

	if c.Verbose {
		c.logger.Debug("ledger request (verbose)", zap.String("method", method), zap.String("path", path), zap.Any("body", body))
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	
	// Capture offset if present in response
	var offsetCapture struct {
		Transaction struct {
			Offset json.RawMessage `json:"offset"`
		} `json:"transaction"`
		Offset json.RawMessage `json:"offset"`
	}
	if json.Unmarshal(respBody, &offsetCapture) == nil {
		c.mu.Lock()
		if len(offsetCapture.Transaction.Offset) > 0 {
			c.lastOffset = offsetCapture.Transaction.Offset
		} else if len(offsetCapture.Offset) > 0 {
			c.lastOffset = offsetCapture.Offset
		}
		c.mu.Unlock()
	}

	// Concise logging unless Verbose is enabled
	if c.Verbose {
		c.logger.Debug("ledger response (verbose)", zap.String("method", method), zap.String("path", path), zap.String("body", string(respBody)))
	} else {
		c.logger.Debug("ledger response", zap.String("method", method), zap.String("path", path), zap.Int("status", resp.StatusCode), zap.Int("bodySize", len(respBody)))
	}

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
