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
	webhookSecret string
}

func NewEscrowService(
	logger *zap.Logger,
	ledger ledger.Client,
	webhookSecret string,
) *EscrowService {

	return &EscrowService{
		logger:        logger,
		ledger:        ledger,
		webhookSecret: webhookSecret,
	}
}

func (s *EscrowService) CreateEscrow(
	ctx context.Context,
	req ledger.CreateEscrowRequest,
) (*ledger.EscrowContract, error) {
	s.logger.Info("creating escrow", zap.Any("request", req))
	return s.ledger.CreateEscrow(ctx, req)
}

func (s *EscrowService) ProposeEscrow(
	ctx context.Context,
	req ledger.CreateEscrowRequest,
) (*ledger.EscrowProposal, error) {
	s.logger.Info("proposing escrow", zap.Any("request", req))
	return s.ledger.ProposeEscrow(ctx, req)
}

func (s *EscrowService) AcceptProposal(
	ctx context.Context,
	id string,
	sellerID string,
) error {
	s.logger.Info("accepting proposal", zap.String("id", id), zap.String("seller", sellerID))
	return s.ledger.AcceptProposal(ctx, id, sellerID)
}

func (s *EscrowService) ListProposals(
	ctx context.Context,
	userID string,
) ([]*ledger.EscrowProposal, error) {
	s.logger.Info("listing proposals for user", zap.String("userID", userID))
	return s.ledger.ListProposals(ctx, userID)
}

func (s *EscrowService) GetEscrow(
	ctx context.Context,
	id string,
	userID string,
) (*ledger.EscrowContract, error) {
	s.logger.Info("getting escrow", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.GetEscrow(ctx, id, userID)
}

func (s *EscrowService) ListEscrows(
	ctx context.Context,
	userID string,
) ([]*ledger.EscrowContract, error) {
	s.logger.Info("listing escrows for user", zap.String("userID", userID))
	return s.ledger.ListEscrows(ctx, userID)
}

func (s *EscrowService) ReleaseFunds(
	ctx context.Context,
	id string,
	userID string,
) error {
	s.logger.Info("releasing funds", zap.String("id", id), zap.String("userID", userID))
	return s.ledger.ReleaseFunds(ctx, id, userID)
}

func (s *EscrowService) RaiseDispute(
	ctx context.Context,
	id string,
) (string, error) {
	s.logger.Info("raising dispute", zap.String("id", id))
	return s.ledger.RaiseDispute(ctx, id)
}

func (s *EscrowService) ResolveDispute(
	ctx context.Context,
	id string,
	payoutToBuyer, payoutToSeller float64,
) error {
	s.logger.Info("resolving dispute", zap.String("id", id), zap.Float64("payoutToBuyer", payoutToBuyer), zap.Float64("payoutToSeller", payoutToSeller))
	return s.ledger.ResolveDispute(ctx, id, payoutToBuyer, payoutToSeller)
}

func (s *EscrowService) RefundBuyer(
	ctx context.Context,
	id string,
) error {
	s.logger.Info("refunding buyer", zap.String("id", id))
	// Use Mediator role to ensure visibility for refund processing
	escrow, err := s.ledger.GetEscrow(ctx, id, ledger.EscrowMediatorUser)
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
		return fmt.Errorf("no funds to refund")
	}

	disputeID, err := s.RaiseDispute(ctx, id)
	if err != nil {
		return err
	}

	return s.ResolveDispute(ctx, disputeID, remaining, 0.0)
}

func (s *EscrowService) RefundBySeller(
	ctx context.Context,
	id string,
) error {
	s.logger.Info("seller refunding buyer", zap.String("id", id))
	return s.ledger.RefundBySeller(ctx, id)
}

func (s *EscrowService) CreateInvitation(
	ctx context.Context,
	inviterID string,
	inviteeEmail string,
	role string,
	inviteeType string,
	terms ledger.EscrowTerms,
) (*ledger.EscrowInvitation, error) {
	s.logger.Info("creating invitation", zap.String("inviter", inviterID), zap.String("invitee", inviteeEmail))
	return s.ledger.CreateInvitation(ctx, inviterID, inviteeEmail, role, inviteeType, terms)
}

func (s *EscrowService) ClaimInvitation(
	ctx context.Context,
	inviteID string,
	claimantID string,
) (*ledger.EscrowProposal, error) {
	s.logger.Info("claiming invitation", zap.String("id", inviteID), zap.String("claimant", claimantID))
	return s.ledger.ClaimInvitation(ctx, inviteID, claimantID)
}

