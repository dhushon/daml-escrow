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
		Host          string                     `mapstructure:"host" yaml:"host"`
		Port          int                        `mapstructure:"port" yaml:"port"`
		ParticipantID string                     `mapstructure:"participantId" yaml:"participantId"` // e.g. bank, buyer, seller
		Nodes         map[string]ParticipantNode `mapstructure:"nodes" yaml:"nodes"`
		Packages      struct {
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
	Region       string `mapstructure:"region" yaml:"region"`
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
	v := viper.GetViper()

	// Set Defaults
	v.SetDefault("server.port", 8081)
	v.SetDefault("auth.environment", "production")
	v.SetDefault("auth.authBypass", false)
	v.SetDefault("region", "us-central1")

	if path != "" {
		v.SetConfigFile(path)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// High-Assurance: Authoritatively bind environment variables AFTER reading config file
	// to ensure environment-level overrides take precedence in containerized workloads.
	_ = v.BindEnv("auth.environment", "ENVIRONMENT")
	_ = v.BindEnv("auth.authBypass", "AUTH_BYPASS")
	
	_ = v.BindEnv("ledger.host", "LEDGER_HOST")
	_ = v.BindEnv("ledger.port", "LEDGER_PORT")
	_ = v.BindEnv("ledger.participantId", "PARTICIPANT_ID")
	
	// High-Assurance: Bind node-specific host overrides for GKE tripartite nodes
	_ = v.BindEnv("ledger.nodes.bank.host", "LEDGER_NODES_BANK_HOST")
	_ = v.BindEnv("ledger.nodes.buyer.host", "LEDGER_NODES_BUYER_HOST")
	_ = v.BindEnv("ledger.nodes.seller.host", "LEDGER_NODES_SELLER_HOST")
	
	_ = v.BindEnv("userConfig.dsn", "USER_CONFIG_DSN")
	_ = v.BindEnv("gcpProjectId", "GCP_PROJECT_ID")
	_ = v.BindEnv("region", "REGION")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// ResolveSecrets authoritatively fetches sensitive values from GCP Secret Manager.
func ResolveSecrets(ctx context.Context, cfg *Config) error {
	if cfg.GCPProjectID == "" {
		return nil
	}

	resolver, err := NewSecretResolver(ctx, cfg.GCPProjectID)
	if err != nil {
		return fmt.Errorf("failed to initialize secret resolver: %w", err)
	}
	defer func() { _ = resolver.Close() }()

	if secret, err := resolver.GetSecret(ctx, "okta-client-secret-"+cfg.Auth.Environment); err == nil {
		cfg.Auth.ClientSecret = secret
	}
	if secret, err := resolver.GetSecret(ctx, "bitgo-access-token-"+cfg.Auth.Environment); err == nil {
		cfg.Stablecoin.BitGo.AccessToken = secret
	}
	if secret, err := resolver.GetSecret(ctx, "circle-api-key-"+cfg.Auth.Environment); err == nil {
		cfg.Stablecoin.Circle.APIKey = secret
	}
	if secret, err := resolver.GetSecret(ctx, "user-config-dsn-"+cfg.Auth.Environment); err == nil {
		cfg.UserConfig.DSN = secret
	}

	return nil
}
