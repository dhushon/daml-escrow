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

	"daml-escrow/internal/crypto"
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"

	"github.com/DATA-DOG/go-sqlmock"
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
	svc := services.NewEscrowService(logger, mockLedger, mockStablecoin, compliance, "secret", signer)
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
	svc := services.NewEscrowService(logger, mockLedger, mockStablecoin, compliance, "secret", signer)
	
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
		req, _ := http.NewRequest("GET", "/api/v1/auth/discover?email=user@test.com", nil)
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
	svc := services.NewEscrowService(logger, mockLedger, mockStablecoin, compliance, "secret", signer)
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
