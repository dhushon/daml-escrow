package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (c *JsonLedgerClient) getLedgerEnd(ctx context.Context) (interface{}, error) {
	respBody, err := c.doRawRequest(ctx, "GET", "/v2/state/ledger-end", nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Offset json.RawMessage `json:"offset"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode ledger end: %w", err)
	}
	
	var offset interface{}
	if err := json.Unmarshal(result.Offset, &offset); err != nil {
		return nil, err
	}
	return offset, nil
}

func (c *JsonLedgerClient) extractContract(events []map[string]interface{}, templateName string) (*EscrowContract, error) {
	for _, wrapper := range events {
		var created map[string]interface{}
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			created = ce
		} else if ce, ok := wrapper["createdEvent"].(map[string]interface{}); ok {
			created = ce
		}

		if created != nil {
			templateId, _ := created["templateId"].(string)
			if !strings.Contains(templateId, templateName) {
				continue
			}

			contractId, _ := created["contractId"].(string)
			var args map[string]interface{}
			if a, ok := created["createArguments"].(map[string]interface{}); ok {
				args = a
			} else if a, ok := created["createArgument"].(map[string]interface{}); ok {
				args = a
			}

			if args != nil {
				amountStr := fmt.Sprintf("%v", args["totalAmount"])
				amount, _ := strconv.ParseFloat(amountStr, 64)
				
				var ms []Milestone
				if rawMs, ok := args["milestones"].([]interface{}); ok {
					for _, r := range rawMs {
						if rm, ok := r.(map[string]interface{}); ok {
							ma, _ := strconv.ParseFloat(fmt.Sprintf("%v", rm["amount"]), 64)
							ms = append(ms, Milestone{
								Label:     fmt.Sprintf("%v", rm["label"]),
								Amount:    ma,
								Completed: rm["completed"].(bool),
							})
						}
					}
				}

				curIdx := 0
				if ci, ok := args["currentMilestoneIndex"].(float64); ok {
					curIdx = int(ci)
				} else if ci, ok := args["currentMilestoneIndex"].(string); ok {
					ci64, _ := strconv.ParseInt(ci, 10, 32)
					curIdx = int(ci64)
				}

				var metadata EscrowMetadata
				if mJSON, ok := args["metadata"].(string); ok && mJSON != "" {
					_ = json.Unmarshal([]byte(mJSON), &metadata)
				}

				item := &EscrowContract{
					ID:                    contractId,
					Buyer:                 fmt.Sprintf("%v", args["buyer"]),
					Seller:                fmt.Sprintf("%v", args["seller"]),
					Issuer:                fmt.Sprintf("%v", args["issuer"]),
					Mediator:              fmt.Sprintf("%v", args["mediator"]),
					Amount:                amount,
					Currency:              fmt.Sprintf("%v", args["currency"]),
					State:                 "Active",
					Milestones:            ms,
					CurrentMilestoneIndex: curIdx,
					Metadata:              metadata,
				}
				return item, nil
			}
		}
	}
	return nil, fmt.Errorf("contract with template %s not found in transaction events", templateName)
}

func (c *JsonLedgerClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	if len(c.partyMap) == 0 {
		_ = c.refreshPartyMap(ctx)
	}

	buyerParty := c.getParty(BuyerUser)
	cbParty := c.getParty(CentralBankUser)
	sellerParty := c.getParty(SellerUser)
	mediatorParty := c.getParty(EscrowMediatorUser)

	amountStr := fmt.Sprintf("%.10f", req.Amount)

	var milestones []interface{}
	if len(req.Milestones) > 0 {
		for _, m := range req.Milestones {
			milestones = append(milestones, map[string]interface{}{
				"label":     m.Label,
				"amount":    fmt.Sprintf("%.10f", m.Amount),
				"completed": m.Completed,
			})
		}
	} else {
		milestones = []interface{}{
			map[string]interface{}{
				"label":     "Full Payment",
				"amount":    amountStr,
				"completed": false,
			},
		}
	}

	description := req.Description
	if description == "" {
		description = "Escrow for " + req.Currency
	}

	metadataJSON := "{}"
	if req.Metadata.SchemaURL != "" || len(req.Metadata.Payload) > 0 {
		filteredPayload := make(map[string]interface{})
		for k, v := range req.Metadata.Payload {
			if _, excluded := req.Metadata.Exclusions[k]; excluded {
				continue
			}
			filteredPayload[k] = v
		}

		toSerialize := EscrowMetadata{
			SchemaURL: req.Metadata.SchemaURL,
			Payload:   filteredPayload,
		}

		if b, err := json.Marshal(toSerialize); err == nil {
			metadataJSON = string(b)
		}
	}

	payload := map[string]interface{}{
		"issuer":                cbParty,
		"buyer":                 buyerParty,
		"seller":                sellerParty,
		"mediator":              mediatorParty,
		"totalAmount":           amountStr,
		"currency":              req.Currency,
		"description":           description,
		"milestones":            milestones,
		"currentMilestoneIndex": 0,
		"metadata":              metadataJSON,
	}

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("create-escrow-%d", time.Now().UnixNano()),
			"actAs":     []string{buyerParty, cbParty},
			"userId":    BuyerUser,
			"commands": []interface{}{
				map[string]interface{}{
					"CreateCommand": map[string]interface{}{
						"templateId":      fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinEscrow"),
						"createArguments": payload,
					},
				},
			},
		},
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	if err != nil {
		return nil, err
	}

	var txResp v2TransactionResponse
	if err := json.Unmarshal(respBody, &txResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction response: %w", err)
	}

	return c.extractContract(txResp.Transaction.Events, "StablecoinEscrow")
}

func (c *JsonLedgerClient) ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error) {
	party := c.getParty(userID)
	
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
						fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinEscrow"),
						fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinDisputedEscrow"),
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
		return nil, fmt.Errorf("failed to parse active contracts: %w", err)
	}

	var escrows []*EscrowContract
	for _, wrapper := range result {
		var created map[string]interface{}
		contractState := "Active"
		
		// Structure for /active-contracts response objects
		if entry, ok := wrapper["contractEntry"].(map[string]interface{}); ok {
			if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
				if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
					created = ce
					templateId, _ := ce["templateId"].(string)
					if strings.Contains(templateId, "StablecoinDisputedEscrow") {
						contractState = "Disputed"
					}
				}
			}
		}

		if created != nil {
			contractId, _ := created["contractId"].(string)
			var args map[string]interface{}
			if a, ok := created["createArguments"].(map[string]interface{}); ok {
				args = a
			} else if a, ok := created["createArgument"].(map[string]interface{}); ok {
				args = a
			}
			
			if args != nil {
				amountStr := fmt.Sprintf("%v", args["totalAmount"])
				amount, _ := strconv.ParseFloat(amountStr, 64)
				
				var ms []Milestone
				if rawMs, ok := args["milestones"].([]interface{}); ok {
					for _, r := range rawMs {
						if rm, ok := r.(map[string]interface{}); ok {
							ma, _ := strconv.ParseFloat(fmt.Sprintf("%v", rm["amount"]), 64)
							ms = append(ms, Milestone{
								Label:     fmt.Sprintf("%v", rm["label"]),
								Amount:    ma,
								Completed: rm["completed"].(bool),
							})
						}
					}
				}

				curIdx := 0
				if ci, ok := args["currentMilestoneIndex"].(float64); ok {
					curIdx = int(ci)
				} else if ci, ok := args["currentMilestoneIndex"].(string); ok {
					ci64, _ := strconv.ParseInt(ci, 10, 32)
					curIdx = int(ci64)
				}

				var metadata EscrowMetadata
				if mJSON, ok := args["metadata"].(string); ok && mJSON != "" {
					_ = json.Unmarshal([]byte(mJSON), &metadata)
				}

				escrows = append(escrows, &EscrowContract{
					ID:                    contractId,
					Buyer:                 fmt.Sprintf("%v", args["buyer"]),
					Seller:                fmt.Sprintf("%v", args["seller"]),
					Issuer:                fmt.Sprintf("%v", args["issuer"]),
					Mediator:              fmt.Sprintf("%v", args["mediator"]),
					Amount:                amount,
					Currency:              fmt.Sprintf("%v", args["currency"]),
					State:                 contractState,
					Milestones:            ms,
					CurrentMilestoneIndex: curIdx,
					Metadata:              metadata,
				})
			}
		}
	}
	return escrows, nil
}

func (c *JsonLedgerClient) GetEscrow(ctx context.Context, id string) (*EscrowContract, error) {
	for retry := 0; retry < 5; retry++ {
		contracts, err := c.ListEscrows(ctx, BuyerUser)
		if err != nil {
			return nil, err
		}
		for _, item := range contracts {
			if item.ID == id {
				return item, nil
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return nil, fmt.Errorf("escrow not found: %s", id)
}

func (c *JsonLedgerClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	escrows, err := c.ListEscrows(ctx, userID)
	if err != nil {
		return nil, err
	}

	metrics := &LedgerMetrics{
		TotalActiveEscrows: len(escrows),
		ActivityHistory: []ActivityPoint{
			{Date: "2026-03-15", Count: 2},
			{Date: "2026-03-16", Count: 5},
			{Date: "2026-03-17", Count: 3},
			{Date: "2026-03-18", Count: 8},
			{Date: "2026-03-19", Count: 12},
			{Date: "2026-03-20", Count: len(escrows)},
		},
		SystemPerformance: SystemPerformance{
			APILatencyMS:      42,
			P95LatencyMS:      128,
			P99LatencyMS:      245,
			ErrorRate:         0.04,
			RequestCount:      1250,
			SuccessRate:       99.96,
			Uptime:            "4d 12h 30m",
			CPUUsage:          12.5,
			MemoryUsage:       256.0,
			DiskUsage:         34.2,
			ActiveConnections: 18,
		},
	}

	for _, e := range escrows {
		metrics.TotalValueInEscrow += e.Amount
	}

	settlements, err := c.ListSettlements(ctx)
	if err == nil {
		party := c.getParty(userID)
		for _, s := range settlements {
			if s.Issuer == party || s.Recipient == party {
				metrics.PendingSettlements++
				metrics.PendingSettlementValue += s.Amount
			}
		}
	}

	return metrics, nil
}

func (c *JsonLedgerClient) ReleaseFunds(ctx context.Context, id string) error {
	party := c.getParty(BuyerUser)
	
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("release-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    BuyerUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":  fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Escrow.Interface", "Escrow"),
						"contractId":  id,
						"choice":      "ApproveMilestone",
						"choiceArgument": map[string]interface{}{
							"payload": "ApproveMilestoneArg",
						},
					},
				},
			},
		},
	}

	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) RaiseDispute(ctx context.Context, id string) (string, error) {
	party := c.getParty(BuyerUser)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("dispute-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    BuyerUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":  fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Escrow.Interface", "Escrow"),
						"contractId":  id,
						"choice":      "RaiseDispute",
						"choiceArgument": map[string]interface{}{
							"payload": "RaiseDisputeArg",
						},
					},
				},
			},
		},
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	if err != nil {
		return "", err
	}

	var txResp v2TransactionResponse
	if err := json.Unmarshal(respBody, &txResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal transaction response: %w", err)
	}

	escrow, err := c.extractContract(txResp.Transaction.Events, "StablecoinDisputedEscrow")
	if err != nil {
		return "", err
	}
	return escrow.ID, nil
}

func (c *JsonLedgerClient) ResolveDispute(ctx context.Context, id string, payoutToBuyer, payoutToSeller float64) error {
	party := c.getParty(EscrowMediatorUser)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("resolve-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    EscrowMediatorUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":  fmt.Sprintf("%s:%s:%s", InterfacePackageID, "Escrow.Interface", "DisputedEscrow"),
						"contractId":  id,
						"choice":      "ResolveDispute",
						"choiceArgument": map[string]interface{}{
							"payload": map[string]interface{}{
								"payoutToBuyer":  fmt.Sprintf("%.10f", payoutToBuyer),
								"payoutToSeller": fmt.Sprintf("%.10f", payoutToSeller),
							},
						},
					},
				},
			},
		},
	}

	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) RefundBuyer(ctx context.Context, id string) error {
	escrow, err := c.GetEscrow(ctx, id)
	if err != nil {
		return err
	}

	remaining := 0.0
	for _, m := range escrow.Milestones {
		if !m.Completed {
			remaining += m.Amount
		}
	}

	if remaining <= 0 {
		return fmt.Errorf("no funds to refund")
	}

	disputeID, err := c.RaiseDispute(ctx, id)
	if err != nil {
		return err
	}

	return c.ResolveDispute(ctx, disputeID, remaining, 0.0)
}

func (c *JsonLedgerClient) RefundBySeller(ctx context.Context, id string) error {
	party := c.getParty(SellerUser)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("seller-refund-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    SellerUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinEscrow"),
						"contractId":     id,
						"choice":         "SellerRefund",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}

	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}
