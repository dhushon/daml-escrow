package services

import (
	"context"
	"daml-escrow/internal/ledger"
	"fmt"

	"go.uber.org/zap"
)

type EscrowService struct {
	logger        *zap.Logger
	ledger        ledger.Client
	stablecoin    ledger.StablecoinProvider
	compliance    ComplianceService
	webhookSecret string
}

func NewEscrowService(
	logger *zap.Logger,
	ledger ledger.Client,
	stablecoin ledger.StablecoinProvider,
	compliance ComplianceService,
	webhookSecret string,
) *EscrowService {
	return &EscrowService{
		logger:        logger,
		ledger:        ledger,
		stablecoin:    stablecoin,
		compliance:    compliance,
		webhookSecret: webhookSecret,
	}
}

// ---------------------------------------------------------------------------
// Lifecycle Operations
// ---------------------------------------------------------------------------

func (s *EscrowService) ProposeEscrow(ctx context.Context, req ledger.CreateEscrowRequest) (*ledger.EscrowProposal, error) {
	s.logger.Info("proposing new escrow", zap.String("buyer", req.Buyer), zap.String("seller", req.Seller))
	return s.ledger.ProposeEscrow(ctx, req)
}

func (s *EscrowService) SellerAccept(ctx context.Context, id string, userID string) (string, error) {
	s.logger.Info("seller accepting proposal", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.SellerAccept(ctx, id, userID)
}

func (s *EscrowService) Fund(ctx context.Context, id string, custodyRef string, holdingCid string, userID string) error {
	s.logger.Info("funding escrow", zap.String("id", id), zap.String("ref", custodyRef), zap.String("holdingCid", holdingCid), zap.String("userID", userID))
	return s.ledger.Fund(ctx, id, custodyRef, holdingCid, userID)
}

func (s *EscrowService) Activate(ctx context.Context, id string, actAs []string) (string, error) {
	s.logger.Info("activating escrow", zap.String("id", id), zap.Strings("actAs", actAs))
	return s.ledger.Activate(ctx, id, actAs)
}

func (s *EscrowService) ConfirmConditions(ctx context.Context, id string, userID string) error {
	s.logger.Info("confirming conditions", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.ConfirmConditions(ctx, id, userID)
}

func (s *EscrowService) RaiseDispute(ctx context.Context, id string, userID string) error {
	s.logger.Info("raising dispute", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.RaiseDispute(ctx, id, userID)
}

func (s *EscrowService) ProposeSettlement(ctx context.Context, id string, proposal ledger.SettlementTerms, userID string) (string, error) {
	s.logger.Info("proposing settlement", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.ProposeSettlement(ctx, id, proposal, userID)
}

func (s *EscrowService) RatifySettlement(ctx context.Context, id string, userID string) (string, error) {
	s.logger.Info("ratifying settlement", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.RatifySettlement(ctx, id, userID)
}

func (s *EscrowService) FinalizeSettlement(ctx context.Context, id string, userID string) (string, error) {
	s.logger.Info("finalizing settlement", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.FinalizeSettlement(ctx, id, userID)
}

func (s *EscrowService) Disburse(ctx context.Context, id string, actAs []string) error {
	s.logger.Info("executing disbursement", zap.String("id", id), zap.Strings("actAs", actAs))
	return s.ledger.Disburse(ctx, id, actAs)
}

func (s *EscrowService) Cancel(ctx context.Context, id string, userID string) error {
	s.logger.Info("cancelling escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.Cancel(ctx, id, userID)
}

func (s *EscrowService) ExpireEscrow(ctx context.Context, id string, userID string) (string, error) {
	s.logger.Info("expiring escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.ExpireEscrow(ctx, id, userID)
}

// ---------------------------------------------------------------------------
// Utility and Query methods
// ---------------------------------------------------------------------------

func (s *EscrowService) ListEscrows(ctx context.Context, userID string) ([]*ledger.EscrowContract, error) {
	return s.ledger.ListEscrows(ctx, userID)
}

func (s *EscrowService) GetEscrow(ctx context.Context, id string, userID string) (*ledger.EscrowContract, error) {
	return s.ledger.GetEscrow(ctx, id, userID)
}

func (s *EscrowService) GetLedgerClient() ledger.Client {
	return s.ledger
}

func (s *EscrowService) GetOracleSecret() string {
	return s.webhookSecret
}

func (s *EscrowService) ProvisionUser(ctx context.Context, oktaSub string, email string, scopes []string) (*ledger.UserIdentity, error) {
	return s.ledger.ProvisionUser(ctx, oktaSub, email, scopes)
}

func (s *EscrowService) ListIdentities(ctx context.Context) ([]*ledger.UserIdentity, error) {
	return s.ledger.ListIdentities(ctx)
}

func (s *EscrowService) GetMetrics(ctx context.Context, userID string) (*ledger.LedgerMetrics, error) {
	return s.ledger.GetMetrics(ctx, userID)
}

func (s *EscrowService) ListWallets(ctx context.Context, userID string) ([]*ledger.Wallet, error) {
	return s.ledger.ListWallets(ctx, userID)
}

func (s *EscrowService) GetIdentity(ctx context.Context, oktaSub string) (*ledger.UserIdentity, error) {
	return s.ledger.GetIdentity(ctx, oktaSub)
}

func (s *EscrowService) CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, asset ledger.Asset, terms ledger.EscrowTerms) (*ledger.EscrowInvitation, error) {
	return s.ledger.CreateInvitation(ctx, inviterID, inviteeEmail, role, inviteeType, asset, terms)
}

func (s *EscrowService) ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*ledger.EscrowProposal, error) {
	return s.ledger.ClaimInvitation(ctx, inviteID, claimantID)
}

func (s *EscrowService) ListInvitations(ctx context.Context, userID string) ([]*ledger.EscrowInvitation, error) {
	return s.ledger.ListInvitations(ctx, userID)
}

func (s *EscrowService) GetInvitationByToken(ctx context.Context, tokenHash string) (*ledger.EscrowInvitation, error) {
	return s.ledger.GetInvitationByToken(ctx, tokenHash)
}

func (s *EscrowService) ListSettlements(ctx context.Context) ([]*ledger.EscrowSettlement, error) {
	return s.ledger.ListSettlements(ctx)
}

func (s *EscrowService) SettlePayment(ctx context.Context, settlementID string) error {
	return s.ledger.SettlePayment(ctx, settlementID)
}

func (s *EscrowService) OracleMilestoneTrigger(ctx context.Context, escrowID string, milestoneIndex int, event string, signature string) error {
	// Verify signature
	if !s.compliance.VerifyOracleSignature(escrowID, milestoneIndex, event, signature, s.webhookSecret) {
		return fmt.Errorf("invalid oracle signature")
	}

	// Fetch current state
	escrow, err := s.ledger.GetEscrow(ctx, escrowID, "CentralBank") // Internal lookup
	if err != nil {
		return fmt.Errorf("escrow not found: %w", err)
	}

	if escrow.CurrentMilestoneIndex != milestoneIndex {
		return fmt.Errorf("milestone index mismatch: expected %d, got %d", escrow.CurrentMilestoneIndex, milestoneIndex)
	}

	// Execute authoritative transition (Internal admin action)
	// In institutional models, the Oracle trigger might drive a co-signed disbursement.
	_, err = s.ledger.Activate(ctx, escrowID, []string{"CentralBank"}) 
	return err
}
