package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

// refreshPartyMap authoritatively synchronizes the local party cache with the Canton ledger.
func (c *JsonLedgerClient) refreshPartyMap(ctx context.Context) error {
	var resp struct {
		Result []struct {
			Party       string `json:"party"`
			DisplayName string `json:"displayName"`
		} `json:"result"`
	}

	body, err := c.DoRawRequest(ctx, "GET", "/v2/parties", nil)
	if err != nil {
		return fmt.Errorf("failed to list parties: %w", err)
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return fmt.Errorf("failed to unmarshal party list: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, p := range resp.Result {
		c.partyMap[p.DisplayName] = p.Party
	}

	c.logger.Info("party map refreshed", zap.Int("totalParties", len(c.partyMap)))
	return nil
}

// GetParty authoritatively resolves a logical User ID to a Canton Party ID.
func (c *JsonLedgerClient) GetParty(user string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if id, ok := c.partyMap[user]; ok {
		return id
	}
	return user // Fallback to raw string if not mapped
}

func (c *JsonLedgerClient) ListIdentities(ctx context.Context) ([]*UserIdentity, error) {
	// For Phase 9/11, we return the high-assurance tripartite identities
	// In production, this would be a lookup against the User Management API metadata.
	return []*UserIdentity{
		{OktaSub: "u-bank", DamlUserID: "bank", Email: "bob@bank.com", Role: "Mediator"},
		{OktaSub: "u-buyer", DamlUserID: "buyer", Email: "joey@buyer.com", Role: "Buyer"},
		{OktaSub: "u-seller", DamlUserID: "seller", Email: "jimmy@seller.com", Role: "Seller"},
	}, nil
}

func (c *JsonLedgerClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	damlUserID := "u-" + strings.ReplaceAll(oktaSub, "|", "-")
	
	// Verification check
	_, err := c.DoRawRequest(ctx, "GET", "/v2/users/"+damlUserID, nil)
	if err != nil {
		return nil, err
	}

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  damlUserID,
		DamlPartyID: c.GetParty(damlUserID),
		Role:        "Buyer", // Placeholder: In prod, fetch from metadata
	}, nil
}

func (c *JsonLedgerClient) ProvisionUser(ctx context.Context, oktaSub string, email string, role string, scopes []string) (*UserIdentity, error) {
	damlUserID := "u-" + strings.ReplaceAll(oktaSub, "|", "-")
	
	// 1. Allocate a new Party for the user
	partyReq := map[string]interface{}{
		"identifierHint": damlUserID,
		"displayName":    email,
	}
	
	var partyResp struct {
		Result struct {
			Party string `json:"party"`
		} `json:"result"`
	}

	body, err := c.DoRawRequest(ctx, "POST", "/v2/parties", partyReq)
	if err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return nil, fmt.Errorf("failed to allocate party: %w", err)
		}
		// Try to refresh to pick up the existing party
		_ = c.refreshPartyMap(ctx)
	} else {
		_ = json.Unmarshal(body, &partyResp)
	}

	partyID := c.GetParty(email)
	if partyID == email && partyResp.Result.Party != "" {
		partyID = partyResp.Result.Party
	}

	// 2. Create the User and link to the Party
	userReq := map[string]interface{}{
		"userId":       damlUserID,
		"primaryParty": partyID,
		"rights": []map[string]interface{}{
			{"type": "CanActAs", "party": partyID},
			{"type": "CanReadAs", "party": partyID},
		},
		"metadata": map[string]string{
			"email": email,
			"role":  role,
		},
	}

	_, err = c.DoRawRequest(ctx, "POST", "/v2/users", userReq)
	if err != nil {
		if !strings.Contains(err.Error(), "409") {
			return nil, fmt.Errorf("failed to create user: %w", err)
		}
	}

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  damlUserID,
		DamlPartyID: partyID,
		Email:       email,
		Role:        role,
	}, nil
}

func (c *JsonLedgerClient) ListParties(ctx context.Context) ([]*Wallet, error) {
	var resp struct {
		Result []struct {
			Party       string `json:"party"`
			DisplayName string `json:"displayName"`
		} `json:"result"`
	}

	body, err := c.DoRawRequest(ctx, "GET", "/v2/parties", nil)
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(body, &resp)

	wallets := make([]*Wallet, 0, len(resp.Result))
	for _, p := range resp.Result {
		wallets = append(wallets, &Wallet{
			ID:    p.Party,
			Owner: p.DisplayName,
		})
	}
	return wallets, nil
}
