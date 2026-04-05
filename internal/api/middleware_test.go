package api

import (
	"context"
	"daml-escrow/internal/config"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockVerifier for testing
type MockVerifier struct {
	mock.Mock
}

func (m *MockVerifier) Verify(ctx context.Context, token string) (*oidc.IDToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*oidc.IDToken), args.Error(1)
}

func TestAuthMiddleware_Bypass(t *testing.T) {
	logger := zap.NewNop()
	
	t.Run("Bypass enabled in dev", func(t *testing.T) {
		authCfg := config.AuthConfig{
			Environment: "dev",
			AuthBypass:  true,
		}
		
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sub := r.Context().Value(AuthSubKey)
			email := r.Context().Value(EmailKey)
			assert.NotNil(t, sub)
			assert.NotNil(t, email)
			w.WriteHeader(http.StatusOK)
		})

		middleware := AuthMiddleware(authCfg, nil, logger)
		h := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/api/v1/escrows", nil)
		req.Header.Set("X-Dev-User", "TestUser")
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Bypass disabled in production", func(t *testing.T) {
		authCfg := config.AuthConfig{
			Environment: "production",
			AuthBypass:  true, // Should be ignored in production
		}
		
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := AuthMiddleware(authCfg, nil, logger)
		h := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/api/v1/escrows", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		// Should fail because verifier is nil and bypass is ignored
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("OPTIONS bypass", func(t *testing.T) {
		authCfg := config.AuthConfig{Environment: "production"}
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// sub and email are NOT set for OPTIONS
			w.WriteHeader(http.StatusOK)
		})

		middleware := AuthMiddleware(authCfg, nil, logger)
		h := middleware(nextHandler)

		req := httptest.NewRequest("OPTIONS", "/api/v1/escrows", nil)
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("Verification Failure", func(t *testing.T) {
		authCfg := config.AuthConfig{Environment: "production"}
		mockVer := new(MockVerifier)
		mockVer.On("Verify", mock.Anything, "bad-token").Return(nil, fmt.Errorf("invalid signature"))
		
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("nextHandler should not be called")
		})

		middleware := AuthMiddleware(authCfg, mockVer, logger)
		h := middleware(nextHandler)

		req := httptest.NewRequest("GET", "/api/v1/escrows", nil)
		req.Header.Set("Authorization", "Bearer bad-token")
		rr := httptest.NewRecorder()

		h.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		mockVer.AssertExpectations(t)
	})
}

func TestAuthMiddleware_PublicEndpoints(t *testing.T) {
	logger := zap.NewNop()
	authCfg := config.AuthConfig{Environment: "production"}
	
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := AuthMiddleware(authCfg, nil, logger)
	h := middleware(nextHandler)

	publicPaths := []string{
		"/api/v1/health",
		"/api/v1/auth/discover",
		"/swagger/index.html",
		"/api/v1/invites/token/abc-123",
	}

	for _, path := range publicPaths {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest("GET", path, nil)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}
