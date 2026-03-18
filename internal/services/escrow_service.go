package services

import (
	"context"

	"daml-escrow/internal/ledger"

	"go.uber.org/zap"
)

type EscrowService struct {
	logger *zap.Logger
	ledger ledger.Client
}

func NewEscrowService(
	logger *zap.Logger,
	ledger ledger.Client,
) *EscrowService {

	return &EscrowService{
		logger: logger,
		ledger: ledger,
	}
}

func (s *EscrowService) CreateEscrow(
	ctx context.Context,
	req ledger.CreateEscrowRequest,
) (*ledger.EscrowContract, error) {
	s.logger.Info("creating escrow", zap.Any("request", req))
	return s.ledger.CreateEscrow(ctx, req)
}

func (s *EscrowService) GetEscrow(
	ctx context.Context,
	id string,
) (*ledger.EscrowContract, error) {
	s.logger.Info("getting escrow", zap.String("id", id))
	return s.ledger.GetEscrow(ctx, id)
}

func (s *EscrowService) ReleaseFunds(
	ctx context.Context,
	id string,
) error {
	s.logger.Info("releasing funds", zap.String("id", id))
	return s.ledger.ReleaseFunds(ctx, id)
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
