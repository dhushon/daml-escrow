package services

import (
	"context"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type AuthProvider struct {
	Type        string `json:"type" yaml:"type"`
	Issuer      string `json:"issuer" yaml:"issuer"`
	TenantID    string `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
	LoginURL    string `json:"loginUrl,omitempty" yaml:"loginUrl,omitempty"`
	Description string `json:"description" yaml:"description"`
}

type IdentityConfig struct {
	Providers map[string]AuthProvider `yaml:"providers"`
}

type IdentityService struct {
	config IdentityConfig
}

func NewIdentityService(configPath string) (*IdentityService, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config IdentityConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &IdentityService{
		config: config,
	}, nil
}

func (s *IdentityService) DiscoverProvider(ctx context.Context, email string) AuthProvider {
	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return s.getDefaultProvider()
	}

	domain := parts[1]
	if provider, ok := s.config.Providers[domain]; ok {
		return provider
	}

	return s.getDefaultProvider()
}

func (s *IdentityService) getDefaultProvider() AuthProvider {
	if provider, ok := s.config.Providers["default"]; ok {
		return provider
	}
	return AuthProvider{
		Type:        "OIDC",
		Issuer:      "https://accounts.google.com",
		Description: "Google Public Identity",
	}
}
