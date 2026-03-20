package ledger

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

func (c *JsonLedgerClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	party := c.getParty(CentralBankUser)
	offset, err := c.getLedgerEnd(ctx)
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"activeAtOffset": offset,
		"eventFormat": map[string]interface{}{
			"verbose": true,
			"filtersByParty": map[string]interface{}{
				party: map[string]interface{}{
					"templateIds": []string{
						fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "EscrowSettlement"),
					},
				},
			},
		},
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/state/active-contracts", body)
	if err != nil {
		return nil, err
	}

	result, err := parseNDJSON(respBody)
	if err != nil {
		return nil, err
	}
	
	var list []*EscrowSettlement
	for _, item := range result {
		var created map[string]interface{}
		if entry, ok := item["contractEntry"].(map[string]interface{}); ok {
			if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
				if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
					created = ce
				}
			}
		}

		if created != nil {
			var args map[string]interface{}
			if a, ok := created["createArguments"].(map[string]interface{}); ok {
				args = a
			} else if a, ok := created["createArgument"].(map[string]interface{}); ok {
				args = a
			}

			if args != nil {
				amt, _ := strconv.ParseFloat(fmt.Sprintf("%v", args["amount"]), 64)
				list = append(list, &EscrowSettlement{
					ID:        fmt.Sprintf("%v", created["contractId"]),
					Issuer:    fmt.Sprintf("%v", args["issuer"]),
					Recipient: fmt.Sprintf("%v", args["recipient"]),
					Amount:    amt,
					Currency:  fmt.Sprintf("%v", args["currency"]),
					Status:    fmt.Sprintf("%v", args["status"]),
				})
			}
		}
	}
	return list, nil
}

func (c *JsonLedgerClient) SettlePayment(ctx context.Context, settlementID string) error {
	party := c.getParty(CentralBankUser)
	
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("settle-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    CentralBankUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":  fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Escrow.Interface", "Settlement"),
						"contractId":  settlementID,
						"choice":      "Settle",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}

	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}
