package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (c *JsonLedgerClient) CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, terms EscrowTerms) (*EscrowInvitation, error) {
	inviterParty := c.getParty(inviterID)
	mediatorParty := c.getParty(EscrowMediatorUser)
	cbParty := c.getParty(CentralBankUser)

	tokenHash := fmt.Sprintf("invite-%d", time.Now().UnixNano())

	var milestones []interface{}
	for _, m := range terms.Milestones {
		milestones = append(milestones, map[string]interface{}{
			"label":     m.Label,
			"amount":    fmt.Sprintf("%.10f", m.Amount),
			"completed": m.Completed,
		})
	}

	payload := map[string]interface{}{
		"inviter":      inviterParty,
		"mediator":     mediatorParty,
		"issuer":       cbParty,
		"inviteeEmail": inviteeEmail,
		"inviteeRole":  role,
		"inviteeType":  inviteeType,
		"tokenHash":    tokenHash,
		"terms": map[string]interface{}{
			"totalAmount": fmt.Sprintf("%.10f", terms.TotalAmount),
			"currency":    terms.Currency,
			"description": terms.Description,
			"milestones":  milestones,
			"metadata":    terms.Metadata,
		},
	}

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("create-invite-%d", time.Now().UnixNano()),
			"actAs":     []string{inviterParty, mediatorParty},
			"userId":    inviterID,
			"commands": []interface{}{
				map[string]interface{}{
					"CreateCommand": map[string]interface{}{
						"templateId":      fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowInvitation"),
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

	for _, wrapper := range txResp.Transaction.Events {
		var created map[string]interface{}
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			created = ce
		}

		if created != nil && strings.Contains(created["templateId"].(string), "EscrowInvitation") {
			args := created["createArgument"].(map[string]interface{})
			return &EscrowInvitation{
				ID:           created["contractId"].(string),
				Inviter:      fmt.Sprintf("%v", args["inviter"]),
				Mediator:     fmt.Sprintf("%v", args["mediator"]),
				Issuer:       fmt.Sprintf("%v", args["issuer"]),
				InviteeEmail: fmt.Sprintf("%v", args["inviteeEmail"]),
				InviteeRole:  fmt.Sprintf("%v", args["inviteeRole"]),
				InviteeType:  fmt.Sprintf("%v", args["inviteeType"]),
				TokenHash:    fmt.Sprintf("%v", args["tokenHash"]),
			}, nil
		}
	}

	return nil, fmt.Errorf("invitation created but not found in response")
}

func (c *JsonLedgerClient) ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*EscrowProposal, error) {
	claimantParty := c.getParty(claimantID)
	
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("claim-%d", time.Now().UnixNano()),
			"actAs":     []string{claimantParty},
			"userId":    claimantID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowInvitation"),
						"contractId":     inviteID,
						"choice":         "Claim",
						"choiceArgument": map[string]interface{}{
							"claimingParty": claimantParty,
						},
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

	for _, wrapper := range txResp.Transaction.Events {
		var created map[string]interface{}
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			created = ce
		}
		if created != nil && strings.Contains(created["templateId"].(string), "EscrowProposal") {
			args := created["createArgument"].(map[string]interface{})
			amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", args["totalAmount"]), 64)
			return &EscrowProposal{
				ID:          created["contractId"].(string),
				Buyer:       fmt.Sprintf("%v", args["buyer"]),
				Seller:      fmt.Sprintf("%v", args["seller"]),
				Issuer:      fmt.Sprintf("%v", args["issuer"]),
				Mediator:    fmt.Sprintf("%v", args["mediator"]),
				Amount:      amount,
				Currency:    fmt.Sprintf("%v", args["currency"]),
				Description: fmt.Sprintf("%v", args["description"]),
			}, nil
		}
	}

	return nil, fmt.Errorf("invitation claimed but proposal not found in response")
}

func (c *JsonLedgerClient) ListInvitations(ctx context.Context, userID string) ([]*EscrowInvitation, error) {
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
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowInvitation"),
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

	var list = []*EscrowInvitation{}
	for _, wrapper := range result {
		var created map[string]interface{}
		if entry, ok := wrapper["contractEntry"].(map[string]interface{}); ok {
			if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
				if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
					created = ce
				}
			}
		}

		if created != nil {
			args, ok := created["createArgument"].(map[string]interface{})
			if ok {
				list = append(list, &EscrowInvitation{
					ID:           created["contractId"].(string),
					Inviter:      fmt.Sprintf("%v", args["inviter"]),
					Mediator:     fmt.Sprintf("%v", args["mediator"]),
					Issuer:       fmt.Sprintf("%v", args["issuer"]),
					InviteeEmail: fmt.Sprintf("%v", args["inviteeEmail"]),
					TokenHash:    fmt.Sprintf("%v", args["tokenHash"]),
				})
			}
		}
	}
	return list, nil
}

