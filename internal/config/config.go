package config

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port" yaml:"port"`
	} `mapstructure:"server" yaml:"server"`
	Ledger struct {
		Host     string                     `mapstructure:"host" yaml:"host"`
		Port     int                        `mapstructure:"port" yaml:"port"`
		Nodes    map[string]ParticipantNode `mapstructure:"nodes" yaml:"nodes"`
		Packages struct {
			Implementation string `mapstructure:"implementation" yaml:"implementation"`
			Interfaces     string `mapstructure:"interfaces" yaml:"interfaces"`
		} `mapstructure:"packages" yaml:"packages"`
		Parties struct {
			Issuer   string `mapstructure:"issuer" yaml:"issuer"`
			Buyer    string `mapstructure:"buyer" yaml:"buyer"`
			Seller   string `mapstructure:"seller" yaml:"seller"`
			Mediator string `mapstructure:"mediator" yaml:"mediator"`
		} `mapstructure:"parties" yaml:"parties"`
	} `mapstructure:"ledger" yaml:"ledger"`
	UserConfig struct {
		DSN string `mapstructure:"dsn" yaml:"dsn"`
	} `mapstructure:"userConfig" yaml:"userConfig"`
	Auth   AuthConfig `mapstructure:"auth" yaml:"auth"`
	Stablecoin struct {
		Provider string `mapstructure:"provider" yaml:"provider"` // mock, bitgo, circle
		BitGo    struct {
			ExpressURL  string `mapstructure:"expressUrl" yaml:"expressUrl"`
			AccessToken string `mapstructure:"accessToken" yaml:"accessToken"`
			Enterprise  string `mapstructure:"enterprise" yaml:"enterprise"`
			Coin        string `mapstructure:"coin" yaml:"coin"` // e.g. teth:usdc
		} `mapstructure:"bitgo" yaml:"bitgo"`
		Circle struct {
			BaseURL      string `mapstructure:"baseUrl" yaml:"baseUrl"`
			APIKey       string `mapstructure:"apiKey" yaml:"apiKey"`
			EntitySecret string `mapstructure:"entitySecret" yaml:"entitySecret"`
		} `mapstructure:"circle" yaml:"circle"`
	} `mapstructure:"stablecoin" yaml:"stablecoin"`
	Oracle struct {
		WebhookSecret string `mapstructure:"webhookSecret" yaml:"webhookSecret"`
	} `mapstructure:"oracle" yaml:"oracle"`
	GCPProjectID string `mapstructure:"gcpProjectId" yaml:"gcpProjectId"`
}

type AuthConfig struct {
	Issuer       string `mapstructure:"issuer" yaml:"issuer"`
	ClientID     string `mapstructure:"clientId" yaml:"clientId"`
	ClientSecret string `mapstructure:"clientSecret" yaml:"clientSecret"`
	Audience     string `mapstructure:"audience" yaml:"audience"`
	Environment  string `mapstructure:"environment" yaml:"environment"`
	AuthBypass   bool   `mapstructure:"authBypass" yaml:"authBypass"`
}

type ParticipantNode struct {
	Host string `mapstructure:"host" yaml:"host"`
	Port int    `mapstructure:"port" yaml:"port"`
}

func LoadConfig(path string) (*Config, error) {
	// Use global viper to allow external flag overrides
	v := viper.GetViper()

	// Set Defaults
	v.SetDefault("server.port", 8081)
	v.SetDefault("auth.environment", "production")
	v.SetDefault("auth.authBypass", false)

	// Configure file loading
	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	// Environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	// Explicitly bind top-level env vars to nested config keys
	_ = v.BindEnv("auth.environment", "ENVIRONMENT")
	_ = v.BindEnv("auth.authBypass", "AUTH_BYPASS")
	
	// Stablecoin Environment Variables
	_ = v.BindEnv("stablecoin.provider", "STABLECOIN_PROVIDER")
	_ = v.BindEnv("stablecoin.bitgo.expressUrl", "BITGO_EXPRESS_URL")
	_ = v.BindEnv("stablecoin.bitgo.accessToken", "BITGO_ACCESS_TOKEN")
	_ = v.BindEnv("stablecoin.bitgo.enterprise", "BITGO_ENTERPRISE")
	_ = v.BindEnv("stablecoin.bitgo.coin", "BITGO_COIN")

	// Circle Environment Variables
	_ = v.BindEnv("stablecoin.circle.baseUrl", "CIRCLE_BASE_URL")
	_ = v.BindEnv("stablecoin.circle.apiKey", "CIRCLE_API_KEY")
	_ = v.BindEnv("stablecoin.circle.entitySecret", "CIRCLE_ENTITY_SECRET")
	
	_ = v.BindEnv("gcpProjectId", "GCP_PROJECT_ID")

	// Load file if it exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// High-Assurance: Fetch secrets from GCP Secret Manager if configured
	if cfg.GCPProjectID != "" {
		ctx := context.Background()
		resolver, err := NewSecretResolver(ctx, cfg.GCPProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize secret resolver: %w", err)
		}
		defer func() { _ = resolver.Close() }()

		// Authoritatively override sensitive values from Cloud Secret Manager
		if secret, err := resolver.GetSecret(ctx, "okta-client-secret"); err == nil {
			cfg.Auth.ClientSecret = secret
		}
		if secret, err := resolver.GetSecret(ctx, "bitgo-access-token"); err == nil {
			cfg.Stablecoin.BitGo.AccessToken = secret
		}
		if secret, err := resolver.GetSecret(ctx, "circle-api-key"); err == nil {
			cfg.Stablecoin.Circle.APIKey = secret
		}
	}

	return &cfg, nil
}
