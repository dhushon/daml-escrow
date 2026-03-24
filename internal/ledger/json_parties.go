package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func (c *JsonLedgerClient) refreshPartyMap(ctx context.Context) error {
	respBody, err := c.doRawRequest(ctx, "GET", "/v2/parties", nil)
	if err != nil {
		return err
	}

	// Corrected response structure for SDK 3.4.x
	var response struct {
		PartyDetails []struct {
			Party       string `json:"party"`
			DisplayName string `json:"displayName"`
			IsLocal     bool   `json:"isLocal"`
		} `json:"partyDetails"`
	}
	
	if err := json.Unmarshal(respBody, &response); err != nil {
		return fmt.Errorf("failed to unmarshal party list: %w", err)
	}

	for _, p := range response.PartyDetails {
		if strings.HasPrefix(p.Party, "Buyer::") {
			c.partyMap[BuyerUser] = p.Party
		} else if strings.HasPrefix(p.Party, "CentralBank::") {
			c.partyMap[CentralBankUser] = p.Party
		} else if strings.HasPrefix(p.Party, "Seller::") {
			c.partyMap[SellerUser] = p.Party
		} else if strings.HasPrefix(p.Party, "EscrowMediator::") {
			c.partyMap[EscrowMediatorUser] = p.Party
		}
	}
	
	c.logger.Info("party map refreshed", zap.Any("mappings", c.partyMap))
	return nil
}

func (c *JsonLedgerClient) getParty(user string) string {
	if id, ok := c.partyMap[user]; ok {
		return id
	}
	return user
}

// sanitizeSub converts a Google sub (or any string) into a valid Daml User ID.
// Daml User ID regex: ^[a-z0-9][a-z0-9\-]{0,254}$
func sanitizeSub(sub string) string {
	clean := strings.ToLower(sub)
	clean = strings.ReplaceAll(clean, "|", "-")
	clean = strings.ReplaceAll(clean, "_", "-")
	clean = strings.ReplaceAll(clean, "@", "-")
	clean = strings.ReplaceAll(clean, ".", "-")
	
	// Prepend with a known string to ensure it starts with a letter/number
	return "u-" + clean
}

func (c *JsonLedgerClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	damlUserID := sanitizeSub(oktaSub)

	path := fmt.Sprintf("/v2/users/%s", damlUserID)
	respBody, err := c.doRawRequest(ctx, "GET", path, nil)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to fetch user identity: %w", err)
	}

	var response struct {
		User struct {
			Id           string `json:"id"`
			PrimaryParty string `json:"primaryParty"`
		} `json:"user"`
	}

	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user identity: %w", err)
	}

	// Update local map cache
	if response.User.PrimaryParty != "" {
		c.partyMap[damlUserID] = response.User.PrimaryParty
	}

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  response.User.Id,
		DamlPartyID: response.User.PrimaryParty,
	}, nil
}

func (c *JsonLedgerClient) ProvisionUser(ctx context.Context, oktaSub string, email string) (*UserIdentity, error) {
	damlUserID := sanitizeSub(oktaSub)

	// 1. Allocate Party
	partyBody := map[string]interface{}{
		"partyIdHint": damlUserID,
		"displayName": email,
	}
	partyResp, err := c.doRawRequest(ctx, "POST", "/v2/parties", partyBody)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate party: %w", err)
	}

	var partyResponse struct {
		PartyDetails struct {
			Party string `json:"party"`
		} `json:"partyDetails"`
	}
	if err := json.Unmarshal(partyResp, &partyResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal party allocation: %w", err)
	}
	allocatedParty := partyResponse.PartyDetails.Party

	// 2. Create User
	userBody := map[string]interface{}{
		"user": map[string]interface{}{
			"id":                 damlUserID,
			"primaryParty":       allocatedParty,
			"isDeactivated":      false,
			"identityProviderId": "",
		},
	}
	_, err = c.doRawRequest(ctx, "POST", "/v2/users", userBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create daml user: %w", err)
	}

	// 3. Grant actAs Rights
	rightsBody := map[string]interface{}{
		"userId":             damlUserID,
		"identityProviderId": "",
		"grant": []map[string]interface{}{
			{
				"type":  "actAs",
				"party": allocatedParty,
			},
		},
	}
	rightsPath := fmt.Sprintf("/v2/users/%s/rights", damlUserID)
	_, err = c.doRawRequest(ctx, "POST", rightsPath, rightsBody)
	if err != nil {
		return nil, fmt.Errorf("failed to grant rights: %w", err)
	}

	// Update local map cache
	c.partyMap[damlUserID] = allocatedParty

	c.logger.Info("provisioned new identity", zap.String("oktaSub", oktaSub), zap.String("damlParty", allocatedParty))

	return &UserIdentity{
		OktaSub:     oktaSub,
		DamlUserID:  damlUserID,
		DamlPartyID: allocatedParty,
		Email:       email,
	}, nil
}
