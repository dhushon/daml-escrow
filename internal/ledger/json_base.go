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

func (c *JsonLedgerClient) Discover(ctx context.Context, wait bool) error {
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

				// If we have packages, verify they exist on ledger
				respBody, err := c.doRawRequest(ctx, "GET", "/v2/packages", nil)
				if err == nil {
					var listResponse struct {
						PackageIds []string `json:"packageIds"`
					}
					if err := json.Unmarshal(respBody, &listResponse); err == nil {
						found := false
						for _, pid := range listResponse.PackageIds {
							if pid == c.PackageID {
								found = true
								break
							}
						}
						if found && len(state.Parties) > 0 {
							c.logger.Info("ledger state verified against live ledger")
							_ = c.refreshPartyMap(ctx)
							return nil
						}
					}
				}
				c.logger.Warn("ledger state file package ID not found on ledger, forcing active discovery")
			}
		}
	}

	// 2. Fallback to active discovery if file missing or invalid
	c.logger.Info("performing active package discovery...")

	pid, err := c.SearchPackageID(ctx, "stablecoin-escrow")
	if err != nil {
		return fmt.Errorf("failed to discover stablecoin-escrow package: %w", err)
	}
	c.PackageID = pid

	ipid, err := c.SearchPackageID(ctx, "stablecoin-escrow-interfaces")
	if err == nil {
		c.InterfacePackageID = ipid
	} else {
		c.InterfacePackageID = c.PackageID
	}

	c.logger.Info("discovered institutional packages", 
		zap.String("implementation", c.PackageID),
		zap.String("interfaces", c.InterfacePackageID))

	// 3. Resolve Party IDs
	if err := c.refreshPartyMap(ctx); err != nil {
		return err
	}

	// 4. Deterministic Readiness Check
	if !wait {
		c.logger.Info("readiness check skipped (non-blocking discovery)")
		return nil
	}

	// The package listing might return the ID before the node is ready to accept commands for it.
	// We'll perform a dry-run fetch for any Active Contracts as a health check for the indexer.
	c.logger.Info("waiting for indexer readiness on ledger node...")
	
	for i := 0; i < 15; i++ {
		// V2 compliant ACS query as a readiness probe
		body := map[string]interface{}{
			"activeAtOffset": "0", // Mandatory in V2 for state queries
			"filter": map[string]interface{}{
				"filtersByParty": map[string]interface{}{
					c.GetParty(CentralBankUser): map[string]interface{}{},
				},
			},
		}
		_, err := c.doRawRequest(ctx, "POST", "/v2/state/active-contracts", body)
		if err == nil {
			c.logger.Info("indexer is ready and responding to tripartite queries")
			return nil
		}
		c.logger.Debug("indexer not ready, retrying...", zap.Int("retry", i), zap.Error(err))
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("ledger indexer failed to become ready after retries")
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
				// High-Assurance: Avoid collision where "stablecoin-escrow" matches "stablecoin-escrow-interfaces" or "stablecoin-escrow-tests"
				if name == "stablecoin-escrow" {
					if bytes.Contains(pkgBody, []byte("stablecoin-escrow-interfaces")) || bytes.Contains(pkgBody, []byte("stablecoin-escrow-tests")) {
						continue
					}
				}
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

func (c *JsonLedgerClient) GetOffset() interface{} {
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
	defer func() { _ = resp.Body.Close() }()

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
