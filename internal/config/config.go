package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Ledger struct {
		Host    string `yaml:"host"`
		Port    int    `yaml:"port"`
		Nodes   map[string]ParticipantNode `yaml:"nodes"`
		Packages struct {
			Implementation string `yaml:"implementation"`
			Interfaces     string `yaml:"interfaces"`
		} `yaml:"packages"`
		Parties struct {
			Issuer   string `yaml:"issuer"`
			Buyer    string `yaml:"buyer"`
			Seller   string `yaml:"seller"`
			Mediator string `yaml:"mediator"`
		} `yaml:"parties"`
	} `yaml:"ledger"`
	UserConfig struct {
		DSN string `yaml:"dsn"`
	} `yaml:"userConfig"`
	Auth struct {
		Issuer   string `yaml:"issuer"`
		ClientID string `yaml:"clientId"`
		Audience string `yaml:"audience"`
	} `yaml:"auth"`
	Oracle struct {
		WebhookSecret string `yaml:"webhookSecret"`
	} `yaml:"oracle"`
}

type ParticipantNode struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
