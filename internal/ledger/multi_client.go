package ledger

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

type MultiLedgerClient struct {
	logger  *zap.Logger
	clients map[string]Client // key is node name (bank, buyer, seller)
}

func NewMultiLedgerClient(logger *zap.Logger, clients map[string]Client) *MultiLedgerClient {
	return &MultiLedgerClient{
		logger:  logger,
		clients: clients,
	}
}

func (m *MultiLedgerClient) getClientForUser(userID string) Client {
	// Simple routing based on email domain or hardcoded users
	// bank.com -> bank
	// buyer.com -> buyer
	// seller.com -> seller
	
	if userID == CentralBankUser || userID == EscrowMediatorUser || strings.HasSuffix(userID, "@bank.com") || strings.HasSuffix(userID, "@mediator.com") {
		return m.clients["bank"]
	}
	if userID == BuyerUser || strings.HasSuffix(userID, "@buyer.com") {
		return m.clients["buyer"]
	}
	if userID == SellerUser || strings.HasSuffix(userID, "@seller.com") {
		return m.clients["seller"]
	}

	// Default to buyer for now if unknown
	return m.clients["buyer"]
}

func (m *MultiLedgerClient) Discover(ctx context.Context) error {
	for name, client := range m.clients {
		if err := client.Discover(ctx); err != nil {
			m.logger.Error("discovery failed for node", zap.String("node", name), zap.Error(err))
			return err
		}
	}
	return nil
}

func (m *MultiLedgerClient) ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error) {
	return m.getClientForUser(req.Buyer).ProposeEscrow(ctx, req)
}

func (m *MultiLedgerClient) Fund(ctx context.Context, id string, custodyRef string, userID string) error {
	return m.getClientForUser(userID).Fund(ctx, id, custodyRef, userID)
}

func (m *MultiLedgerClient) Activate(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).Activate(ctx, id, userID)
}

func (m *MultiLedgerClient) ConfirmConditions(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).ConfirmConditions(ctx, id, userID)
}

func (m *MultiLedgerClient) RaiseDispute(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).RaiseDispute(ctx, id, userID)
}

func (m *MultiLedgerClient) ProposeSettlement(ctx context.Context, id string, proposal SettlementTerms, userID string) error {
	return m.getClientForUser(userID).ProposeSettlement(ctx, id, proposal, userID)
}

func (m *MultiLedgerClient) RatifySettlement(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).RatifySettlement(ctx, id, userID)
}

func (m *MultiLedgerClient) FinalizeSettlement(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).FinalizeSettlement(ctx, id, userID)
}

func (m *MultiLedgerClient) Disburse(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).Disburse(ctx, id, userID)
}

func (m *MultiLedgerClient) Cancel(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).Cancel(ctx, id, userID)
}

func (m *MultiLedgerClient) ExpireEscrow(ctx context.Context, id string, userID string) error {
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

func (m *MultiLedgerClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	return m.getClientForUser(userID).GetMetrics(ctx, userID)
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

func (m *MultiLedgerClient) GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error) {
	for _, client := range m.clients {
		inv, err := client.GetInvitationByToken(ctx, tokenHash)
		if err == nil {
			return inv, nil
		}
	}
	return nil, fmt.Errorf("invitation not found")
}

func (m *MultiLedgerClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	for _, client := range m.clients {
		ident, err := client.GetIdentity(ctx, oktaSub)
		if err == nil && ident != nil {
			return ident, nil
		}
	}
	return nil, nil
}

func (m *MultiLedgerClient) ProvisionUser(ctx context.Context, oktaSub string, email string) (*UserIdentity, error) {
	return m.getClientForUser(email).ProvisionUser(ctx, oktaSub, email)
}

func (m *MultiLedgerClient) getParty(user string) string {
	return m.getClientForUser(user).getParty(user)
}

func (m *MultiLedgerClient) getOffset() interface{} {
	// Offset is tricky in multi-client, but usually only one node's offset matters for a specific stream
	return m.clients["bank"].getOffset()
}

// LEGACY
func (m *MultiLedgerClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	return m.getClientForUser(req.Buyer).CreateEscrow(ctx, req)
}

func (m *MultiLedgerClient) ReleaseFunds(ctx context.Context, id string, userID string) error {
	return m.getClientForUser(userID).ReleaseFunds(ctx, id, userID)
}

func (m *MultiLedgerClient) ResolveDispute(ctx context.Context, id string, b, s float64, userID string) error {
	return m.getClientForUser(userID).ResolveDispute(ctx, id, b, s, userID)
}

func (m *MultiLedgerClient) RefundBuyer(ctx context.Context, id string) error {
	// Legacy refund usually Bank action
	return m.clients["bank"].RefundBuyer(ctx, id)
}

func (m *MultiLedgerClient) RefundBySeller(ctx context.Context, id string) error {
	// Seller node
	return m.clients["seller"].RefundBySeller(ctx, id)
}

var _ Client = (*MultiLedgerClient)(nil)
