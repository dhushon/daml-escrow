package api

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"daml-escrow/internal/crypto"
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHandler_GetHealth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockLedger := new(ledger.MockLedgerClient)
	mockStablecoin := new(ledger.MockStablecoinProvider)
	compliance := services.NewMockCompliance()
	signer, _ := crypto.NewLocalSigner()
	svc := services.NewEscrowService(logger, mockLedger, mockStablecoin, compliance, "secret", signer, nil, nil)
	metrics := services.NewMetricsService()

	db, _, _ := sqlmock.New()
	defer func() { _ = db.Close() }()
	configSvc := services.NewMockConfigService(db)

	h := NewHandler(logger, svc, metrics, configSvc, nil, nil, nil, nil, nil)

	t.Run("Health returns 200 and UP status", func(t *testing.T) {
		mockLedger.On("SearchPackageID", mock.Anything, "stablecoin-escrow").Return("pkg-123", nil)

		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		rr := httptest.NewRecorder()

		h.GetHealth(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp ledger.HealthResponse
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "UP", resp.Status)
	})
}

func TestHandler_GetIdentity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockLedger := new(ledger.MockLedgerClient)
	mockStablecoin := new(ledger.MockStablecoinProvider)
	compliance := services.NewMockCompliance()
	signer, _ := crypto.NewLocalSigner()
	svc := services.NewEscrowService(logger, mockLedger, mockStablecoin, compliance, "secret", signer, nil, nil)
	
	configContent := `
providers:
  test.com:
    type: OIDC
    issuer: https://oidc.test.com
`
	tmpFile, _ := os.CreateTemp("", "idp*.yaml")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.Write([]byte(configContent))
	_ = tmpFile.Close()

	db, smock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	idSvc, _ := services.NewIdentityService(tmpFile.Name(), db)
	h := NewHandler(logger, svc, nil, nil, nil, idSvc, nil, nil, nil)

	t.Run("Existing Identity", func(t *testing.T) {
		smock.ExpectQuery("SELECT .* FROM identities").
			WithArgs("user-123").
			WillReturnError(sql.ErrNoRows)

		ctx := context.WithValue(context.Background(), AuthSubKey, "user-123")
		ctx = context.WithValue(ctx, EmailKey, "user@test.com")
		req, _ := http.NewRequestWithContext(ctx, "GET", "/api/v1/auth/me", nil)

		mockLedger.On("GetIdentity", mock.Anything, "user-123").Return(&ledger.UserIdentity{
			OktaSub:    "user-123",
			DamlUserID: "user_party",
			DamlPartyID: "p-123",
			Email: "user@test.com",
			Role: "Depositor",
		}, nil)

		smock.ExpectExec("INSERT INTO identities").
			WithArgs("user-123", "user_party", "p-123", "user@test.com", "", "Depositor").
			WillReturnResult(sqlmock.NewResult(1, 1))

		rr := httptest.NewRecorder()
		h.GetIdentity(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockLedger.AssertExpectations(t)
	})
}