func (c *JsonLedgerClient) GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error) {
	for retry := 0; retry < 10; retry++ {
		offset, err := c.getLedgerEnd(ctx)
		if err != nil {
			return nil, err
		}

		body := map[string]interface{}{
			"activeAtOffset": offset,
			"eventFormat": map[string]interface{}{
				"verbose": true,
				"filtersForAnyParty": map[string]interface{}{
					"templateIds": []string{
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowInvitation"),
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

		for _, wrapper := range result {
			var created map[string]interface{}
			if entry, ok := wrapper["contractEntry"].(map[string]interface{}); ok {
				if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
					if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
						created = ce
					}
				}
			}

			if created != nil {
				args, ok := created["createArgument"].(map[string]interface{})
				if ok && fmt.Sprintf("%v", args["tokenHash"]) == tokenHash {
					return &EscrowInvitation{
						ID:           created["contractId"].(string),
						Inviter:      fmt.Sprintf("%v", args["inviter"]),
						Mediator:     fmt.Sprintf("%v", args["mediator"]),
						Issuer:       fmt.Sprintf("%v", args["issuer"]),
						InviteeEmail: fmt.Sprintf("%v", args["inviteeEmail"]),
						InviteeRole:  fmt.Sprintf("%v", args["inviteeRole"]),
						InviteeType:  fmt.Sprintf("%v", args["inviteeType"]),
						TokenHash:    tokenHash,
					}, nil
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	return nil, fmt.Errorf("invitation not found for token")
}

func (c *JsonLedgerClient) GetEscrow(ctx context.Context, id string, userID string) (*EscrowContract, error) {
	for retry := 0; retry < 15; retry++ {
		contracts, err := c.ListEscrows(ctx, userID)
		if err != nil {
			return nil, err
		}
		for _, item := range contracts {
			if item.ID == id {
				return item, nil
			}
		}
		time.Sleep(2 * time.Second)
	}
	return nil, fmt.Errorf("escrow not found: %s", id)
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
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "StablecoinEscrow"),
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "StablecoinDisputedEscrow"),
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

	var escrows = []*EscrowContract{}
	for _, wrapper := range result {
		var created map[string]interface{}
		contractState := "Active"
		
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
			args, ok := created["createArgument"].(map[string]interface{})
			if ok {
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

func (c *JsonLedgerClient) ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error) {
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
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowProposal"),
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

	var list = []*EscrowProposal{}
	for _, wrapper := range result {
		var created map[string]interface{}
		if entry, ok := wrapper["contractEntry"].(map[string]interface{}); ok {
			if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
				if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
					created = ce
				}
			}
		}

		if created != nil {
			args, ok := created["createArgument"].(map[string]interface{})
			if ok {
				amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", args["totalAmount"]), 64)
				list = append(list, &EscrowProposal{
					ID:          created["contractId"].(string),
					Buyer:       fmt.Sprintf("%v", args["buyer"]),
					Seller:      fmt.Sprintf("%v", args["seller"]),
					Issuer:      fmt.Sprintf("%v", args["issuer"]),
					Mediator:    fmt.Sprintf("%v", args["mediator"]),
					Amount:      amount,
					Currency:    fmt.Sprintf("%v", args["currency"]),
					Description: fmt.Sprintf("%v", args["description"]),
				})
			}
		}
	}
	return list, nil
}

func (c *JsonLedgerClient) ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error) {
	buyerParty := c.getParty(BuyerUser)
	cbParty := c.getParty(CentralBankUser)
	sellerParty := c.getParty(req.Seller)
	mediatorParty := c.getParty(EscrowMediatorUser)

	amountStr := fmt.Sprintf("%.10f", req.Amount)

	var milestones []interface{}
	for _, m := range req.Milestones {
		milestones = append(milestones, map[string]interface{}{
			"label":     m.Label,
			"amount":    fmt.Sprintf("%.10f", m.Amount),
			"completed": m.Completed,
		})
	}

	metadataJSON := "{}"
	if req.Metadata.SchemaURL != "" || len(req.Metadata.Payload) > 0 {
		b, _ := json.Marshal(req.Metadata)
		metadataJSON = string(b)
	}

	payload := map[string]interface{}{
		"issuer":      cbParty,
		"buyer":       buyerParty,
		"seller":      sellerParty,
		"mediator":    mediatorParty,
		"totalAmount": amountStr,
		"currency":    req.Currency,
		"description": req.Description,
		"milestones":  milestones,
		"metadata":    metadataJSON,
	}

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("propose-escrow-%d", time.Now().UnixNano()),
			"actAs":     []string{buyerParty, cbParty},
			"userId":    BuyerUser,
			"commands": []interface{}{
				map[string]interface{}{
					"CreateCommand": map[string]interface{}{
						"templateId":      fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowProposal"),
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

	for _, wrapper := range txResp.Transaction.Events {
		var created map[string]interface{}
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			created = ce
		}
		if created != nil && strings.Contains(created["templateId"].(string), "EscrowProposal") {
			args := created["createArgument"].(map[string]interface{})
			amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", args["totalAmount"]), 64)
			return &EscrowProposal{
				ID:          created["contractId"].(string),
				Buyer:       fmt.Sprintf("%v", args["buyer"]),
				Seller:      fmt.Sprintf("%v", args["seller"]),
				Issuer:      fmt.Sprintf("%v", args["issuer"]),
				Mediator:    fmt.Sprintf("%v", args["mediator"]),
				Amount:      amount,
				Currency:    fmt.Sprintf("%v", args["currency"]),
				Description: fmt.Sprintf("%v", args["description"]),
			}, nil
		}
	}

	return nil, fmt.Errorf("proposal created but not found in response")
}

func (c *JsonLedgerClient) AcceptProposal(ctx context.Context, id string, sellerID string) error {
	party := c.getParty(sellerID)
	
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("accept-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    sellerID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowProposal"),
						"contractId":     id,
						"choice":         "Accept",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}

	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
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
		"description":           req.Description,
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
						"templateId":      fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "StablecoinEscrow"),
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

	for _, wrapper := range txResp.Transaction.Events {
		var created map[string]interface{}
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			created = ce
		}
		if created != nil && strings.Contains(created["templateId"].(string), "StablecoinEscrow") {
			args := created["createArgument"].(map[string]interface{})
			amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", args["totalAmount"]), 64)
			return &EscrowContract{
				ID:                    created["contractId"].(string),
				Buyer:                 fmt.Sprintf("%v", args["buyer"]),
				Seller:                fmt.Sprintf("%v", args["seller"]),
				Issuer:                fmt.Sprintf("%v", args["issuer"]),
				Mediator:              fmt.Sprintf("%v", args["mediator"]),
				Amount:                amount,
				Currency:              fmt.Sprintf("%v", args["currency"]),
				State:                 "Active",
				CurrentMilestoneIndex: 0,
			}, nil
		}
	}
	return nil, fmt.Errorf("escrow created but not found in response")
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
		TPSHistory: []ActivityPoint{
			{Date: "10:00", Count: 5},
			{Date: "10:05", Count: 8},
			{Date: "10:10", Count: 12},
			{Date: "10:15", Count: 7},
			{Date: "10:20", Count: 15},
			{Date: "10:25", Count: 10},
		},
		LedgerHealth: LedgerHealth{
			TPS:                12.4,
			CommandSuccessRate: 99.8,
			ActiveContracts:    len(escrows) * 3,
			ParticipantUptime:  "12d 4h 15m",
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
						"templateId":  fmt.Sprintf("%s:%s:%s", Interfacec.PackageID, "Escrow.Interface", "Escrow"),
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
						"templateId":  fmt.Sprintf("%s:%s:%s", Interfacec.PackageID, "Escrow.Interface", "Escrow"),
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

	for _, wrapper := range txResp.Transaction.Events {
		var created map[string]interface{}
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			created = ce
		}
		if created != nil && strings.Contains(created["templateId"].(string), "StablecoinDisputedEscrow") {
			return created["contractId"].(string), nil
		}
	}
	return "", fmt.Errorf("disputed escrow not found in response")
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
						"templateId": fmt.Sprintf("%s:%s:%s", Interfacec.PackageID, "Escrow.Interface", "DisputedEscrow"),
						"contractId": id,
						"choice":     "ResolveDispute",
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
	disputeID, err := c.RaiseDispute(ctx, id)
	if err != nil {
		return err
	}
	return c.ResolveDispute(ctx, disputeID, 100.0, 0.0) // Mock full refund
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
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "StablecoinEscrow"),
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

func (c *JsonLedgerClient) getLedgerEnd(ctx context.Context) (interface{}, error) {
	respBody, err := c.doRawRequest(ctx, "GET", "/v2/state/ledger-end", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Offset json.RawMessage `json:"offset"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	var offset interface{}
	_ = json.Unmarshal(result.Offset, &offset)
	return offset, nil
}

func (c *JsonLedgerClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	offset, err := c.getLedgerEnd(ctx)
	if err != nil { return nil, err }
	body := map[string]interface{}{
		"activeAtOffset": offset,
		"eventFormat": map[string]interface{}{"verbose": true},
		"filtersForAnyParty": map[string]interface{}{
			"templateIds": []string{fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowSettlement")},
		},
	}
	respBody, err := c.doRawRequest(ctx, "POST", "/v2/state/active-contracts", body)
	if err != nil { return nil, err }
	result, _ := parseNDJSON(respBody)
	var list []*EscrowSettlement
	for _, wrapper := range result {
		if entry, ok := wrapper["contractEntry"].(map[string]interface{}); ok {
			if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
				if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
					args := ce["createArgument"].(map[string]interface{})
					amount, _ := strconv.ParseFloat(fmt.Sprintf("%v", args["amount"]), 64)
					list = append(list, &EscrowSettlement{
						ID: ce["contractId"].(string),
						Issuer: fmt.Sprintf("%v", args["issuer"]),
						Recipient: fmt.Sprintf("%v", args["recipient"]),
						Amount: amount,
						Currency: fmt.Sprintf("%v", args["currency"]),
						Status: fmt.Sprintf("%v", args["status"]),
					})
				}
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
			"actAs": []string{party},
			"userId": CentralBankUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId": fmt.Sprintf("%s:%s:%s", Interfacec.PackageID, "Escrow.Interface", "Settlement"),
						"contractId": settlementID,
						"choice": "Settle",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}
