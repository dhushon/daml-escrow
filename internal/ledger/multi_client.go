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
	if strings.Contains(userID, "depositor.com") || userID == DepositorUser {
		return m.clients["depositor"]
	}
	if strings.Contains(userID, "beneficiary.com") || userID == BeneficiaryUser {
		return m.clients["beneficiary"]
	}

	return m.clients["depositor"]
}

func (m *MultiLedgerClient) Discover(ctx context.Context, wait bool) error {
	coreParties := []string{CentralBankUser, DepositorUser, BeneficiaryUser, EscrowMediatorUser}
	
	var lastErr error
	for retry := 0; retry < 1; retry++ {
		newMap := make(map[string]string)
		discoveredClients := make(map[Client]bool)

		for _, client := range m.clients {
			if discoveredClients[client] {
				for k, v := range client.GetPartyMap() {
					newMap[k] = v
				}
				continue
			}

			if err := client.Discover(ctx, wait); err != nil {
				lastErr = err
				continue
			}
			discoveredClients[client] = true

			for k, v := range client.GetPartyMap() {
				newMap[k] = v
			}
		}

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

func (m *MultiLedgerClient) Activate(ctx context.Context, id string, actAs []string) (string, error) {
	return m.getClientForUser(actAs[0]).Activate(ctx, id, actAs)
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

func (m *MultiLedgerClient) Disburse(ctx context.Context, id string, actAs []string) error {
	return m.getClientForUser(actAs[0]).Disburse(ctx, id, actAs)
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
	for _, node := range []string{"bank", "depositor", "beneficiary"} {
		if client, ok := m.clients[node]; ok {
			identity, err := client.GetIdentity(ctx, oktaSub)
			if err == nil {
				return identity, nil
			}
		}
	}
	return nil, fmt.Errorf("identity not found on any node: %s", oktaSub)
}

func (m *MultiLedgerClient) ListIdentities(ctx context.Context) ([]*UserIdentity, error) {
	uniqueUsers := make(map[string]*UserIdentity)
	for _, client := range m.clients {
		identities, err := client.ListIdentities(ctx)
		if err != nil {
			continue
		}
		for _, id := range identities {
			uniqueUsers[id.DamlUserID] = id
		}
	}
	result := make([]*UserIdentity, 0, len(uniqueUsers))
	for _, id := range uniqueUsers {
		result = append(result, id)
	}
	return result, nil
}

func (m *MultiLedgerClient) ProvisionUser(ctx context.Context, oktaSub string, email string, role string, scopes []string) (*UserIdentity, error) {
	var firstIdentity *UserIdentity
	var lastErr error
	for _, node := range []string{"bank", "depositor", "beneficiary"} {
		if client, ok := m.clients[node]; ok {
			identity, err := client.ProvisionUser(ctx, oktaSub, email, role, scopes)
			if err != nil {
				if strings.Contains(err.Error(), "already exists") {
					if firstIdentity == nil {
						if id, err := client.GetIdentity(ctx, oktaSub); err == nil {
							firstIdentity = id
						}
					}
					continue
				}
				lastErr = err
			} else {
				if firstIdentity == nil {
					firstIdentity = identity
				}
			}
		}
	}
	if firstIdentity != nil {
		return firstIdentity, nil
	}
	return nil, fmt.Errorf("failed to provision user on any node: %v", lastErr)
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
