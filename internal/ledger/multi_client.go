package ledger

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

type MultiLedgerClient struct {
	logger   *zap.Logger
	clients  map[string]Client // key is node name (bank, buyer, seller)
	partyMap map[string]string
	mu       sync.RWMutex
}

func NewMultiLedgerClient(logger *zap.Logger, clients map[string]Client) *MultiLedgerClient {
	return &MultiLedgerClient{
		logger:   logger,
		clients:  clients,
		partyMap: make(map[string]string),
	}
}

func (m *MultiLedgerClient) getClientForUser(userID string) Client {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Simple routing based on email domain or hardcoded users
	if strings.Contains(userID, "bank.com") || userID == CentralBankUser || userID == EscrowMediatorUser {
		return m.clients["bank"]
	}
	if strings.Contains(userID, "buyer.com") || userID == BuyerUser {
		return m.clients["buyer"]
	}
	if strings.Contains(userID, "seller.com") || userID == SellerUser {
		return m.clients["seller"]
	}

	// Default to buyer node for dynamic users (like u-google-oauth2-...)
	// unless we implement more complex node-to-party mapping
	return m.clients["buyer"]
}

func (m *MultiLedgerClient) Discover(ctx context.Context) error {
	coreParties := []string{CentralBankUser, BuyerUser, SellerUser, EscrowMediatorUser}
	
	var lastErr error
	for retry := 0; retry < 10; retry++ {
		newMap := make(map[string]string)
		for _, client := range m.clients {
			if err := client.Discover(ctx); err != nil {
				lastErr = err
				continue
			}
			// Aggregate party mappings from this node
			for k, v := range client.GetPartyMap() {
				newMap[k] = v
			}
		}

		// Check if all core parties are found
		allFound := true
		for _, p := range coreParties {
			if _, ok := newMap[p]; !ok {
				allFound = false
				break
			}
		}

		if allFound && len(newMap) >= 4 {
			m.mu.Lock()
			m.partyMap = newMap
			m.mu.Unlock()

			// Push the aggregated map back to all children
			for _, client := range m.clients {
				client.SetPartyMap(newMap)
			}

			m.logger.Info("multi-node party map refreshed and synchronized", zap.Int("totalParties", len(newMap)))
			return nil
		}

		m.logger.Warn("core parties not yet discovered, retrying...", zap.Int("retry", retry), zap.Int("found", len(newMap)))
		time.Sleep(5 * time.Second)
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("failed to discover all core parties after retries")
}

func (m *MultiLedgerClient) SetPartyMap(newMap map[string]string) {
	m.mu.Lock()
	m.partyMap = newMap
	m.mu.Unlock()
	for _, client := range m.clients {
		client.SetPartyMap(newMap)
	}
}

func (m *MultiLedgerClient) GetPartyMap() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make(map[string]string)
	for k, v := range m.partyMap {
		copy[k] = v
	}
	return copy
}

func (m *MultiLedgerClient) GetParty(user string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if id, ok := m.partyMap[user]; ok {
		return id
	}
	return user
}

func (m *MultiLedgerClient) GetOffset() interface{} {
	return m.clients["bank"].GetOffset()
}

func (m *MultiLedgerClient) GetInterfacePackageID() string {
	return m.clients["bank"].GetInterfacePackageID()
}

func (m *MultiLedgerClient) DoRawRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	return m.clients["bank"].DoRawRequest(ctx, method, path, body)
}

func (m *MultiLedgerClient) ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error) {
	return m.clients["bank"].ProposeEscrow(ctx, req)
}

func (m *MultiLedgerClient) SellerAccept(ctx context.Context, id string, userID string) (string, error) {
	return m.getClientForUser(userID).SellerAccept(ctx, id, userID)
}

func (m *MultiLedgerClient) Fund(ctx context.Context, id string, custodyRef string, holdingCid string, userID string) error {
	return m.getClientForUser(userID).Fund(ctx, id, custodyRef, holdingCid, userID)
}

func (m *MultiLedgerClient) Activate(ctx context.Context, id string, userID string) (string, error) {
	return m.getClientForUser(userID).Activate(ctx, id, userID)
}

func (m *MultiLedgerClient) ConfirmConditions(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).ConfirmConditions(ctx, id, userID)
}

func (m *MultiLedgerClient) RaiseDispute(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).RaiseDispute(ctx, id, userID)
}

func (m *MultiLedgerClient) ProposeSettlement(ctx context.Context, id string, proposal SettlementTerms, userID string) (string, error) {
	return m.getClientForUser(userID).ProposeSettlement(ctx, id, proposal, userID)
}

func (m *MultiLedgerClient) RatifySettlement(ctx context.Context, id string, userID string) (string, error) {
	return m.getClientForUser(userID).RatifySettlement(ctx, id, userID)
}

