package api

import (
	"context"
	"daml-escrow/internal/ledger"
	"daml-escrow/internal/services"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestHandler_GetHealth(t *testing.T) {
	// 1. Setup mock dependencies
	logger := zap.NewNop()
	metrics := services.NewMetricsService()
	
	db, mockDB, _ := sqlmock.New(sqlmock.MonitorPingsOption(true))
	defer db.Close()
	configSvc := services.NewMockConfigService(db)
	mockDB.ExpectPing()

	mockLedger := new(ledger.MockLedgerClient)
	mockLedger.On("SearchPackageID", mock.Anything, "stablecoin-escrow").Return("pkg-123", nil)

	escrowSvc := services.NewEscrowService(logger, mockLedger, nil, nil, "test-secret")
	
	h := NewHandler(logger, escrowSvc, metrics, configSvc, nil, nil)

	t.Run("Health returns 200 and UP status", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/health", nil)
		rr := httptest.NewRecorder()
		
		h.GetHealth(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "\"status\":\"UP\"")
	})
}

func TestHandler_GetIdentity(t *testing.T) {
	logger := zap.NewNop()
	mockLedger := new(ledger.MockLedgerClient)
	escrowSvc := services.NewEscrowService(logger, mockLedger, nil, nil, "test-secret")
	h := NewHandler(logger, escrowSvc, nil, nil, nil, nil)

	t.Run("Existing Identity", func(t *testing.T) {
		sub := "user-123"
		mockLedger.On("GetIdentity", mock.Anything, sub).Return(&ledger.UserIdentity{OktaSub: sub, Email: "test@test.com"}, nil)

		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		ctx := context.WithValue(req.Context(), AuthSubKey, sub)
		rr := httptest.NewRecorder()

		h.GetIdentity(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusOK, rr.Code)
		mockLedger.AssertExpectations(t)
	})

	t.Run("JIT Provisioning", func(t *testing.T) {
		sub := "new-user"
		email := "new@test.com"
		mockLedger.On("GetIdentity", mock.Anything, sub).Return(nil, nil)
		mockLedger.On("ProvisionUser", mock.Anything, sub, email, []string{"scope1"}).Return(&ledger.UserIdentity{OktaSub: sub, Email: email}, nil)

		req, _ := http.NewRequest("GET", "/api/v1/auth/me", nil)
		ctx := context.WithValue(req.Context(), AuthSubKey, sub)
		ctx = context.WithValue(ctx, EmailKey, email)
		ctx = context.WithValue(ctx, ScopesKey, []string{"scope1"})
		rr := httptest.NewRecorder()

		h.GetIdentity(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusOK, rr.Code)
		mockLedger.AssertExpectations(t)
	})
}

func TestHandler_DiscoverAuth(t *testing.T) {
	logger := zap.NewNop()
	// Mock IdentityService
	configContent := `
providers:
  test.com:
    type: OIDC
    issuer: https://oidc.test.com
`
	tmpFile, _ := os.CreateTemp("", "idp*.yaml")
	defer os.Remove(tmpFile.Name())
	tmpFile.Write([]byte(configContent))
	tmpFile.Close()

	idSvc, _ := services.NewIdentityService(tmpFile.Name())
	h := NewHandler(logger, nil, nil, nil, nil, idSvc)

	t.Run("Successful Discovery", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/auth/discover?email=user@test.com", nil)
		rr := httptest.NewRecorder()

		h.DiscoverAuth(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "https://oidc.test.com")
	})
}

func TestHandler_ClaimInvitation(t *testing.T) {
	logger := zap.NewNop()
	mockLedger := new(ledger.MockLedgerClient)
	escrowSvc := services.NewEscrowService(logger, mockLedger, nil, nil, "test-secret")
	h := NewHandler(logger, escrowSvc, nil, nil, nil, nil)

	token := "valid-token"
	inviteeEmail := "invitee@test.com"
	invite := &ledger.EscrowInvitation{ID: "invite-123", InviteeEmail: inviteeEmail}

	t.Run("Authorized Claim", func(t *testing.T) {
		mockLedger.On("GetInvitationByToken", mock.Anything, token).Return(invite, nil)
		mockLedger.On("ClaimInvitation", mock.Anything, invite.ID, "user-123").Return(&ledger.EscrowProposal{ID: "prop-123"}, nil)

		req, _ := http.NewRequest("POST", "/api/v1/invites/token/"+token+"/claim", nil)
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext())
		routeCtx := ctx.Value(chi.RouteCtxKey).(*chi.Context)
		routeCtx.URLParams.Add("token", token)
		
		ctx = context.WithValue(ctx, AuthSubKey, "user-123")
		ctx = context.WithValue(ctx, EmailKey, inviteeEmail)
		rr := httptest.NewRecorder()

		h.ClaimInvitation(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusOK, rr.Code)
		mockLedger.AssertExpectations(t)
	})

	t.Run("Unauthorized Claim - Email Mismatch", func(t *testing.T) {
		mockLedger.On("GetInvitationByToken", mock.Anything, token).Return(invite, nil)

		req, _ := http.NewRequest("POST", "/api/v1/invites/token/"+token+"/claim", nil)
		ctx := context.WithValue(req.Context(), chi.RouteCtxKey, chi.NewRouteContext())
		routeCtx := ctx.Value(chi.RouteCtxKey).(*chi.Context)
		routeCtx.URLParams.Add("token", token)
		
		ctx = context.WithValue(ctx, AuthSubKey, "attacker-123")
		ctx = context.WithValue(ctx, EmailKey, "attacker@test.com")
		rr := httptest.NewRecorder()

		h.ClaimInvitation(rr, req.WithContext(ctx))

		assert.Equal(t, http.StatusForbidden, rr.Code)
		assert.Contains(t, rr.Body.String(), "unauthorized")
	})
}

func (h *Handler) TestRoute(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "escrowID")
	h.renderJSON(w, map[string]string{"id": id, "status": "mocked"})
}
