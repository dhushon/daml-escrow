package ledger

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Lifecycle Transitions (Directive 05)
// ---------------------------------------------------------------------------

func (c *JsonLedgerClient) ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error) {
	buyerParty := c.getParty(BuyerUser)
	cbParty := c.getParty(CentralBankUser)
	sellerParty := c.getParty(req.Seller)
	mediatorParty := c.getParty(EscrowMediatorUser)

	payload := map[string]interface{}{
		"issuer":   cbParty,
		"buyer":    buyerParty,
		"seller":   sellerParty,
		"mediator": mediatorParty,
		"asset":    toDamlAsset(req.Asset),
		"terms":    toDamlTerms(req.Terms),
		"metadata": toMetadataJSON(req.Metadata),
	}

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("propose-%d", time.Now().UnixNano()),
			"actAs":     []string{buyerParty, cbParty, sellerParty},
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

	return c.parseProposalResponse(respBody)
}

func (c *JsonLedgerClient) Fund(ctx context.Context, id string, custodyRef string, userID string) error {
	party := c.getParty(userID)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("fund-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId": fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowProposal"),
						"contractId": id,
						"choice":     "Fund",
						"choiceArgument": map[string]interface{}{
							"custodyRef": custodyRef,
						},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) Activate(ctx context.Context, id string, userID string) error {
	party := c.getParty(userID)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("activate-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowContract"),
						"contractId":     id,
						"choice":         "Activate",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) ConfirmConditions(ctx context.Context, id string, userID string) error {
	party := c.getParty(userID)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("confirm-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowContract"),
						"contractId":     id,
						"choice":         "ConfirmConditions",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) RaiseDispute(ctx context.Context, id string, userID string) error {
	party := c.getParty(userID)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("dispute-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowContract"),
						"contractId":     id,
						"choice":         "RaiseDispute",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) ProposeSettlement(ctx context.Context, id string, proposal SettlementTerms, userID string) error {
	party := c.getParty(userID)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("propose-settlement-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId": fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "DisputeRecord"),
						"contractId": id,
						"choice":     "ProposeSettlement",
						"choiceArgument": map[string]interface{}{
							"proposal": map[string]interface{}{
								"settlementType": proposal.SettlementType,
								"buyerReturn":    fmt.Sprintf("%.10f", proposal.BuyerReturn),
								"sellerPayment":  fmt.Sprintf("%.10f", proposal.SellerPayment),
								"mediatorFee":    fmt.Sprintf("%.10f", proposal.MediatorFee),
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

func (c *JsonLedgerClient) RatifySettlement(ctx context.Context, id string, userID string) error {
	party := c.getParty(userID)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("ratify-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId": fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "SettlementRecord"),
						"contractId": id,
						"choice":     "RatifySettlement",
						"choiceArgument": map[string]interface{}{
							"actor": party,
						},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) FinalizeSettlement(ctx context.Context, id string, userID string) error {
	buyerParty := c.getParty(BuyerUser)
	sellerParty := c.getParty(SellerUser)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("finalize-%d", time.Now().UnixNano()),
			"actAs":     []string{buyerParty, sellerParty},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "SettlementRecord"),
						"contractId":     id,
						"choice":         "FinalizeSettlement",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) Disburse(ctx context.Context, id string, userID string) error {
	party := c.getParty(userID)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("disburse-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    userID,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "DisbursementOrder"),
						"contractId":     id,
						"choice":         "Disburse",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) Cancel(ctx context.Context, id string, userID string) error {
	return nil
}

func (c *JsonLedgerClient) ExpireEscrow(ctx context.Context, id string, userID string) error {
	party := c.getParty(CentralBankUser)
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("expire-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    CentralBankUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowContract"),
						"contractId":     id,
						"choice":         "ExpireEscrow",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}
	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

// ---------------------------------------------------------------------------
// Queries & Parsers
// ---------------------------------------------------------------------------

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
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowContract"),
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "DisputeRecord"),
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "SettlementRecord"),
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "DisbursementOrder"),
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

	var escrows = []*EscrowContract{}
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
			e, err := c.mapContractToEscrow(created)
			if err == nil {
				escrows = append(escrows, e)
			}
		}
	}
	return escrows, nil
}

