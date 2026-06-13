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
		PartyDetails []struct {
			Party       string `json:"party"`
			DisplayName string `json:"displayName"`
		} `json:"partyDetails"`
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
	for _, p := range resp.PartyDetails {
		// High-Assurance: In V3, displayName might be empty, fallback to party handle
		name := p.DisplayName
		if name == "" {
			name = strings.Split(p.Party, "::")[0]
		}
		c.partyMap[name] = p.Party
	}

	return nil
}

func (c *JsonLedgerClient) GetParty(user string) string {
	if strings.Contains(user, "::") {
		return user
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	if id, ok := c.partyMap[user]; ok {
		return id
	}
	return user
}

func (c *JsonLedgerClient) ListIdentities(ctx context.Context) ([]*UserIdentity, error) {
	var resp struct {
		Users []struct {
			ID           string `json:"id"`
			PrimaryParty string `json:"primaryParty"`
			Metadata     struct {
				Annotations map[string]string `json:"annotations"`
			} `json:"metadata"`
		} `json:"users"`
	}

	body, err := c.DoRawRequest(ctx, "GET", "/v2/users", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch live institutional directory: %w", err)
	}
	_ = json.Unmarshal(body, &resp)

	identities := make([]*UserIdentity, 0, len(resp.Users))
	for _, u := range resp.Users {
		if u.ID == "participant_admin" {
			continue
		}
		
		email := u.Metadata.Annotations["email"]
		role := u.Metadata.Annotations["role"]
		if email == "" {
			switch u.ID {
			case "Depositor":
				email = "joey@depositor.devlocal"
			case "Beneficiary":
				email = "jimmy@beneficiary.devlocal"
			case "EscrowMediator":
				email = "sally@mediator.devlocal"
			case "CentralBank":
				email = "bob@banker.devlocal"
			default:
				email = strings.ToLower(u.ID) + "@devlocal"
			}
		}
		if role == "" {
			idLower := strings.ToLower(u.ID)
			if strings.Contains(idLower, "beneficiary") || strings.Contains(idLower, "seller") || strings.Contains(idLower, "pledgee") {
				role = "Beneficiary"
			} else if strings.Contains(idLower, "mediator") || strings.Contains(idLower, "banker") {
				role = "Mediator"
			} else {
				role = "Depositor"
			}
		}

		displayName := u.ID
		switch u.ID {
		case "Depositor", "u-joey-depositor-devlocal":
			displayName = "Joey"
		case "Beneficiary", "u-jimmy-beneficiary-devlocal":
			displayName = "Jimmy"
		case "EscrowMediator", "u-sally-mediator-devlocal":
			displayName = "Sally"
		case "CentralBank", "u-bob-banker-devlocal":
			displayName = "Bob"
		default:
			if strings.Contains(strings.ToLower(u.ID), "joey") {
				displayName = "Joey"
			} else if strings.Contains(strings.ToLower(u.ID), "jimmy") {
				displayName = "Jimmy"
			} else if strings.Contains(strings.ToLower(u.ID), "sally") {
				displayName = "Sally"
			} else if strings.Contains(strings.ToLower(u.ID), "bob") {
				displayName = "Bob"
			}
		}

		identities = append(identities, &UserIdentity{
			DamlUserID:  u.ID,
			DamlPartyID: u.PrimaryParty,
			Email:       email,
			Role:        role,
			DisplayName: displayName,
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
		User struct {
			ID           string `json:"id"`
			PrimaryParty string `json:"primaryParty"`
			Metadata     struct {
				Annotations map[string]string `json:"annotations"`
			} `json:"metadata"`
		} `json:"user"`
	}
	_ = json.Unmarshal(body, &resp)

	email := resp.User.Metadata.Annotations["email"]
	role := resp.User.Metadata.Annotations["role"]
	if email == "" {
		switch resp.User.ID {
		case "Depositor":
			email = "joey@depositor.devlocal"
		case "Beneficiary":
			email = "jimmy@beneficiary.devlocal"
		case "EscrowMediator":
			email = "sally@mediator.devlocal"
		case "CentralBank":
			email = "bob@banker.devlocal"
		default:
			email = strings.ToLower(resp.User.ID) + "@devlocal"
		}
	}
	if role == "" {
		idLower := strings.ToLower(resp.User.ID)
		if strings.Contains(idLower, "beneficiary") || strings.Contains(idLower, "seller") || strings.Contains(idLower, "pledgee") {
			role = "Beneficiary"
		} else if strings.Contains(idLower, "mediator") || strings.Contains(idLower, "banker") {
			role = "Mediator"
		} else {
			role = "Depositor"
		}
	}

	displayName := resp.User.ID
	switch resp.User.ID {
	case "Depositor", "u-joey-depositor-devlocal":
		displayName = "Joey"
	case "Beneficiary", "u-jimmy-beneficiary-devlocal":
		displayName = "Jimmy"
	case "EscrowMediator", "u-sally-mediator-devlocal":
		displayName = "Sally"
	case "CentralBank", "u-bob-banker-devlocal":
		displayName = "Bob"
	default:
		if strings.Contains(strings.ToLower(resp.User.ID), "joey") {
			displayName = "Joey"
		} else if strings.Contains(strings.ToLower(resp.User.ID), "jimmy") {
			displayName = "Jimmy"
		} else if strings.Contains(strings.ToLower(resp.User.ID), "sally") {
			displayName = "Sally"
		} else if strings.Contains(strings.ToLower(resp.User.ID), "bob") {
			displayName = "Bob"
		}
	}

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  damlUserID,
		DamlPartyID: resp.User.PrimaryParty,
		Email:       email,
		Role:        role,
		DisplayName: displayName,
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
		PartyDetails struct {
			Party string `json:"party"`
		} `json:"partyDetails"`
	}

	var partyID string
	body, err := c.DoRawRequest(ctx, "POST", "/v2/parties", partyReq)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			// High-Assurance: If party already exists, skip the slow wait and proceed
			partyID = damlUserID
		} else {
			return nil, fmt.Errorf("failed to allocate party: %w", err)
		}
	} else {
		_ = json.Unmarshal(body, &partyResp)
		partyID = partyResp.PartyDetails.Party
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

	// Faster, more authoritative retry loop for User creation
	var provisionErr error
	for i := 0; i < 25; i++ {
		_, err = c.DoRawRequest(ctx, "POST", "/v2/users", userReq)
		if err == nil {
			provisionErr = nil
			break
		}
		if strings.Contains(err.Error(), "409") {
			provisionErr = nil
			break
		}
		if strings.Contains(err.Error(), "UNKNOWN_RESOURCE") {
			c.logger.Warn("user service synchronization lag, retrying...", zap.Int("attempt", i+1))
			time.Sleep(500 * time.Millisecond) // Faster polling
			provisionErr = err
			continue
		}
		return nil, fmt.Errorf("failed to create user: %w", err)
	}


	if provisionErr != nil {
		return nil, fmt.Errorf("failed to create user after retries: %w", provisionErr)
	}

	displayName := damlUserID
	switch damlUserID {
	case "Depositor", "u-joey-depositor-devlocal":
		displayName = "Joey"
	case "Beneficiary", "u-jimmy-beneficiary-devlocal":
		displayName = "Jimmy"
	case "EscrowMediator", "u-sally-mediator-devlocal":
		displayName = "Sally"
	case "CentralBank", "u-bob-banker-devlocal":
		displayName = "Bob"
	default:
		if strings.Contains(strings.ToLower(damlUserID), "joey") {
			displayName = "Joey"
		} else if strings.Contains(strings.ToLower(damlUserID), "jimmy") {
			displayName = "Jimmy"
		} else if strings.Contains(strings.ToLower(damlUserID), "sally") {
			displayName = "Sally"
		} else if strings.Contains(strings.ToLower(damlUserID), "bob") {
			displayName = "Bob"
		}
	}

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  damlUserID,
		DamlPartyID: partyID,
		Email:       email,
		Role:        role,
		DisplayName: displayName,
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