func (s *EscrowService) ListInvitations(
	ctx context.Context,
	userID string,
) ([]*ledger.EscrowInvitation, error) {
	s.logger.Info("listing invitations for user", zap.String("userID", userID))
	return s.ledger.ListInvitations(ctx, userID)
}

func (s *EscrowService) GetInvitationByToken(
	ctx context.Context,
	tokenHash string,
) (*ledger.EscrowInvitation, error) {
	s.logger.Info("fetching invitation by token", zap.String("token", tokenHash))
	return s.ledger.GetInvitationByToken(ctx, tokenHash)
}

func (s *EscrowService) ListSettlements(
	ctx context.Context,
) ([]*ledger.EscrowSettlement, error) {
	s.logger.Info("listing settlements")
	return s.ledger.ListSettlements(ctx)
}

func (s *EscrowService) SettlePayment(
	ctx context.Context,
	settlementID string,
) error {
	s.logger.Info("settling payment", zap.String("settlementID", settlementID))
	return s.ledger.SettlePayment(ctx, settlementID)
}

func (s *EscrowService) GetMetrics(
	ctx context.Context,
	userID string,
) (*ledger.LedgerMetrics, error) {
	s.logger.Info("getting metrics for user", zap.String("userID", userID))
	return s.ledger.GetMetrics(ctx, userID)
}

func (s *EscrowService) ListWallets(
	ctx context.Context,
	userID string,
) ([]*ledger.Wallet, error) {
	s.logger.Info("listing wallets for user", zap.String("userID", userID))
	return s.ledger.ListWallets(ctx, userID)
}

func (s *EscrowService) ProcessOracleWebhook(
	ctx context.Context,
	req ledger.OracleWebhookRequest,
) error {
	s.logger.Info("processing oracle webhook", 
		zap.String("escrowId", req.EscrowID),
		zap.Int("milestoneIndex", req.MilestoneIndex),
		zap.String("event", req.Event),
		zap.String("provider", req.OracleProvider))

	// 1. Verify Signature
	if err := s.verifySignature(req); err != nil {
		s.logger.Error("webhook signature verification failed", zap.Error(err))
		return fmt.Errorf("unauthorized: %w", err)
	}

	// 2. Fetch current contract state using Mediator role for full visibility
	escrow, err := s.ledger.GetEscrow(ctx, req.EscrowID, ledger.EscrowMediatorUser)
	if err != nil {
		s.logger.Error("failed to fetch escrow for webhook processing", zap.Error(err))
		return fmt.Errorf("escrow not found: %w", err)
	}

	// 3. Guards
	if escrow.State == "Disputed" {
		return fmt.Errorf("cannot automate approval for disputed escrow %s", req.EscrowID)
	}

	if req.MilestoneIndex != escrow.CurrentMilestoneIndex {
		return fmt.Errorf("webhook milestone index %d does not match current escrow milestone index %d", 
			req.MilestoneIndex, escrow.CurrentMilestoneIndex)
	}

	// 4. Execute Automated Choice
	s.logger.Info("oracle automation triggered milestone approval", 
		zap.String("escrowId", req.EscrowID), 
		zap.Int("index", req.MilestoneIndex))
		
	return s.ledger.ReleaseFunds(ctx, req.EscrowID, ledger.EscrowMediatorUser)
}

func (s *EscrowService) GetIdentity(
	ctx context.Context,
	oktaSub string,
) (*ledger.UserIdentity, error) {
	s.logger.Info("getting identity", zap.String("sub", oktaSub))
	return s.ledger.GetIdentity(ctx, oktaSub)
}

func (s *EscrowService) ProvisionUser(
	ctx context.Context,
	oktaSub string,
	email string,
) (*ledger.UserIdentity, error) {
	s.logger.Info("provisioning user", zap.String("sub", oktaSub), zap.String("email", email))
	return s.ledger.ProvisionUser(ctx, oktaSub, email)
}

func (s *EscrowService) verifySignature(req ledger.OracleWebhookRequest) error {
	if s.webhookSecret == "" {
		s.logger.Warn("webhook secret not configured, skipping verification")
		return nil
	}

	if req.Signature == "" {
		return fmt.Errorf("missing signature")
	}

	// Payload for signature: escrowId|milestoneIndex|event|oracleProvider
	payload := fmt.Sprintf("%s|%d|%s|%s", req.EscrowID, req.MilestoneIndex, req.Event, req.OracleProvider)
	
	h := hmac.New(sha256.New, []byte(s.webhookSecret))
	h.Write([]byte(payload))
	expected := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(req.Signature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}
