package api

import (
	"context"
	"daml-escrow/internal/config"
	"daml-escrow/internal/services"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"go.uber.org/zap"
)

type contextKey string

const (
	AuthSubKey      contextKey = "auth_sub"
	EmailKey        contextKey = "email"
	ScopesKey       contextKey = "scopes"
	OriginDomainKey contextKey = "origin_domain"
)

// Permission Scopes
const (
	ScopeEscrowRead   = "escrow:read"
	ScopeEscrowWrite  = "escrow:write"
	ScopeEscrowAccept = "escrow:accept"
	ScopeSystemAdmin  = "system:admin"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			logger.Info(
				"http_request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

func MetricsMiddleware(metrics *services.MetricsService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{w, http.StatusOK}
			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start)
			isError := rw.statusCode >= 400
			metrics.RecordRequest(duration, isError)
		})
	}
}

// TokenVerifier defines the interface for OIDC token verification.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*oidc.IDToken, error)
}

// RealVerifier is the production implementation using go-oidc.
type RealVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func (v *RealVerifier) Verify(ctx context.Context, token string) (*oidc.IDToken, error) {
	return v.verifier.Verify(ctx, token)
}

func NewRealVerifier(ctx context.Context, issuer, audience string) (TokenVerifier, error) {
	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return nil, err
	}
	return &RealVerifier{
		verifier: provider.Verifier(&oidc.Config{ClientID: audience}),
	}, nil
}

func AuthMiddleware(authConfig config.AuthConfig, verifier TokenVerifier, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Handle CORS pre-flight
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			// Check for the new, explicit development bypass first.
			isDevBypass := authConfig.Environment == "dev" && authConfig.AuthBypass

			// Bypass auth for health, swagger, and anonymous token endpoints
			if strings.HasPrefix(r.URL.Path, "/api/v1/health") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/auth/discover") ||
				strings.HasPrefix(r.URL.Path, "/swagger") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/invites/token/") {

				// Special Case: In dev bypass mode, we still want to inject the Dev User
				// if they provided the header, so they can 'Claim' an invitation token.
				if isDevBypass {
					devUser := r.Header.Get("X-Dev-User")
					if devUser != "" {
						scopes := []string{ScopeEscrowRead, ScopeEscrowWrite, ScopeEscrowAccept}
						ctx := context.WithValue(r.Context(), AuthSubKey, devUser)
						ctx = context.WithValue(ctx, EmailKey, devUser+"@dev.local")
						ctx = context.WithValue(ctx, ScopesKey, scopes)
						next.ServeHTTP(w, r.WithContext(ctx))
						return
					}
				}
				next.ServeHTTP(w, r)
				return
			}

			// Namespaced Development Mode Bypass
			if isDevBypass {
				devUser := r.Header.Get("X-Dev-User")
				if devUser == "" {
					devUser = "Buyer"
				}

				// Grant broad scopes for development flexibility
				scopes := []string{ScopeEscrowRead, ScopeEscrowWrite, ScopeEscrowAccept, ScopeSystemAdmin}

				ctx := context.WithValue(r.Context(), AuthSubKey, devUser)
				ctx = context.WithValue(ctx, EmailKey, devUser+"@dev.local")
				ctx = context.WithValue(ctx, ScopesKey, scopes)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr == authHeader {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}

			// Verify the ID token using the injected verifier.
			idToken, err := verifier.Verify(r.Context(), tokenStr)
			if err != nil {
				logger.Warn("failed to verify ID token", zap.Error(err))
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			// Extract claims from the validated token.
			var claims struct {
				Email        string   `json:"email"`
				Scopes       []string `json:"scp"`
				OriginDomain string   `json:"origin_domain"`
			}
			if err := idToken.Claims(&claims); err != nil {
				logger.Error("failed to extract claims from token", zap.Error(err))
				http.Error(w, "invalid claims", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AuthSubKey, idToken.Subject)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, ScopesKey, claims.Scopes)

			// Determine origin domain with fallback for IdPs that don't support custom claims
			originDomain := claims.OriginDomain
			if originDomain == "" && claims.Email != "" {
				parts := strings.Split(claims.Email, "@")
				if len(parts) >= 2 {
					originDomain = parts[len(parts)-1]
				}
			}
			ctx = context.WithValue(ctx, OriginDomainKey, originDomain)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireScope(ctx context.Context, required string) bool {
	scopes, ok := ctx.Value(ScopesKey).([]string)
	if !ok {
		return false
	}
	for _, s := range scopes {
		if s == required {
			return true
		}
	}
	return false
}
