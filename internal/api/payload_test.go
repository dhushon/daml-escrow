package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestPayloadGeneration_DryRun(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	require := assert.New(t)

	db, smock, err := sqlmock.New()
	require.NoError(err)
	defer func() { _ = db.Close() }()

	// 1. Setup Services & Handler (Mocks)
	mockLedger := new(ledger.MockLedgerClient)
	escrowSvc := services.NewEscrowService(logger, mockLedger, nil, nil, "secret", nil, nil)
	
	// Write a mock yaml config for IdentityService
	configContent := `
providers:
  wallet.devlocal:
    type: OIDC
`
	tmpFile, err := os.CreateTemp("", "idp*.yaml")
	require.NoError(err)
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.Write([]byte(configContent))
	_ = tmpFile.Close()

	idSvc, err := services.NewIdentityService(tmpFile.Name(), db)
	require.NoError(err)
	
	h := NewHandler(logger, escrowSvc, nil, nil, nil, idSvc, nil, nil, nil)

	t.Run("POST /escrows/propose - Wallet Dry Run", func(t *testing.T) {
		reqBody := ProposeEscrowRequest{
			Beneficiary: "u-beneficiary-test",
			Mediator:    "u-mediator-test",
			Amount:      15000.0,
			Currency:    "USD",
		}
		body, _ := json.Marshal(reqBody)

		req, _ := http.NewRequest("POST", "/api/v1/escrows/propose", bytes.NewBuffer(body))
		
		// Inject wallet authentication context
		ctx := context.WithValue(req.Context(), AuthMethodKey, "wallet")
		ctx = context.WithValue(ctx, AuthSubKey, "wallet:u-depositor-test::123")
		ctx = context.WithValue(ctx, EmailKey, "depositor@wallet.devlocal")
		
		// Mock the Postgres SELECT for IdentityService
		rows := sqlmock.NewRows([]string{
			"okta_sub", "daml_user_id", "daml_party_id", "email", "display_name", "role",
			"title", "affiliation", "organization", "physical_address", "kyc_status",
		}).AddRow(
			"wallet:u-depositor-test::123", "u-depositor-test", "u-depositor-test::123", "depositor@wallet.devlocal",
			"Depositor", "Depositor", "", "", "", "", "VERIFIED",
		)

		smock.ExpectQuery("SELECT .* FROM identities WHERE okta_sub =").
			WithArgs("wallet:u-depositor-test::123").
			WillReturnRows(rows)

		rr := httptest.NewRecorder()
		h.ProposeEscrow(rr, req.WithContext(ctx))

		require.Equal(http.StatusOK, rr.Code)

		var resp DryRunResponse
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(err)
		require.True(resp.IsDryRun)
		require.Len(resp.Commands, 1)

		cmd := resp.Commands[0]
		require.Equal("create", cmd.CommandType)
		require.Equal("StablecoinEscrow:EscrowProposal", cmd.TemplateID)
		require.Equal("u-depositor-test", cmd.Argument["depositor"])
		require.Equal("u-beneficiary-test", cmd.Argument["beneficiary"])
		require.Equal(15000.0, cmd.Argument["amount"])
	})

	t.Run("POST /escrows/{escrowID}/activate - Wallet Dry Run", func(t *testing.T) {
		escrowID := "escrow-999"
		
		// Define Chi Router Context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("escrowID", escrowID)

		req, _ := http.NewRequest("POST", "/api/v1/escrows/"+escrowID+"/activate", nil)
		
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, rctx)
		ctx = context.WithValue(ctx, AuthMethodKey, "wallet")
		ctx = context.WithValue(ctx, AuthSubKey, "wallet:Depositor")

		rr := httptest.NewRecorder()
		h.Activate(rr, req.WithContext(ctx))

		require.Equal(http.StatusOK, rr.Code)

		var resp DryRunResponse
		err = json.Unmarshal(rr.Body.Bytes(), &resp)
		require.NoError(err)
		require.True(resp.IsDryRun)
		require.Len(resp.Commands, 1)

		cmd := resp.Commands[0]
		require.Equal("exercise", cmd.CommandType)
		require.Equal("StablecoinEscrow:Escrow", cmd.TemplateID)
		require.Equal(escrowID, cmd.ContractID)
		require.Equal("activate", cmd.Choice)
	})
}