func (c *JsonLedgerClient) mapContractToEscrow(created map[string]interface{}) (*EscrowContract, error) {
	templateID, _ := created["templateId"].(string)
	args, ok := created["createArgument"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing createArgument")
	}

	e := &EscrowContract{
		ID:       created["contractId"].(string),
		Buyer:    fmt.Sprintf("%v", args["buyer"]),
		Seller:   fmt.Sprintf("%v", args["seller"]),
		Issuer:   fmt.Sprintf("%v", args["issuer"]),
		Mediator: fmt.Sprintf("%v", args["mediator"]),
	}

	// State mapping
	if strings.Contains(templateID, "DisputeRecord") {
		e.State = "DISPUTED"
	} else if strings.Contains(templateID, "SettlementRecord") {
		e.State = "PROPOSED"
		e.BuyerAccepted = args["buyerAccepted"].(bool)
		e.SellerAccepted = args["sellerAccepted"].(bool)
	} else if strings.Contains(templateID, "DisbursementOrder") {
		e.State = "SETTLED"
	} else if strings.Contains(templateID, "EscrowContract") {
		e.State = fmt.Sprintf("%v", args["state"])
	} else if strings.Contains(templateID, "EscrowProposal") {
		e.State = "DRAFT"
	}

	// Asset
	if assetMap, ok := args["asset"].(map[string]interface{}); ok {
		e.Asset = Asset{
			AssetType: fmt.Sprintf("%v", assetMap["assetType"]),
			AssetID:   fmt.Sprintf("%v", assetMap["assetId"]),
			Amount:    c.parseFloat(assetMap["amount"]),
			Currency:  fmt.Sprintf("%v", assetMap["currency"]),
		}
		if ref, ok := assetMap["custodyRef"].(string); ok {
			e.Asset.CustodyRef = ref
		}
	}

	// Terms
	if termsMap, ok := args["terms"].(map[string]interface{}); ok {
		e.Terms = EscrowTerms{
			ConditionDescription: fmt.Sprintf("%v", termsMap["conditionDescription"]),
			ConditionType:        fmt.Sprintf("%v", termsMap["conditionType"]),
			EvidenceRequired:     fmt.Sprintf("%v", termsMap["evidenceRequired"]),
		}
	}

	return e, nil
}

func (c *JsonLedgerClient) parseFloat(v interface{}) float64 {
	f, _ := strconv.ParseFloat(fmt.Sprintf("%v", v), 64)
	return f
}

// ---------------------------------------------------------------------------
// Invitation & Metrics
// ---------------------------------------------------------------------------

func (c *JsonLedgerClient) CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, asset Asset, terms EscrowTerms) (*EscrowInvitation, error) {
	inviterParty := c.getParty(inviterID)
	mediatorParty := c.getParty(EscrowMediatorUser)
	issuerParty := c.getParty(CentralBankUser)

	tokenHash := fmt.Sprintf("hash-%d", time.Now().UnixNano())

	payload := map[string]interface{}{
		"inviter":      inviterParty,
		"mediator":     mediatorParty,
		"issuer":       issuerParty,
		"inviteeEmail": inviteeEmail,
		"inviteeRole":  role,
		"inviteeType":  inviteeType,
		"tokenHash":    tokenHash,
		"asset":        toDamlAsset(asset),
		"terms":        toDamlTerms(terms),
		"metadata":     "{}",
	}

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("invite-%d", time.Now().UnixNano()),
			"actAs":     []string{inviterParty, mediatorParty, issuerParty},
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
		return nil, err
	}
	
	for _, wrapper := range txResp.Transaction.Events {
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			if strings.Contains(ce["templateId"].(string), "EscrowInvitation") {
				return &EscrowInvitation{
					ID:           ce["contractId"].(string),
					TokenHash:    tokenHash,
					InviteeEmail: inviteeEmail,
				}, nil
			}
		}
	}
	
	return nil, fmt.Errorf("invitation not found in transaction")
}

func (c *JsonLedgerClient) ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*EscrowProposal, error) {
	party := c.getParty(EscrowMediatorUser)
	claimantParty := c.getParty(claimantID)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("claim-%d", time.Now().UnixNano()),
			"actAs":     []string{party},
			"userId":    EscrowMediatorUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId": fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "EscrowInvitation"),
						"contractId": inviteID,
						"choice":     "Claim",
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

	return c.parseProposalResponse(respBody)
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

	var invites = []*EscrowInvitation{}
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
			args := created["createArgument"].(map[string]interface{})
			invites = append(invites, &EscrowInvitation{
				ID:           created["contractId"].(string),
				Inviter:      fmt.Sprintf("%v", args["inviter"]),
				Mediator:     fmt.Sprintf("%v", args["mediator"]),
				Issuer:       fmt.Sprintf("%v", args["issuer"]),
				InviteeEmail: fmt.Sprintf("%v", args["inviteeEmail"]),
				InviteeRole:  fmt.Sprintf("%v", args["inviteeRole"]),
				TokenHash:    fmt.Sprintf("%v", args["tokenHash"]),
			})
		}
	}
	return invites, nil
}

func (c *JsonLedgerClient) GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error) {
	invites, err := c.ListInvitations(ctx, EscrowMediatorUser)
	if err != nil {
		return nil, err
	}
	for _, inv := range invites {
		if inv.TokenHash == tokenHash {
			return inv, nil
		}
	}
	return nil, fmt.Errorf("invitation not found for token")
}

