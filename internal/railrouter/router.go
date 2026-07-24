package railrouter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"daml-escrow/internal/ledger"
)

type SettlementRail string

const (
	RailStablecoin SettlementRail = "Stablecoin"
	RailFiat       SettlementRail = "Fiat"
)

type TransferRequest struct {
	EscrowID       string  `json:"escrowId"`
	RecipientEmail string  `json:"recipientEmail"`
	Amount         float64 `json:"amount"`
	Currency       string  `json:"currency"`
	Reference      string  `json:"reference"`
}

type TransferRef string
type TransferStatus string

const (
	StatusPending   TransferStatus = "PENDING"
	StatusCompleted TransferStatus = "COMPLETED"
	StatusFailed    TransferStatus = "FAILED"
)

// FiatProvider defines the swappable interface for fiat payments orchestration.
type FiatProvider interface {
	InitiateTransfer(ctx context.Context, req TransferRequest) (TransferRef, error)
	GetStatus(ctx context.Context, ref TransferRef) (TransferStatus, error)
}

type LedgerClient interface {
	Disburse(ctx context.Context, id string, actAs []string) error
	GetEscrow(ctx context.Context, id string, userID string) (*ledger.EscrowContract, error)
}

type Router struct {
	ledgerClient LedgerClient
	fiatProvider FiatProvider
}

func NewRouter(lc LedgerClient, fp FiatProvider) *Router {
	return &Router{
		ledgerClient: lc,
		fiatProvider: fp,
	}
}

func (r *Router) Route(ctx context.Context, escrowID string, rail SettlementRail, actAs []string, userID string) error {
	if rail == RailStablecoin {
		return r.ledgerClient.Disburse(ctx, escrowID, actAs)
	}

	if rail == RailFiat {
		if r.fiatProvider == nil {
			return errors.New("fiat provider not configured")
		}

		escrow, err := r.ledgerClient.GetEscrow(ctx, escrowID, userID)
		if err != nil {
			return fmt.Errorf("failed to fetch escrow: %w", err)
		}

		// Initiate the fiat transfer
		req := TransferRequest{
			EscrowID:       escrowID,
			RecipientEmail: escrow.Beneficiary,
			Amount:         escrow.Asset.Amount,
			Currency:       escrow.Asset.Currency,
			Reference:      fmt.Sprintf("ESCROW-%s-%d", escrowID[:8], time.Now().Unix()),
		}

		_, err = r.fiatProvider.InitiateTransfer(ctx, req)
		if err != nil {
			return fmt.Errorf("failed to initiate fiat transfer: %w", err)
		}

		return nil
	}

	return fmt.Errorf("unknown settlement rail: %s", rail)
}
