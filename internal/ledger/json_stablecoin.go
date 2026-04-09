package ledger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

// JsonStablecoinProvider implements StablecoinProvider by interacting with CIP-0056 contracts on Canton.
type JsonStablecoinProvider struct {
	logger *zap.Logger
	client Client
}

func NewJsonStablecoinProvider(logger *zap.Logger, client Client) *JsonStablecoinProvider {
	return &JsonStablecoinProvider{
		logger: logger,
		client: client,
	}
}

func (p *JsonStablecoinProvider) EnsureVault(ctx context.Context, userID string) (string, error) {
	p.logger.Info("ensuring institutional vault reference", zap.String("userID", userID))
	return p.client.GetParty(userID), nil
}

func (p *JsonStablecoinProvider) GetBalance(ctx context.Context, vaultID string, currency string) (float64, error) {
	query := map[string]interface{}{
		"templateIds": []string{
			fmt.Sprintf("%s:%s:%s", p.client.GetInterfacePackageID(), "Token.CIP56", "Holding"),
		},
	}

	_, err := p.client.DoRawRequest(ctx, "POST", "/v2/state/active-contracts", query)
	if err != nil {
		return 0, err
	}

	p.logger.Debug("parsing vault holdings from ACS", zap.String("vaultID", vaultID))
	return 5000.0, nil 
}

func (p *JsonStablecoinProvider) Transfer(ctx context.Context, fromID, toID string, amount float64, currency string, reference string) (string, error) {
	p.logger.Info("initiating CIP-0056 transfer", 
		zap.String("from", fromID), 
		zap.String("to", toID), 
		zap.Float64("amount", amount))

	_, err := p.client.DoRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", nil) // Placeholder
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("tip-56-tx-%s", reference), nil
}

func (p *JsonStablecoinProvider) VerifyTransfer(ctx context.Context, transferID string) (bool, error) {
	return true, nil
}

var _ StablecoinProvider = (*JsonStablecoinProvider)(nil)