func (c *JsonLedgerClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	escrows, _ := c.ListEscrows(ctx, userID)
	
	var totalValue float64
	activeCount := 0
	for _, e := range escrows {
		if e.State == "ACTIVE" || e.State == "FUNDED" {
			activeCount++
			totalValue += e.Asset.Amount
		}
	}

	return &LedgerMetrics{
		TotalActiveEscrows: activeCount,
		TotalValueInEscrow: totalValue,
	}, nil
}

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
						fmt.Sprintf("%s:%s:%s", c.PackageID, "StablecoinEscrow", "DisbursementOrder"),
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

	var settlements = []*EscrowSettlement{}
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
			args := created["createArgument"].(map[string]interface{})
			asset := args["asset"].(map[string]interface{})
			settlements = append(settlements, &EscrowSettlement{
				ID:        created["contractId"].(string),
				Issuer:    fmt.Sprintf("%v", args["issuer"]),
				Recipient: "Multiple",
				Amount:    c.parseFloat(asset["amount"]),
				Currency:  fmt.Sprintf("%v", asset["currency"]),
				Status:    "Pending Disbursement",
			})
		}
	}
	return settlements, nil
}

func (c *JsonLedgerClient) SettlePayment(ctx context.Context, settlementID string) error {
	return c.Disburse(ctx, settlementID, CentralBankUser)
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

	var proposals = []*EscrowProposal{}
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
			args := created["createArgument"].(map[string]interface{})
			asset := args["asset"].(map[string]interface{})
			proposals = append(proposals, &EscrowProposal{
				ID:       created["contractId"].(string),
				Buyer:    fmt.Sprintf("%v", args["buyer"]),
				Seller:   fmt.Sprintf("%v", args["seller"]),
				Issuer:   fmt.Sprintf("%v", args["issuer"]),
				Mediator: fmt.Sprintf("%v", args["mediator"]),
				Asset: Asset{
					Amount:   c.parseFloat(asset["amount"]),
					Currency: fmt.Sprintf("%v", asset["currency"]),
				},
			})
		}
	}
	return proposals, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func toDamlAsset(a Asset) map[string]interface{} {
	return map[string]interface{}{
		"assetType":  a.AssetType,
		"assetId":    a.AssetID,
		"amount":     fmt.Sprintf("%.10f", a.Amount),
		"currency":   a.Currency,
		"custodyRef": nil,
	}
}

func toDamlTerms(t EscrowTerms) map[string]interface{} {
	return map[string]interface{}{
		"conditionDescription": t.ConditionDescription,
		"conditionType":        t.ConditionType,
		"evidenceRequired":     t.EvidenceRequired,
		"expiryDate":           t.ExpiryDate.Format(time.RFC3339Nano),
		"gracePeriodDays":      t.GracePeriodDays,
		"disputeWindowDays":    t.DisputeWindowDays,
		"partialSchedule":      []interface{}{},
	}
}

func toMetadataJSON(m EscrowMetadata) string {
	b, _ := json.Marshal(m)
	return string(b)
}

func (c *JsonLedgerClient) parseProposalResponse(body []byte) (*EscrowProposal, error) {
	var txResp v2TransactionResponse
	if err := json.Unmarshal(body, &txResp); err != nil {
		return nil, err
	}
	for _, wrapper := range txResp.Transaction.Events {
		var created map[string]interface{}
		if ce, ok := wrapper["CreatedEvent"].(map[string]interface{}); ok {
			created = ce
		}
		if created != nil && strings.Contains(created["templateId"].(string), "EscrowProposal") {
			args := created["createArgument"].(map[string]interface{})
			assetMap := args["asset"].(map[string]interface{})
			return &EscrowProposal{
				ID:       created["contractId"].(string),
				Buyer:    fmt.Sprintf("%v", args["buyer"]),
				Seller:   fmt.Sprintf("%v", args["seller"]),
				Issuer:   fmt.Sprintf("%v", args["issuer"]),
				Mediator: fmt.Sprintf("%v", args["mediator"]),
				Asset: Asset{
					Amount:   c.parseFloat(assetMap["amount"]),
					Currency: fmt.Sprintf("%v", assetMap["currency"]),
				},
			}, nil
		}
	}
	return nil, fmt.Errorf("not found in transaction")
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

// ---------------------------------------------------------------------------
// LEGACY
// ---------------------------------------------------------------------------

func (c *JsonLedgerClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) { return nil, nil }
func (c *JsonLedgerClient) ReleaseFunds(ctx context.Context, id string, userID string) error { return nil }
func (c *JsonLedgerClient) ResolveDispute(ctx context.Context, id string, b, s float64, userID string) error { return nil }
func (c *JsonLedgerClient) RefundBuyer(ctx context.Context, id string) error { return nil }
func (c *JsonLedgerClient) RefundBySeller(ctx context.Context, id string) error { return nil }
