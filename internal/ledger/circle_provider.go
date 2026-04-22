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

// CircleStablecoinProvider implements the StablecoinProvider interface using the Circle WaaS API.
type CircleStablecoinProvider struct {
	logger       *zap.Logger
	httpClient   *http.Client
	baseURL      string
	apiKey       string
	entitySecret string
}

func NewCircleStablecoinProvider(logger *zap.Logger, baseURL, apiKey, entitySecret string) *CircleStablecoinProvider {
	if baseURL == "" {
		baseURL = "https://api.circle.com"
	}
	return &CircleStablecoinProvider{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:      baseURL,
		apiKey:       apiKey,
		entitySecret: entitySecret,
	}
}

func (p *CircleStablecoinProvider) EnsureVault(ctx context.Context, userID string) (string, error) {
	// For Circle, a 'Vault' is a Wallet ID.
	// In production, we would use the 'idempotencyKey' (deterministic based on userID)
	// to ensure we only create one wallet per platform user.
	p.logger.Info("ensuring Circle vault", zap.String("userID", userID))
	
	// Mock lookup: In production, this would query /v1/wallets or use a deterministic create.
	return fmt.Sprintf("circle-wallet-%s", userID), nil
}

func (p *CircleStablecoinProvider) GetBalance(ctx context.Context, vaultID string, currency string) (float64, error) {
	url := fmt.Sprintf("%s/v1/wallets/%s/balances", p.baseURL, vaultID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("circle request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("circle error (%d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data []struct {
			Amount   string `json:"amount"`
			Currency string `json:"currency"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, fmt.Errorf("failed to decode circle response: %w", err)
	}

	for _, b := range result.Data {
		if b.Currency == currency {
			var val float64
			_, err := fmt.Sscanf(b.Amount, "%f", &val)
			if err != nil {
				return 0, fmt.Errorf("failed to parse balance amount: %w", err)
			}
			return val, nil
		}
	}

	return 0, nil
}

func (p *CircleStablecoinProvider) Transfer(ctx context.Context, fromID, toID string, amount float64, currency string, reference string) (string, error) {
	url := fmt.Sprintf("%s/v1/developer/transactions/transfer", p.baseURL)
	
	body := map[string]interface{}{
		"idempotencyKey": fmt.Sprintf("escrow-xfer-%s-%d", fromID, time.Now().UnixNano()),
		"entitySecret":   p.entitySecret,
		"amounts": []string{
			fmt.Sprintf("%.6f", amount),
		},
		"sourceWalletId": fromID,
		"destination": map[string]string{
			"type":     "wallet",
			"walletId": toID,
		},
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("circle transfer failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("circle transfer error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode transfer response: %w", err)
	}

	return result.Data.ID, nil
}

func (p *CircleStablecoinProvider) VerifyTransfer(ctx context.Context, transferID string) (bool, error) {
	url := fmt.Sprintf("%s/v1/transactions/%s", p.baseURL, transferID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var result struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	return result.Data.Status == "COMPLETE", nil
}

var _ StablecoinProvider = (*CircleStablecoinProvider)(nil)
