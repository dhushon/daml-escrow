package ledger

import (
	"context"
	"fmt"
)

// DamlClient implements the LedgerClient interface using gRPC.
// Currently acts as a placeholder for Phase 8.
type DamlClient struct {
}

func NewDamlClient() *DamlClient {
	return &DamlClient{}
}

func (c *DamlClient) Discover(ctx context.Context) error {
	return fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ProposeEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowProposal, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) SellerAccept(ctx context.Context, id string, userID string) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) Fund(ctx context.Context, id string, custodyRef string, holdingCid string, userID string) error {
	return fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) Activate(ctx context.Context, id string, userID string) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ConfirmConditions(ctx context.Context, id string, userID string) error {
	return fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) RaiseDispute(ctx context.Context, id string, userID string) error {
	return fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ProposeSettlement(ctx context.Context, id string, proposal SettlementTerms, userID string) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) RatifySettlement(ctx context.Context, id string, userID string) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) FinalizeSettlement(ctx context.Context, id string, userID string) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) Disburse(ctx context.Context, id string, userID string) error {
	return fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) Cancel(ctx context.Context, id string, userID string) error {
	return fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ExpireEscrow(ctx context.Context, id string, userID string) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ListEscrows(ctx context.Context, userID string) ([]*EscrowContract, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ListProposals(ctx context.Context, userID string) ([]*EscrowProposal, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) GetEscrow(ctx context.Context, id string, userID string) (*EscrowContract, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) GetPartyMap() map[string]string {
	return make(map[string]string)
}

func (c *DamlClient) SetPartyMap(m map[string]string) {}

func (c *DamlClient) CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, asset Asset, terms EscrowTerms) (*EscrowInvitation, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*EscrowProposal, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ListInvitations(ctx context.Context, userID string) ([]*EscrowInvitation, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) GetInvitationByToken(ctx context.Context, tokenHash string) (*EscrowInvitation, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) GetMetrics(ctx context.Context, userID string) (*LedgerMetrics, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ListSettlements(ctx context.Context) ([]*EscrowSettlement, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) SettlePayment(ctx context.Context, settlementID string) error {
	return fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ListWallets(ctx context.Context, userID string) ([]*Wallet, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) GetIdentity(ctx context.Context, oktaSub string) (*UserIdentity, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) ProvisionUser(ctx context.Context, oktaSub string, email string) (*UserIdentity, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) CreateContract(ctx context.Context, userID string, templateID string, payload map[string]interface{}) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) SearchPackageID(ctx context.Context, name string) (string, error) {
	return "", fmt.Errorf("gRPC not implemented")
}

func (c *DamlClient) GetParty(user string) string {
	return user
}

func (c *DamlClient) GetOffset() interface{} {
	return nil
}

func (c *DamlClient) GetInterfacePackageID() string {
	return ""
}

func (c *DamlClient) DoRawRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}

// LEGACY
func (c *DamlClient) CreateEscrow(ctx context.Context, req CreateEscrowRequest) (*EscrowContract, error) {
	return nil, fmt.Errorf("gRPC not implemented")
}
func (c *DamlClient) ReleaseFunds(ctx context.Context, id string, userID string) error { return nil }
func (c *DamlClient) ResolveDispute(ctx context.Context, id string, b, s float64, userID string) error {
	return nil
}
func (c *DamlClient) RefundBuyer(ctx context.Context, id string) error   { return nil }
func (c *DamlClient) RefundBySeller(ctx context.Context, id string) error { return nil }

var _ Client = (*DamlClient)(nil)
