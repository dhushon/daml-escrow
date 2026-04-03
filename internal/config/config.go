package config

import (
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
	Oracle struct {
		WebhookSecret string `mapstructure:"webhookSecret" yaml:"webhookSecret"`
	} `mapstructure:"oracle" yaml:"oracle"`
}

type AuthConfig struct {
	Issuer      string `mapstructure:"issuer" yaml:"issuer"`
	ClientID    string `mapstructure:"clientId" yaml:"clientId"`
	Audience    string `mapstructure:"audience" yaml:"audience"`
	Environment string `mapstructure:"environment" yaml:"environment"`
	AuthBypass  bool   `mapstructure:"authBypass" yaml:"authBypass"`
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

	return &cfg, nil
}
