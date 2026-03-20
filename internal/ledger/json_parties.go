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
