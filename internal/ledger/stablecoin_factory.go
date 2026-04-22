package ledger

import (
	"daml-escrow/internal/config"
	"fmt"

	"go.uber.org/zap"
)

// StablecoinFactory handles the dynamic resolution of institutional custody providers.
type StablecoinFactory struct {
	logger *zap.Logger
}

func NewStablecoinFactory(logger *zap.Logger) *StablecoinFactory {
	return &StablecoinFactory{logger: logger}
}

// CreateProvider initializes a provider based on the configured STABLECOIN_PROVIDER.
// This implements a dynamically pluggable pattern for additional CIP-0056 compliant providers.
func (f *StablecoinFactory) CreateProvider(cfg *config.Config, ledgerClient Client) (StablecoinProvider, error) {
	providerType := cfg.Stablecoin.Provider
	if providerType == "" {
		providerType = "mock"
	}

	f.logger.Info("initializing institutional stablecoin provider", zap.String("type", providerType))

	switch providerType {
	case "bitgo":
		if cfg.Stablecoin.BitGo.ExpressURL == "" || cfg.Stablecoin.BitGo.AccessToken == "" {
			return nil, fmt.Errorf("bitgo provider requested but configuration is incomplete (check BITGO_EXPRESS_URL and BITGO_ACCESS_TOKEN)")
		}
		return NewBitGoStablecoinProvider(
			f.logger,
			cfg.Stablecoin.BitGo.ExpressURL,
			cfg.Stablecoin.BitGo.AccessToken,
			cfg.Stablecoin.BitGo.Enterprise,
			cfg.Stablecoin.BitGo.Coin,
		), nil

	case "circle":
		if cfg.Stablecoin.Circle.APIKey == "" || cfg.Stablecoin.Circle.EntitySecret == "" {
			return nil, fmt.Errorf("circle provider requested but configuration is incomplete (check CIRCLE_API_KEY and CIRCLE_ENTITY_SECRET)")
		}
		return NewCircleStablecoinProvider(
			f.logger,
			cfg.Stablecoin.Circle.BaseURL,
			cfg.Stablecoin.Circle.APIKey,
			cfg.Stablecoin.Circle.EntitySecret,
		), nil

	case "mock":
		return NewJsonStablecoinProvider(f.logger, ledgerClient), nil

	default:
		return nil, fmt.Errorf("unsupported stablecoin provider: %s", providerType)
	}
}