func TestHandler_DiscoverAuth(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configContent := `
providers:
  test.com:
    type: OIDC
    issuer: https://oidc.test.com
`
	tmpFile, _ := os.CreateTemp("", "idp*.yaml")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.Write([]byte(configContent))
	_ = tmpFile.Close()

	db2, _, _ := sqlmock.New()
	defer func() { _ = db2.Close() }()
	idSvc, _ := services.NewIdentityService(tmpFile.Name(), db2)
	h := NewHandler(logger, nil, nil, nil, nil, idSvc, nil, nil, nil)

	t.Run("Successful Discovery", func(t *testing.T) {
		body := map[string]string{"email": "user@test.com"}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/auth/discover", bytes.NewBuffer(jsonBody))
		rr := httptest.NewRecorder()

		h.DiscoverAuth(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp services.AuthProvider
		_ = json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.Equal(t, "https://oidc.test.com", resp.Issuer)
	})
}

func TestHandler_OracleMilestoneTrigger(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockLedger := new(ledger.MockLedgerClient)
	mockStablecoin := new(ledger.MockStablecoinProvider)
	compliance := services.NewMockCompliance()
	signer, _ := crypto.NewLocalSigner()
	svc := services.NewEscrowService(logger, mockLedger, mockStablecoin, compliance, "secret", signer, nil, nil)
	h := NewHandler(logger, svc, nil, nil, nil, nil, nil, nil, nil)

	t.Run("Valid HMAC Trigger", func(t *testing.T) {
		body := OracleWebhookRequest{
			EscrowID:       "escrow-123",
			MilestoneIndex: 0,
			Event:          "DELIVERED",
			Signature:      "valid-sig",
			Asymmetric:     false,
		}
		jsonBody, _ := json.Marshal(body)
		req, _ := http.NewRequest("POST", "/api/v1/webhooks/milestone", bytes.NewBuffer(jsonBody))

		mockLedger.On("GetEscrow", mock.Anything, "escrow-123", "CentralBank").Return(&ledger.EscrowContract{
			ID:                    "escrow-123",
			CurrentMilestoneIndex: 0,
			State:                 "ACTIVE",
		}, nil)
		mockLedger.On("Activate", mock.Anything, "escrow-123", []string{"CentralBank"}).Return("escrow-123", nil)

		rr := httptest.NewRecorder()
		h.OracleMilestoneTrigger(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestHandler_WithdrawDraft(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockLedger := new(ledger.MockLedgerClient)
	mockStablecoin := new(ledger.MockStablecoinProvider)
	compliance := services.NewMockCompliance()
	signer, _ := crypto.NewLocalSigner()
	svc := services.NewEscrowService(logger, mockLedger, mockStablecoin, compliance, "secret", signer, nil, nil)

	configContent := "providers:\n  test.com:\n    type: OIDC\n    issuer: https://oidc.test.com\n"
	tmpFile, _ := os.CreateTemp("", "idp*.yaml")
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.Write([]byte(configContent))
	_ = tmpFile.Close()

	db, smock, _ := sqlmock.New()
	defer func() { _ = db.Close() }()

	idSvc, _ := services.NewIdentityService(tmpFile.Name(), db)
	configSvc := services.NewMockConfigService(db)

	h := NewHandler(logger, svc, nil, configSvc, nil, idSvc, nil, nil, nil)

	t.Run("WithdrawDraft - Success", func(t *testing.T) {
		smock.ExpectQuery("SELECT okta_sub, daml_user_id, daml_party_id, email, display_name, role, title, affiliation, organization, physical_address, kyc_status FROM identities").
			WithArgs("user-123").
			WillReturnRows(sqlmock.NewRows([]string{"okta_sub", "daml_user_id", "daml_party_id", "email", "display_name", "role", "title", "affiliation", "organization", "physical_address", "kyc_status"}).
				AddRow("user-123", "user_party", "p-123", "user@test.com", "Depositor", "Depositor", "", "", "", "", "APPROVED"))

		mockRows := sqlmock.NewRows([]string{
			"id", "root_id", "version", "proposer_id", "invitation_code", "contract_type", "initiator_id",
			"initiator_role", "depositor_id", "beneficiary_email", "beneficiary_id", "mediator_id",
			"amount", "currency", "terms", "metadata", "change_summary", "approvals", "status", "created_at",
		}).AddRow(
			"draft-id-123", "root-id-456", 1, "user1", "code", "Corporate", "user1",
			"Depositor", "user1", "user2@email.com", "user2", "mediator",
			100.0, "USD", []byte(`{}`), []byte(`{"ledgerId": "old-ledger-id"}`), "summary", []byte(`["user1"]`), "PROMOTED", time.Now(),
		)
		smock.ExpectQuery("SELECT id, root_id, version, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, change_summary, approvals, status, created_at FROM draft_escrows WHERE root_id = \\$1").
			WithArgs("root-id-456").
			WillReturnRows(mockRows)

		mockLedger.On("WithdrawProposal", mock.Anything, "old-ledger-id", "p-123").Return(nil)

		mockRowsAgain := sqlmock.NewRows([]string{
			"id", "root_id", "version", "proposer_id", "invitation_code", "contract_type", "initiator_id",
			"initiator_role", "depositor_id", "beneficiary_email", "beneficiary_id", "mediator_id",
			"amount", "currency", "terms", "metadata", "change_summary", "approvals", "status", "created_at",
		}).AddRow(
			"draft-id-123", "root-id-456", 1, "user1", "code", "Corporate", "user1",
			"Depositor", "user1", "user2@email.com", "user2", "mediator",
			100.0, "USD", []byte(`{}`), []byte(`{"ledgerId": "old-ledger-id"}`), "summary", []byte(`["user1"]`), "PROMOTED", time.Now(),
		)
		smock.ExpectQuery("SELECT id, root_id, version, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, change_summary, approvals, status, created_at FROM draft_escrows WHERE root_id = \\$1").
			WithArgs("root-id-456").
			WillReturnRows(mockRowsAgain)

		smock.ExpectExec("UPDATE draft_escrows SET status = 'DRAFT'").
			WithArgs(sqlmock.AnyArg(), "draft-id-123").
			WillReturnResult(sqlmock.NewResult(1, 1))

		mockRowsFinal := sqlmock.NewRows([]string{
			"id", "root_id", "version", "proposer_id", "invitation_code", "contract_type", "initiator_id",
			"initiator_role", "depositor_id", "beneficiary_email", "beneficiary_id", "mediator_id",
			"amount", "currency", "terms", "metadata", "change_summary", "approvals", "status", "created_at",
		}).AddRow(
			"draft-id-123", "root-id-456", 1, "user1", "code", "Corporate", "user1",
			"Depositor", "user1", "user2@email.com", "user2", "mediator",
			100.0, "USD", []byte(`{}`), []byte(`{"previousProposalId": "old-ledger-id"}`), "summary", []byte(`[]`), "DRAFT", time.Now(),
		)
		smock.ExpectQuery("SELECT id, root_id, version, proposer_id, invitation_code, contract_type, initiator_id, initiator_role, depositor_id, beneficiary_email, beneficiary_id, mediator_id, amount, currency, terms, metadata, change_summary, approvals, status, created_at FROM draft_escrows WHERE root_id = \\$1").
			WithArgs("root-id-456").
			WillReturnRows(mockRowsFinal)

		ctx := context.WithValue(context.Background(), AuthSubKey, "user-123")
		ctx = context.WithValue(ctx, EmailKey, "user@test.com")
		req, _ := http.NewRequestWithContext(ctx, "POST", "/api/v1/drafts/root-id-456/withdraw", nil)
		
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("draftID", "root-id-456")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		h.WithdrawDraft(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var resp services.DraftEscrow
		err := json.Unmarshal(rr.Body.Bytes(), &resp)
		assert.NoError(t, err)
		assert.Equal(t, "DRAFT", resp.Status)
		mockLedger.AssertExpectations(t)
	})
}
