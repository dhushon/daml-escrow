package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// User IDs mapped in init.canton
const (
	CentralBankUser    = "CentralBank"
	BuyerUser          = "Buyer"
	SellerUser         = "Seller"
	EscrowMediatorUser = "EscrowMediator"
)

const PackageID = "fbbab3d109ee72290549160ecf2cdd9c1fa130fd97a93b9fc0dd5e376399b7c3"

type JsonLedgerClient struct {
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
	partyMap   map[string]string // Maps User ID -> Canton Party ID
}

func NewJsonLedgerClient(logger *zap.Logger, host string, port int) *JsonLedgerClient {
	c := &JsonLedgerClient{
		logger: logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  fmt.Sprintf("http://%s:%d", host, port),
		partyMap: make(map[string]string),
	}
	
	// Pre-populate party map
	_ = c.refreshPartyMap(context.Background())
	
	return c
}

func (c *JsonLedgerClient) refreshPartyMap(ctx context.Context) error {
	respBody, err := c.doRawRequest(ctx, "GET", "/v2/parties", nil)
	if err != nil {
		return err
	}

	var result struct {
		PartyDetails []struct {
			Party string `json:"party"`
		} `json:"partyDetails"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("failed to unmarshal party list: %w", err)
	}

	for _, p := range result.PartyDetails {
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

// Internal structures for parsing V2 Transaction responses
type v2TransactionResponse struct {
	Transaction struct {
		Events []map[string]interface{} `json:"events"`
		Offset json.RawMessage          `json:"offset"`
	} `json:"transaction"`
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

				metadata := make(map[string]interface{})
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
	c.logger.Info("creating escrow via JSON API V2", zap.Any("request", req))

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
	if len(req.Metadata) > 0 {
		if b, err := json.Marshal(req.Metadata); err == nil {
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
						fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "DisputedEscrow"),
					},
				},
			},
		},
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/state/active-contracts", body)
	if err != nil {
		return nil, err
	}

	var result []interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		var rawList []interface{}
		if err2 := json.Unmarshal(respBody, &rawList); err2 == nil {
			result = rawList
		} else {
			return nil, fmt.Errorf("failed to decode active contracts result: %w", err)
		}
	}

	var escrows []*EscrowContract
	for _, item := range result {
		if m, ok := item.(map[string]interface{}); ok {
			var created map[string]interface{}
			state := "Active"
			
			if entry, ok := m["contractEntry"].(map[string]interface{}); ok {
				if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
					if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
						created = ce
						templateId, _ := ce["templateId"].(string)
						if strings.Contains(templateId, "DisputedEscrow") {
							state = "Disputed"
							// For DisputedEscrow, the escrow data is nested
							if nest, ok := ce["createArguments"].(map[string]interface{}); ok {
								if e, ok := nest["escrow"].(map[string]interface{}); ok {
									created = map[string]interface{}{
										"contractId":      ce["contractId"],
										"createArguments": e,
									}
								}
							}
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

					metadata := make(map[string]interface{})
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
						State:                 state,
						Milestones:            ms,
						CurrentMilestoneIndex: curIdx,
						Metadata:              metadata,
					})
				}
			}
		}
	}
	return escrows, nil
}

func (c *JsonLedgerClient) GetEscrow(ctx context.Context, id string) (*EscrowContract, error) {
	c.logger.Info("querying escrow via JSON API V2", zap.String("id", id))

	for retry := 0; retry < 5; retry++ {
		// Search across all possible users since we don't know who owns this ID here
		users := []string{BuyerUser, SellerUser, CentralBankUser, EscrowMediatorUser}
		for _, u := range users {
			contracts, err := c.ListEscrows(ctx, u)
			if err != nil {
				continue
			}
			for _, item := range contracts {
				if item.ID == id {
					return item, nil
				}
			}
		}
		
		c.logger.Debug("escrow not found yet, retrying...", zap.Int("attempt", retry+1))
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
	buyerParty := c.getParty(BuyerUser)
	
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("release-%d", time.Now().UnixNano()),
			"actAs":     []string{buyerParty},
			"userId":    BuyerUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinEscrow"),
						"contractId":     id,
						"choice":         "ApproveMilestone",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}

	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) RaiseDispute(ctx context.Context, id string) (string, error) {
	buyerParty := c.getParty(BuyerUser)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("dispute-%d", time.Now().UnixNano()),
			"actAs":     []string{buyerParty},
			"userId":    BuyerUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinEscrow"),
						"contractId":     id,
						"choice":         "RaiseDispute",
						"choiceArgument": map[string]interface{}{},
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

	escrow, err := c.extractContract(txResp.Transaction.Events, "DisputedEscrow")
	if err != nil {
		return "", err
	}
	return escrow.ID, nil
}

func (c *JsonLedgerClient) ResolveDispute(ctx context.Context, id string, payoutToBuyer, payoutToSeller float64) error {
	mediatorParty := c.getParty(EscrowMediatorUser)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("resolve-%d", time.Now().UnixNano()),
			"actAs":     []string{mediatorParty},
			"userId":    EscrowMediatorUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "DisputedEscrow"),
						"contractId":     id,
						"choice":         "ResolveDispute",
						"choiceArgument": map[string]interface{}{
							"payoutToBuyer":  fmt.Sprintf("%.10f", payoutToBuyer),
							"payoutToSeller": fmt.Sprintf("%.10f", payoutToSeller),
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
	c.logger.Info("automating full refund for buyer", zap.String("escrowID", id))
	
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
		return fmt.Errorf("no remaining funds to refund in escrow %s", id)
	}

	disputeID, err := c.RaiseDispute(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to raise dispute for refund: %w", err)
	}

	err = c.ResolveDispute(ctx, disputeID, remaining, 0.0)
	if err != nil {
		return fmt.Errorf("failed to resolve dispute for refund: %w", err)
	}

	c.logger.Info("automated refund complete", zap.String("escrowID", id), zap.Float64("amount", remaining))
	return nil
}

func (c *JsonLedgerClient) RefundBySeller(ctx context.Context, id string) error {
	sellerParty := c.getParty(SellerUser)

	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("seller-refund-%d", time.Now().UnixNano()),
			"actAs":     []string{sellerParty},
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

func (c *JsonLedgerClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	// We check for settlements visible to any of our major users
	majorUsers := []string{CentralBankUser, BuyerUser, SellerUser}
	uniqueMap := make(map[string]*EscrowSettlement)

	for _, u := range majorUsers {
		party := c.getParty(u)
		offset, err := c.getLedgerEnd(ctx)
		if err != nil {
			continue
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
			continue
		}

		var result []interface{}
		_ = json.Unmarshal(respBody, &result)
		
		for _, item := range result {
			if m, ok := item.(map[string]interface{}); ok {
				var created map[string]interface{}
				if entry, ok := m["contractEntry"].(map[string]interface{}); ok {
					if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
						if ce, ok := active["createdEvent"].(map[string]interface{}); ok {
							created = ce
						}
					}
				}

				if created != nil {
					cid, _ := created["contractId"].(string)
					var args map[string]interface{}
					if a, ok := created["createArguments"].(map[string]interface{}); ok {
						args = a
					} else if a, ok := created["createArgument"].(map[string]interface{}); ok {
						args = a
					}

					if args != nil {
						amt, _ := strconv.ParseFloat(fmt.Sprintf("%v", args["amount"]), 64)
						uniqueMap[cid] = &EscrowSettlement{
							ID:        cid,
							Issuer:    fmt.Sprintf("%v", args["issuer"]),
							Recipient: fmt.Sprintf("%v", args["recipient"]),
							Amount:    amt,
							Currency:  fmt.Sprintf("%v", args["currency"]),
							Status:    fmt.Sprintf("%v", args["status"]),
						}
					}
				}
			}
		}
	}

	var list []*EscrowSettlement
	for _, v := range uniqueMap {
		list = append(list, v)
	}
	return list, nil
}

func (c *JsonLedgerClient) SettlePayment(ctx context.Context, settlementID string) error {
	cbParty := c.getParty(CentralBankUser)
	
	body := map[string]interface{}{
		"commands": map[string]interface{}{
			"commandId": fmt.Sprintf("settle-%d", time.Now().UnixNano()),
			"actAs":     []string{cbParty},
			"userId":    CentralBankUser,
			"commands": []interface{}{
				map[string]interface{}{
					"ExerciseCommand": map[string]interface{}{
						"templateId":     fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "EscrowSettlement"),
						"contractId":     settlementID,
						"choice":         "Settle",
						"choiceArgument": map[string]interface{}{},
					},
				},
			},
		},
	}

	_, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait-for-transaction", body)
	return err
}

func (c *JsonLedgerClient) getOffset() interface{} {
	return nil // Not used externally in this refactor
}

func (c *JsonLedgerClient) doRawRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := c.baseURL + path
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("JSON API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
