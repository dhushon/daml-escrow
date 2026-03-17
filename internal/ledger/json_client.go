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
	"sync"
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

const PackageID = "ec35fce924adbefbae43d1f546879c29fdc42b9efac531f4de8eaeb39a5693c1"

type JsonLedgerClient struct {
	logger     *zap.Logger
	httpClient *http.Client
	baseURL    string
	partyMap   map[string]string // Maps User ID -> Canton Party ID
	lastOffset int64
	mu         sync.RWMutex
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

func (c *JsonLedgerClient) updateOffset(off int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if off > c.lastOffset {
		c.lastOffset = off
	}
}

func (c *JsonLedgerClient) getOffset() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastOffset
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

	payload := map[string]interface{}{
		"issuer":      cbParty,
		"buyer":       buyerParty,
		"seller":      sellerParty,
		"mediator":    mediatorParty,
		"totalAmount": amountStr,
		"currency":    req.Currency,
		"description": "Escrow for " + req.Currency,
		"milestones": []interface{}{
			map[string]interface{}{
				"label":     "Full Payment",
				"amount":    amountStr,
				"completed": false,
			},
		},
		"currentMilestoneIndex": 0,
	}

	body := map[string]interface{}{
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
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait", body)
	if err != nil {
		return nil, err
	}

	var submitResp struct {
		CompletionOffset int64 `json:"completionOffset"`
	}
	if err := json.Unmarshal(respBody, &submitResp); err == nil {
		c.updateOffset(submitResp.CompletionOffset)
	}

	time.Sleep(1 * time.Second)
	
	contracts, err := c.listEscrows(ctx, c.getOffset())
	if err != nil {
		return nil, fmt.Errorf("failed to list escrows after creation: %w", err)
	}

	if len(contracts) == 0 {
		return nil, fmt.Errorf("escrow contract not found after creation at offset %d", c.getOffset())
	}

	return contracts[len(contracts)-1], nil
}

func (c *JsonLedgerClient) listEscrows(ctx context.Context, offset int64) ([]*EscrowContract, error) {
	buyerParty := c.getParty(BuyerUser)
	body := map[string]interface{}{
		"activeAtOffset": offset,
		"eventFormat": map[string]interface{}{
			"verbose": true,
			"filtersByParty": map[string]interface{}{
				buyerParty: map[string]interface{}{
					"templateIds": []string{
						fmt.Sprintf("%s:%s:%s", PackageID, "StablecoinEscrow", "StablecoinEscrow"),
					},
				},
			},
		},
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/state/active-contracts", body)
	if err != nil {
		return nil, err
	}

	var envelope struct {
		Result []interface{} `json:"result"`
	}
	if err := json.Unmarshal(respBody, &envelope); err != nil {
		var rawList []interface{}
		if err2 := json.Unmarshal(respBody, &rawList); err2 == nil {
			envelope.Result = rawList
		} else {
			return nil, fmt.Errorf("failed to decode active contracts result: %w", err)
		}
	}

	var escrows []*EscrowContract
	for _, item := range envelope.Result {
		if m, ok := item.(map[string]interface{}); ok {
			var created map[string]interface{}
			if entry, ok := m["contractEntry"].(map[string]interface{}); ok {
				if active, ok := entry["JsActiveContract"].(map[string]interface{}); ok {
					if c, ok := active["createdEvent"].(map[string]interface{}); ok {
						created = c
					}
				}
			}

			if created != nil {
				contractId, _ := created["contractId"].(string)
				args, _ := created["createArgument"].(map[string]interface{})
				
				if args != nil {
					amountStr := fmt.Sprintf("%v", args["totalAmount"])
					amount, _ := strconv.ParseFloat(amountStr, 64)
					escrows = append(escrows, &EscrowContract{
						ID:       contractId,
						Buyer:    fmt.Sprintf("%v", args["buyer"]),
						Seller:   fmt.Sprintf("%v", args["seller"]),
						Amount:   amount,
						Currency: fmt.Sprintf("%v", args["currency"]),
						State:    "Active",
					})
				}
			}
		}
	}
	return escrows, nil
}

func (c *JsonLedgerClient) GetEscrow(ctx context.Context, id string) (*EscrowContract, error) {
	c.logger.Info("querying escrow via JSON API V2", zap.String("id", id))

	contracts, err := c.listEscrows(ctx, c.getOffset())
	if err != nil {
		return nil, err
	}

	for _, item := range contracts {
		if item.ID == id {
			return item, nil
		}
	}

	return nil, fmt.Errorf("escrow not found: %s", id)
}

func (c *JsonLedgerClient) ReleaseFunds(ctx context.Context, id string) error {
	buyerParty := c.getParty(BuyerUser)
	
	body := map[string]interface{}{
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
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait", body)
	if err == nil {
		var submitResp struct {
			CompletionOffset int64 `json:"completionOffset"`
		}
		if err := json.Unmarshal(respBody, &submitResp); err == nil {
			c.updateOffset(submitResp.CompletionOffset)
		}
	}
	return err
}

func (c *JsonLedgerClient) RefundBuyer(ctx context.Context, id string) error {
	buyerParty := c.getParty(BuyerUser)

	body := map[string]interface{}{
		"commandId": fmt.Sprintf("refund-%d", time.Now().UnixNano()),
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
	}

	respBody, err := c.doRawRequest(ctx, "POST", "/v2/commands/submit-and-wait", body)
	if err == nil {
		var submitResp struct {
			CompletionOffset int64 `json:"completionOffset"`
		}
		if err := json.Unmarshal(respBody, &submitResp); err == nil {
			c.updateOffset(submitResp.CompletionOffset)
		}
	}
	return err
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
