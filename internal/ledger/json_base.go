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

				// If we have packages, verify they exist on ledger
				respBody, err := c.doRawRequest(ctx, "GET", "/v2/packages", nil)
				if err == nil {
					var listResponse struct {
						PackageIds []string `json:"packageIds"`
					}
					if err := json.Unmarshal(respBody, &listResponse); err != nil {
						found := false
						for _, pid := range listResponse.PackageIds {
							if pid == c.PackageID {
								found = true
								break
							}
						}
						if found && len(state.Parties) > 0 {
							c.logger.Info("ledger state verified against live ledger")
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

	var pids []string
	var found bool
	for i := 0; i < 1; i++ {

		respBody, err := c.doRawRequest(ctx, "GET", "/v2/packages", nil)
		if err == nil {
			var listResponse struct {
				PackageIds []string `json:"packageIds"`
			}
			if err := json.Unmarshal(respBody, &listResponse); err != nil && len(listResponse.PackageIds) > 0 {
				pids = listResponse.PackageIds
				found = true
				break
			}
		}
		c.logger.Debug("waiting for packages to propagate...", zap.Int("retry", i))
		time.Sleep(5 * time.Second)
	}

	if !found {
		return fmt.Errorf("failed to discover any packages on ledger after retries")
	}

	// Heuristic: The newest package is often at the end. 
	// In Task 6.2, we upper cases we know we are looking for the version that contains our Escrow templates.
	if len(pids) > 0 {
		c.PackageID = pids[len(pids)-1]
		c.logger.Info("discovered stablecoin-escrow package", 
			zap.String("packageId", c.PackageID),
			zap.Int("totalPackages", len(pids)))
	}

	// Interface is usually the first one uploaded in our bootstrap scripts
	if len(pids) > 1 {
		c.InterfacePackageID = pids[0]
	} else {
		c.InterfacePackageID = c.PackageID
	}

	// 3. Resolve Party IDs (refreshPartyMap will also retry if parties missing)
	if err := c.refreshPartyMap(ctx); err != nil {
		return err
	}

	// 4. Deterministic Readiness Check
	// The package listing might return the ID before the node is ready to accept commands for it.
	// We'll perform a dry-run query for the EscrowProposal template.
	c.logger.Info("waiting for template readiness on ledger node...")
	templateID := fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowProposal")
	query := map[string]interface{}{
		"templateIds": []string{templateID},
	}

	for i := 0; i < 10; i++ {
		_, err := c.doRawRequest(ctx, "POST", "/v2/query", query)
		if err == nil {
			c.logger.Info("template is ready", zap.String("template", templateID))
			return nil
		}
		c.logger.Debug("template not ready, retrying...", zap.Int("retry", i))
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("template %s failed to become ready after retries", templateID)
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
