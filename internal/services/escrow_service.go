package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"daml-escrow/internal/ledger"

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
// Lifecycle Management (Directive 05)
// ---------------------------------------------------------------------------

func (s *EscrowService) ProposeEscrow(ctx context.Context, req ledger.CreateEscrowRequest) (*ledger.EscrowProposal, error) {
	s.logger.Info("proposing escrow", zap.Any("request", req))
	return s.ledger.ProposeEscrow(ctx, req)
}

func (s *EscrowService) Fund(ctx context.Context, id string, custodyRef string, holdingCid string, userID string) error {
	s.logger.Info("funding escrow", zap.String("id", id), zap.String("ref", custodyRef), zap.String("holdingCid", holdingCid), zap.String("userID", userID))
	return s.ledger.Fund(ctx, id, custodyRef, holdingCid, userID)
}

func (s *EscrowService) Activate(ctx context.Context, id string, userID string) (string, error) {
	s.logger.Info("activating escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.Activate(ctx, id, userID)
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


func (s *EscrowService) Disburse(ctx context.Context, id string, userID string) error {
	s.logger.Info("executing disbursement", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.Disburse(ctx, id, userID)
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
// Queries
// ---------------------------------------------------------------------------

func (s *EscrowService) GetEscrow(ctx context.Context, id string, userID string) (*ledger.EscrowContract, error) {
	return s.ledger.GetEscrow(ctx, id, userID)
}

func (s *EscrowService) ListEscrows(ctx context.Context, userID string) ([]*ledger.EscrowContract, error) {
	return s.ledger.ListEscrows(ctx, userID)
}

func (s *EscrowService) ListProposals(ctx context.Context, userID string) ([]*ledger.EscrowProposal, error) {
	return s.ledger.ListProposals(ctx, userID)
}

// ---------------------------------------------------------------------------
// Invitation Handlers
// ---------------------------------------------------------------------------

func (s *EscrowService) CreateInvitation(ctx context.Context, inviterID string, inviteeEmail string, role string, inviteeType string, asset ledger.Asset, terms ledger.EscrowTerms) (*ledger.EscrowInvitation, error) {
	s.logger.Info("creating invitation", zap.String("inviter", inviterID), zap.String("invitee", inviteeEmail))
	return s.ledger.CreateInvitation(ctx, inviterID, inviteeEmail, role, inviteeType, asset, terms)
}

func (s *EscrowService) ClaimInvitation(ctx context.Context, inviteID string, claimantID string) (*ledger.EscrowProposal, error) {
	s.logger.Info("claiming invitation", zap.String("id", inviteID), zap.String("claimant", claimantID))
	return s.ledger.ClaimInvitation(ctx, inviteID, claimantID)
}

func (s *EscrowService) ListInvitations(ctx context.Context, userID string) ([]*ledger.EscrowInvitation, error) {
	s.logger.Info("listing invitations for user", zap.String("userID", userID))
	return s.ledger.ListInvitations(ctx, userID)
}

func (s *EscrowService) GetInvitationByToken(ctx context.Context, tokenHash string) (*ledger.EscrowInvitation, error) {
	s.logger.Info("fetching invitation by token", zap.String("token", tokenHash))
	return s.ledger.GetInvitationByToken(ctx, tokenHash)
}

// ---------------------------------------------------------------------------
// Automation & Webhooks
// ---------------------------------------------------------------------------

func (s *EscrowService) ProcessOracleWebhook(ctx context.Context, req ledger.OracleWebhookRequest) error {
	s.logger.Info("processing oracle webhook", zap.String("escrowId", req.EscrowID), zap.String("event", req.Event))

	if err := s.verifySignature(req); err != nil {
		return fmt.Errorf("unauthorized: %w", err)
	}

	escrow, err := s.ledger.GetEscrow(ctx, req.EscrowID, ledger.EscrowMediatorUser)
	if err != nil {
		return err
	}

	if escrow.State != "ACTIVE" {
		return fmt.Errorf("escrow %s is not in ACTIVE state", req.EscrowID)
	}

	if escrow.CurrentMilestoneIndex != req.MilestoneIndex {
		return fmt.Errorf("webhook milestone index %d does not match current escrow milestone index %d", req.MilestoneIndex, escrow.CurrentMilestoneIndex)
	}

	return s.ledger.ConfirmConditions(ctx, req.EscrowID, ledger.EscrowMediatorUser)
}

func (s *EscrowService) verifySignature(req ledger.OracleWebhookRequest) error {
	if s.webhookSecret == "" {
		return nil
	}
	payload := fmt.Sprintf("%s|%d|%s|%s", req.EscrowID, req.MilestoneIndex, req.Event, req.OracleProvider)
	h := hmac.New(sha256.New, []byte(s.webhookSecret))
	h.Write([]byte(payload))
	expected := hex.EncodeToString(h.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(req.Signature)) {
		return fmt.Errorf("invalid signature")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Identity & Metrics
// ---------------------------------------------------------------------------

func (s *EscrowService) GetIdentity(ctx context.Context, oktaSub string) (*ledger.UserIdentity, error) {
	return s.ledger.GetIdentity(ctx, oktaSub)
}

func (s *EscrowService) ProvisionUser(ctx context.Context, oktaSub string, email string, scopes []string) (*ledger.UserIdentity, error) {
	return s.ledger.ProvisionUser(ctx, oktaSub, email, scopes)
}

func (s *EscrowService) GetMetrics(ctx context.Context, userID string) (*ledger.LedgerMetrics, error) {
	return s.ledger.GetMetrics(ctx, userID)
}

func (s *EscrowService) ListWallets(ctx context.Context, userID string) ([]*ledger.Wallet, error) {
	return s.ledger.ListWallets(ctx, userID)
}

func (s *EscrowService) ListSettlements(ctx context.Context) ([]*ledger.EscrowSettlement, error) {
	return s.ledger.ListSettlements(ctx)
}

func (s *EscrowService) SettlePayment(ctx context.Context, settlementID string) error {
	return s.ledger.SettlePayment(ctx, settlementID)
}

func (s *EscrowService) GetLedgerClient() ledger.Client {
	return s.ledger
}

func (s *EscrowService) GetOracleSecret() string {
	return s.webhookSecret
}
