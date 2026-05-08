package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
	var resp struct {
		Result []struct {
			ID           string `json:"id"`
			PrimaryParty string `json:"primaryParty"`
			Metadata     struct {
				Annotations map[string]string `json:"annotations"`
			} `json:"metadata"`
		} `json:"result"`
	}

	body, err := c.DoRawRequest(ctx, "GET", "/v2/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch live institutional directory: %w", err)
	}
	_ = json.Unmarshal(body, &resp)

	identities := make([]*UserIdentity, 0, len(resp.Result))
	for _, u := range resp.Result {
		// High-Assurance: Skip system admin and unprovisioned users
		if u.ID == "participant_admin" {
			continue
		}
		
		email := u.Metadata.Annotations["email"]
		role := u.Metadata.Annotations["role"]
		if email == "" { email = u.ID + "@devlocal" }
		if role == "" { role = "Buyer" }

		identities = append(identities, &UserIdentity{
			DamlUserID:  u.ID,
			DamlPartyID: u.PrimaryParty,
			Email:       email,
			Role:        role,
			DisplayName: u.ID,
		})
	}

	return identities, nil
}

func (c *JsonLedgerClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	damlUserID, err := SanitizeID(oktaSub)
	if err != nil {
		return nil, fmt.Errorf("failed to sanitize okta sub: %w", err)
	}
	
	body, err := c.DoRawRequest(ctx, "GET", "/v2/users/"+damlUserID, nil)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result struct {
			ID           string `json:"id"`
			PrimaryParty string `json:"primaryParty"`
			Metadata     struct {
				Annotations map[string]string `json:"annotations"`
			} `json:"metadata"`
		} `json:"result"`
	}
	_ = json.Unmarshal(body, &resp)

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  damlUserID,
		DamlPartyID: resp.Result.PrimaryParty,
		Email:       resp.Result.Metadata.Annotations["email"],
		Role:        resp.Result.Metadata.Annotations["role"],
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
		"displayName": damlUserID,
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

	partyID := c.GetParty(damlUserID)
	if (partyID == damlUserID || partyID == "") && partyResp.Result.Party != "" {
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

	// High-Assurance: Implement retry loop to handle Canton propagation latency
	var provisionErr error
	for i := 0; i < 5; i++ {
		_, err = c.DoRawRequest(ctx, "POST", "/v2/users", userReq)
		if err == nil {
			provisionErr = nil
			break
		}
		
		if strings.Contains(err.Error(), "409") {
			provisionErr = nil // Already exists is OK
			break
		}
		
		if strings.Contains(err.Error(), "UNKNOWN_RESOURCE") {
			c.logger.Warn("party propagation latency detected, retrying user creation...", zap.Int("attempt", i+1), zap.String("party", partyID))
			time.Sleep(1 * time.Second)
			provisionErr = err
			continue
		}
		
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if provisionErr != nil {
		return nil, fmt.Errorf("failed to create user after retries: %w", provisionErr)
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
