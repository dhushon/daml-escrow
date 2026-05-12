package services

import (
	"context"
	"daml-escrow/internal/ledger"
	"fmt"
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

func (s *IdentityService) GetIdPConfigForEmail(email string) (AuthProvider, error) {
	parts := strings.Split(email, "@")
	if len(parts) < 2 {
		return AuthProvider{}, fmt.Errorf("invalid email format")
	}

	domain := parts[1]
	if provider, ok := s.config.Providers[domain]; ok {
		return provider, nil
	}

	return s.getDefaultProvider(), nil
}

func (s *IdentityService) GetOrCreateIdentity(ctx context.Context, oktaSub, email string, ledgerClient ledger.Client) (*ledger.UserIdentity, error) {
	// First, try to fetch existing identity
	identity, err := ledgerClient.GetIdentity(ctx, oktaSub)
	if err == nil && identity != nil {
		return identity, nil
	}

	// Authoritatively determine institutional role
	role := "Depositor" // Default fallback
	emailLower := strings.ToLower(email)
	if strings.Contains(emailLower, "beneficiary") || strings.Contains(emailLower, "seller") || strings.Contains(emailLower, "pledgee") {
		role = "Beneficiary"
	} else if strings.Contains(emailLower, "mediator") || strings.Contains(emailLower, "banker") {
		role = "Mediator"
	}

	// If not found, authoritatively provision the user
	scopes := []string{"escrow:read", "escrow:write", "escrow:accept"}
	return ledgerClient.ProvisionUser(ctx, oktaSub, email, role, scopes)
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
