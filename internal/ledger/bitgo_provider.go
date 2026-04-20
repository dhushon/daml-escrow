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

// BitGoStablecoinProvider implements the StablecoinProvider interface using the BitGo v2 API.
type BitGoStablecoinProvider struct {
	logger      *zap.Logger
	httpClient  *http.Client
	expressURL  string
	accessToken string
	enterprise  string
	coin        string
}

func NewBitGoStablecoinProvider(logger *zap.Logger, expressURL, accessToken, enterprise, coin string) *BitGoStablecoinProvider {
	return &BitGoStablecoinProvider{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		expressURL:  expressURL,
		accessToken: accessToken,
		enterprise:  enterprise,
		coin:        coin,
	}
}

func (p *BitGoStablecoinProvider) EnsureVault(ctx context.Context, userID string) (string, error) {
	// For BitGo, a 'Vault' is an Enterprise Wallet.
	// In a real implementation, we might look up an existing wallet by label or metadata.
	p.logger.Info("ensuring BitGo vault", zap.String("userID", userID))
	
	// Mock lookup: In production, this would query /api/v2/{coin}/wallet
	// For now, we return a deterministic ID based on the user if we can't find one.
	return fmt.Sprintf("bg-wallet-%s", userID), nil
}

func (p *BitGoStablecoinProvider) GetBalance(ctx context.Context, vaultID string, currency string) (float64, error) {
	url := fmt.Sprintf("%s/api/v2/%s/wallet/%s", p.expressURL, p.coin, vaultID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("bitgo request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("bitgo error (%d): %s", resp.StatusCode, string(body))
	}

	var wallet struct {
		Balance          int64 `json:"balance"`
		ConfirmedBalance int64 `json:"confirmedBalance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&wallet); err != nil {
		return 0, fmt.Errorf("failed to decode bitgo response: %w", err)
	}

	// BitGo returns balances in base units (e.g. satoshis or micro-USDC).
	// For USDC (6 decimals), we divide by 1,000,000.
	return float64(wallet.ConfirmedBalance) / 1000000.0, nil
}

func (p *BitGoStablecoinProvider) Transfer(ctx context.Context, fromID, toID string, amount float64, currency string, reference string) (string, error) {
	url := fmt.Sprintf("%s/api/v2/%s/wallet/%s/sendcoins", p.expressURL, p.coin, fromID)
	
	// Convert amount to base units (assuming 6 decimals for USDC)
	baseAmount := int64(amount * 1000000)

	body := map[string]interface{}{
		"address":          toID, // In BitGo, this could be a wallet ID or a destination address
		"amount":           fmt.Sprintf("%d", baseAmount),
		"comment":          reference,
		"walletPassphrase": "REDACTED", // This MUST be handled securely (e.g. from a Vault/KMS)
	}

	jsonBody, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("bitgo transfer failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bitgo transfer error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		TxID string `json:"txid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode transfer response: %w", err)
	}

	return result.TxID, nil
}

func (p *BitGoStablecoinProvider) VerifyTransfer(ctx context.Context, transferID string) (bool, error) {
	url := fmt.Sprintf("%s/api/v2/%s/wallet/transfers/%s", p.expressURL, p.coin, transferID)
	
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", "Bearer "+p.accessToken)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	var transfer struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&transfer); err != nil {
		return false, err
	}

	return transfer.Status == "confirmed", nil
}

var _ StablecoinProvider = (*BitGoStablecoinProvider)(nil)