func (m *MultiLedgerClient) FinalizeSettlement(ctx context.Context, id string, userID string) (string, error) {
	return m.getClientForUser(userID).FinalizeSettlement(ctx, id, userID)
}

func (m *MultiLedgerClient) Disburse(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).Disburse(ctx, id, userID)
}

func (m *MultiLedgerClient) Cancel(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).Cancel(ctx, id, userID)
}

func (m *MultiLedgerClient) ExpireEscrow(ctx context.Context, id string, userID string) (string, error) {
	return m.getClientForUser(userID).ExpireEscrow(ctx, id, userID)
}

func (m *MultiLedgerClient) ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error) {
	return m.getClientForUser(userID).ListEscrows(ctx, userID)
}

func (m *MultiLedgerClient) ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error) {
	return m.getClientForUser(userID).ListProposals(ctx, userID)
}

func (m *MultiLedgerClient) GetEscrow(ctx context.Context, id string, userID string) (*EscrowContract, error) {
	return m.getClientForUser(userID).GetEscrow(ctx, id, userID)
}

func (m *MultiLedgerClient) CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, asset Asset, terms EscrowTerms) (*EscrowInvitation, error) {
	return m.getClientForUser(inviterID).CreateInvitation(ctx, inviterID, inviteeEmail, role, inviteeType, asset, terms)
}

func (m *MultiLedgerClient) ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*EscrowProposal, error) {
	return m.getClientForUser(claimantID).ClaimInvitation(ctx, inviteID, claimantID)
}

func (m *MultiLedgerClient) ListInvitations(ctx context.Context, userID string) ([]*EscrowInvitation, error) {
	return m.getClientForUser(userID).ListInvitations(ctx, userID)
}

func (m *MultiLedgerClient) GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error) {
	// Invitations are typically on the bank node
	return m.clients["bank"].GetInvitationByToken(ctx, tokenHash)
}

func (m *MultiLedgerClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	return m.clients["bank"].GetMetrics(ctx, userID)
}

func (m *MultiLedgerClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	return m.clients["bank"].ListSettlements(ctx)
}

func (m *MultiLedgerClient) SettlePayment(ctx context.Context, settlementID string) error {
	return m.clients["bank"].SettlePayment(ctx, settlementID)
}

func (m *MultiLedgerClient) ListWallets(ctx context.Context, userID string) ([]*Wallet, error) {
	return m.getClientForUser(userID).ListWallets(ctx, userID)
}

func (m *MultiLedgerClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	// Try all nodes until one succeeds
	for _, node := range []string{"bank", "buyer", "seller"} {
		if client, ok := m.clients[node]; ok {
			identity, err := client.GetIdentity(ctx, oktaSub)
			if err == nil {
				return identity, nil
			}
		}
	}
	return nil, fmt.Errorf("identity not found on any node: %s", oktaSub)
}

func (m *MultiLedgerClient) ProvisionUser(ctx context.Context, oktaSub string, email string) (*UserIdentity, error) {
	var lastIdentity *UserIdentity
	var lastErr error

	// Broadcast provisioning to ALL nodes to ensure global visibility
	for _, node := range []string{"bank", "buyer", "seller"} {
		if client, ok := m.clients[node]; ok {
			identity, err := client.ProvisionUser(ctx, oktaSub, email)
			if err != nil {
				m.logger.Warn("provisioning failed on node", zap.String("node", node), zap.Error(err))
				lastErr = err
			} else {
				lastIdentity = identity
			}
		}
	}

	if lastIdentity != nil {
		return lastIdentity, nil
	}
	return nil, fmt.Errorf("failed to provision user on any node: %w", lastErr)
}

func (m *MultiLedgerClient) CreateContract(ctx context.Context, userID string, templateID string, payload map[string]interface{}) (string, error) {
	return m.getClientForUser(userID).CreateContract(ctx, userID, templateID, payload)
}

func (m *MultiLedgerClient) SearchPackageID(ctx context.Context, name string) (string, error) {
	return m.clients["bank"].SearchPackageID(ctx, name)
}

// LEGACY
func (m *MultiLedgerClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	return m.clients["bank"].CreateEscrow(ctx, req)
}
func (m *MultiLedgerClient) ReleaseFunds(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).ReleaseFunds(ctx, id, userID)
}
func (m *MultiLedgerClient) ResolveDispute(ctx context.Context, id string, b, s float64, userID string) error {
	return m.getClientForUser(userID).ResolveDispute(ctx, id, b, s, userID)
}
func (m *MultiLedgerClient) RefundBuyer(ctx context.Context, id string) error {
	return m.clients["bank"].RefundBuyer(ctx, id)
}
func (m *MultiLedgerClient) RefundBySeller(ctx context.Context, id string) error {
	return m.clients["bank"].RefundBySeller(ctx, id)
}

var _ Client = (*MultiLedgerClient)(nil)
