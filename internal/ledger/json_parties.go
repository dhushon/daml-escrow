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

func (c *JsonLedgerClient) GetParty(user string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if id, ok := c.partyMap[user]; ok {
		return id
	}
	return user
}

func (c *JsonLedgerClient) ListIdentities(ctx context.Context) ([]*UserIdentity, error) {
	return []*UserIdentity{
		{OktaSub: "u-bank", DamlUserID: "bank", Email: "bob@bank.devlocal", Role: "Mediator"},
		{OktaSub: "u-buyer", DamlUserID: "buyer", Email: "joey@buyer.devlocal", Role: "Buyer"},
		{OktaSub: "u-seller", DamlUserID: "seller", Email: "jimmy@seller.devlocal", Role: "Seller"},
	}, nil
}

func (c *JsonLedgerClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	damlUserID, err := SanitizeID(oktaSub)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize okta sub: %w", err)
	}
	
	_, err = c.DoRawRequest(ctx, "GET", "/v2/users/"+damlUserID, nil)
	if err != nil {
		return nil, err
	}

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  damlUserID,
		DamlPartyID: c.GetParty(damlUserID),
		Role:        "Buyer", 
	}, nil
}

func (c *JsonLedgerClient) ProvisionUser(ctx context.Context, oktaSub string, email string, role string, scopes []string) (*UserIdentity, error) {
	damlUserID, err := SanitizeID(oktaSub)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize user identity: %w", err)
	}
	
	// 1. Allocate a new Party for the user
	partyReq := map[string]interface{}{
		"partyIdHint": damlUserID,
		"displayName": damlUserID, // authoritatively compliant
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
		_ = c.refreshPartyMap(ctx)
	} else {
		_ = json.Unmarshal(body, &partyResp)
	}

	partyID := c.GetParty(email)
	if (partyID == email || partyID == "") && partyResp.Result.Party != "" {
		partyID = partyResp.Result.Party
	}

	// 2. Create the User and link to the Party
	userReq := map[string]interface{}{
		"user": map[string]interface{}{
			"id":            damlUserID,
			"primaryParty":  partyID,
			"isDeactivated": false,
			"identityProviderId": "",
			"metadata": map[string]interface{}{
				"resourceVersion": "",
				"annotations": map[string]string{
					"email": email,
					"role":  role,
				},
			},
		},
		"rights": []map[string]interface{}{
			{
				"kind": map[string]interface{}{
					"CanActAs": map[string]interface{}{
						"value": map[string]string{"party": partyID},
					},
				},
			},
			{
				"kind": map[string]interface{}{
					"CanReadAs": map[string]interface{}{
						"value": map[string]string{"party": partyID},
					},
				},
			},
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
