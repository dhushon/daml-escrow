package services

import (
	"context"
	"database/sql"
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
	db     *sql.DB
}

func NewIdentityService(configPath string, db *sql.DB) (*IdentityService, error) {
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
		db:     db,
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
	// 1. Try to fetch from Postgres first (contains enriched metadata)
	var identity ledger.UserIdentity
	query := "SELECT okta_sub, daml_user_id, daml_party_id, email, display_name, role, title, affiliation, organization, physical_address, kyc_status FROM identities WHERE okta_sub = $1"
	
	var title, affiliation, organization, address, kyc sql.NullString
	err := s.db.QueryRow(query, oktaSub).Scan(
		&identity.OktaSub, &identity.DamlUserID, &identity.DamlPartyID, &identity.Email, &identity.DisplayName, &identity.Role,
		&title, &affiliation, &organization, &address, &kyc,
	)

	if err == nil {
		identity.Title = title.String
		identity.Affiliation = affiliation.String
		identity.Organization = organization.String
		identity.PhysicalAddress = address.String
		identity.KYCStatus = kyc.String
		return &identity, nil
	}

	// 2. If not in DB, fallback to Ledger discovery
	ledgerIdentity, err := ledgerClient.GetIdentity(ctx, oktaSub)
	if err != nil || ledgerIdentity == nil {
		// 3. If not on ledger, authoritatively determine role and provision
		role := "Depositor"
		emailLower := strings.ToLower(email)
		if strings.Contains(emailLower, "beneficiary") || strings.Contains(emailLower, "seller") || strings.Contains(emailLower, "pledgee") {
			role = "Beneficiary"
		} else if strings.Contains(emailLower, "mediator") || strings.Contains(emailLower, "banker") {
			role = "Mediator"
		}

		scopes := []string{"escrow:read", "escrow:write", "escrow:accept"}
		ledgerIdentity, err = ledgerClient.ProvisionUser(ctx, oktaSub, email, role, scopes)
		if err != nil {
			return nil, err
		}
	}

	// 4. Authoritatively sync Ledger identity to Postgres
	syncQuery := `
		INSERT INTO identities (okta_sub, daml_user_id, daml_party_id, email, display_name, role, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP)
		ON CONFLICT (okta_sub) DO UPDATE SET
			daml_user_id = $2, daml_party_id = $3, email = $4, display_name = $5, role = $6, updated_at = CURRENT_TIMESTAMP
	`
	_, _ = s.db.Exec(syncQuery, oktaSub, ledgerIdentity.DamlUserID, ledgerIdentity.DamlPartyID, ledgerIdentity.Email, ledgerIdentity.DisplayName, ledgerIdentity.Role)

	return ledgerIdentity, nil
}

func (s *IdentityService) ListEnrichedIdentities(ctx context.Context, ledgerClient ledger.Client) ([]*ledger.UserIdentity, error) {
	// JIT provision standard dev bypass users in development to support multi-node directory discovery
	if os.Getenv("ENVIRONMENT") == "dev" && os.Getenv("AUTH_BYPASS") == "true" {
		devEmails := []string{
			"joey@depositor.devlocal",
			"jimmy@beneficiary.devlocal",
			"sally@mediator.devlocal",
			"bob@banker.devlocal",
		}
		rawIdentities, err := ledgerClient.ListIdentities(ctx)
		if err == nil {
			for _, email := range devEmails {
				found := false
				for _, id := range rawIdentities {
					if strings.EqualFold(id.Email, email) {
						found = true
						break
					}
				}
				if !found {
					_, _ = s.GetOrCreateIdentity(ctx, email, email, ledgerClient)
				}
			}
		}
	}

	// 1. Get raw identities from ledger
	rawIdentities, err := ledgerClient.ListIdentities(ctx)
	if err != nil {
		return nil, err
	}

	// 2. Query Postgres directory
	query := "SELECT daml_user_id, email, display_name, role, title, affiliation, organization, physical_address, kyc_status FROM identities"
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		// Fallback to raw identities if DB query fails
		return rawIdentities, nil
	}
	defer func() { _ = rows.Close() }()

	dbMap := make(map[string]*ledger.UserIdentity)
	for rows.Next() {
		var id ledger.UserIdentity
		var title, affiliation, organization, address, kyc sql.NullString
		if err := rows.Scan(&id.DamlUserID, &id.Email, &id.DisplayName, &id.Role, &title, &affiliation, &organization, &address, &kyc); err == nil {
			id.Title = title.String
			id.Affiliation = affiliation.String
			id.Organization = organization.String
			id.PhysicalAddress = address.String
			id.KYCStatus = kyc.String
			dbMap[id.DamlUserID] = &id
		}
	}

	// 3. Decorate raw identities with T1 metadata
	for _, raw := range rawIdentities {
		if dbId, ok := dbMap[raw.DamlUserID]; ok {
			if dbId.DisplayName != "" {
				raw.DisplayName = dbId.DisplayName
			}
			if dbId.Email != "" {
				raw.Email = dbId.Email
			}
			if dbId.Role != "" {
				raw.Role = dbId.Role
			}
			raw.Title = dbId.Title
			raw.Affiliation = dbId.Affiliation
			raw.Organization = dbId.Organization
			raw.PhysicalAddress = dbId.PhysicalAddress
			raw.KYCStatus = dbId.KYCStatus
		}
	}

	return rawIdentities, nil
}

func (s *IdentityService) FindPartyIDByEmail(ctx context.Context, email string) (string, error) {
	var partyID string
	query := "SELECT daml_party_id FROM identities WHERE LOWER(email) = LOWER($1) LIMIT 1"
	err := s.db.QueryRowContext(ctx, query, email).Scan(&partyID)
	if err != nil {
		return "", err
	}
	return partyID, nil
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
