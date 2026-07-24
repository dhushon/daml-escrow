package railrouter

import (
	"context"
	"errors"
	"testing"

	"daml-escrow/internal/ledger"
)

type mockFiatProvider struct {
	initCalled bool
	reqSent    TransferRequest
	errToReturn error
}

func (m *mockFiatProvider) InitiateTransfer(ctx context.Context, req TransferRequest) (TransferRef, error) {
	m.initCalled = true
	m.reqSent = req
	return "REF-123", m.errToReturn
}

func (m *mockFiatProvider) GetStatus(ctx context.Context, ref TransferRef) (TransferStatus, error) {
	return StatusCompleted, nil
}

type mockLedgerClient struct {
	disburseCalled bool
	escrowToReturn *ledger.EscrowContract
	errToReturn    error
}

func (m *mockLedgerClient) Disburse(ctx context.Context, id string, actAs []string) error {
	m.disburseCalled = true
	return m.errToReturn
}

func (m *mockLedgerClient) GetEscrow(ctx context.Context, id string, userID string) (*ledger.EscrowContract, error) {
	if m.escrowToReturn != nil {
		return m.escrowToReturn, nil
	}
	return nil, errors.New("not found")
}

func TestRouteStablecoin(t *testing.T) {
	lc := &mockLedgerClient{}
	fp := &mockFiatProvider{}
	r := NewRouter(lc, fp)

	err := r.Route(context.Background(), "escrow-123", RailStablecoin, []string{"Issuer"}, "Issuer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !lc.disburseCalled {
		t.Error("expected disburse to be called on ledger client")
	}
	if fp.initCalled {
		t.Error("did not expect fiat provider to be invoked for stablecoin")
	}
}

func TestRouteFiat(t *testing.T) {
	escrow := &ledger.EscrowContract{
		ID:          "escrow-123",
		Beneficiary: "beneficiary@vdatacloud.com",
		Asset: ledger.Asset{
			Amount:   5000.0,
			Currency: "USD",
		},
	}
	lc := &mockLedgerClient{escrowToReturn: escrow}
	fp := &mockFiatProvider{}
	r := NewRouter(lc, fp)

	err := r.Route(context.Background(), "escrow-123", RailFiat, []string{"Issuer"}, "Issuer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lc.disburseCalled {
		t.Error("did not expect disburse choice to be exercised immediately for fiat")
	}
	if !fp.initCalled {
		t.Error("expected fiat provider to be invoked")
	}
	if fp.reqSent.Amount != 5000.0 || fp.reqSent.RecipientEmail != "beneficiary@vdatacloud.com" {
		t.Errorf("incorrect transfer details: %+v", fp.reqSent)
	}
}
