package ledger

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockLedgerClient struct {
	mock.Mock
	Client
}

func (m *MockLedgerClient) Discover(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockLedgerClient) ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EscrowProposal), args.Error(1)
}

func (m *MockLedgerClient) SellerAccept(ctx context.Context, id string, userID string) (string, error) {
	args := m.Called(ctx, id, userID)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) Fund(ctx context.Context, id string, custodyRef string, holdingCid string, userID string) error {
	args := m.Called(ctx, id, custodyRef, holdingCid, userID)
	return args.Error(0)
}

func (m *MockLedgerClient) Activate(ctx context.Context, id string, userID string) (string, error) {
	args := m.Called(ctx, id, userID)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) ConfirmConditions(ctx context.Context, id string, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockLedgerClient) RaiseDispute(ctx context.Context, id string, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockLedgerClient) ProposeSettlement(ctx context.Context, id string, proposal SettlementTerms, userID string) (string, error) {
	args := m.Called(ctx, id, proposal, userID)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) RatifySettlement(ctx context.Context, id string, userID string) (string, error) {
	args := m.Called(ctx, id, userID)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) FinalizeSettlement(ctx context.Context, id string, userID string) (string, error) {
	args := m.Called(ctx, id, userID)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) Disburse(ctx context.Context, id string, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockLedgerClient) Cancel(ctx context.Context, id string, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockLedgerClient) ExpireEscrow(ctx context.Context, id string, userID string) (string, error) {
	args := m.Called(ctx, id, userID)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) GetEscrow(ctx context.Context, id string, userID string) (*EscrowContract, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EscrowContract), args.Error(1)
}

func (m *MockLedgerClient) ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*EscrowContract), args.Error(1)
}

func (m *MockLedgerClient) ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*EscrowProposal), args.Error(1)
}

func (m *MockLedgerClient) CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, asset Asset, terms EscrowTerms) (*EscrowInvitation, error) {
	args := m.Called(ctx, inviterID, inviteeEmail, role, inviteeType, asset, terms)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EscrowInvitation), args.Error(1)
}

func (m *MockLedgerClient) ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*EscrowProposal, error) {
	args := m.Called(ctx, inviteID, claimantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EscrowProposal), args.Error(1)
}

func (m *MockLedgerClient) ListInvitations(ctx context.Context, userID string) ([]*EscrowInvitation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*EscrowInvitation), args.Error(1)
}

func (m *MockLedgerClient) GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error) {
	args := m.Called(ctx, tokenHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EscrowInvitation), args.Error(1)
}

func (m *MockLedgerClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	args := m.Called(ctx, oktaSub)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserIdentity), args.Error(1)
}

func (m *MockLedgerClient) ListIdentities(ctx context.Context) ([]*UserIdentity, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*UserIdentity), args.Error(1)
}

func (m *MockLedgerClient) ProvisionUser(ctx context.Context, oktaSub string, email string, scopes []string) (*UserIdentity, error) {
	args := m.Called(ctx, oktaSub, email, scopes)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserIdentity), args.Error(1)
}

func (m *MockLedgerClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*LedgerMetrics), args.Error(1)
}

func (m *MockLedgerClient) ListWallets(ctx context.Context, userID string) ([]*Wallet, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*Wallet), args.Error(1)
}

func (m *MockLedgerClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*EscrowSettlement), args.Error(1)
}

func (m *MockLedgerClient) SettlePayment(ctx context.Context, settlementID string) error {
	args := m.Called(ctx, settlementID)
	return args.Error(0)
}

func (m *MockLedgerClient) GetInterfacePackageID() string {
	return "mock-iface-pkg"
}

func (m *MockLedgerClient) SearchPackageID(ctx context.Context, name string) (string, error) {
	args := m.Called(ctx, name)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) CreateContract(ctx context.Context, userID string, templateID string, payload map[string]interface{}) (string, error) {
	args := m.Called(ctx, userID, templateID, payload)
	return args.String(0), args.Error(1)
}

func (m *MockLedgerClient) GetParty(user string) string {
	return "mock-party-" + user
}

func (m *MockLedgerClient) GetOffset() interface{} {
	return nil
}

func (m *MockLedgerClient) DoRawRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	args := m.Called(ctx, method, path, body)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockLedgerClient) GetPartyMap() map[string]string {
	return map[string]string{}
}

func (m *MockLedgerClient) SetPartyMap(map[string]string) {}

// LEGACY
func (m *MockLedgerClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*EscrowContract), args.Error(1)
}

func (m *MockLedgerClient) ReleaseFunds(ctx context.Context, id string, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockLedgerClient) ResolveDispute(ctx context.Context, id string, b, s float64, userID string) error {
	args := m.Called(ctx, id, b, s, userID)
	return args.Error(0)
}

func (m *MockLedgerClient) RefundBuyer(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLedgerClient) RefundBySeller(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// Stablecoin Mocks
type MockStablecoinProvider struct {
	mock.Mock
}

func (m *MockStablecoinProvider) CreateWallet(ctx context.Context, userID string) (string, error) {
	args := m.Called(ctx, userID)
	return args.String(0), args.Error(1)
}

func (m *MockStablecoinProvider) GetBalance(ctx context.Context, walletID string, currency string) (float64, error) {
	args := m.Called(ctx, walletID, currency)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockStablecoinProvider) Transfer(ctx context.Context, from, to string, amt float64, cur, ref string) (string, error) {
	args := m.Called(ctx, from, to, amt, cur, ref)
	return args.String(0), args.Error(1)
}

func (m *MockStablecoinProvider) VerifyTransfer(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

var _ Client = (*MockLedgerClient)(nil)
var _ StablecoinProvider = (*MockStablecoinProvider)(nil)
