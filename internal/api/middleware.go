package api

import (
	"context"
	"daml-escrow/internal/config"
	"daml-escrow/internal/services"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type contextKey string

const (
	AuthSubKey      contextKey = "auth_sub"
	EmailKey        contextKey = "email"
	ScopesKey       contextKey = "scopes"
	OriginDomainKey contextKey = "origin_domain"
	AuthMethodKey   contextKey = "auth_method"
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

			// High-Assurance: Extract Account and Contract IDs for granular observability
			accountID, _ := r.Context().Value(AuthSubKey).(string)
			if accountID == "" {
				accountID = "anonymous"
			}
			
			// Optional: Extract contract ID from common path patterns if present
			contractID := chi.URLParam(r, "escrowID")

			metrics.RecordRequest(r.Context(), duration, isError, accountID, contractID)
		})
	}
}

// TokenClaims defines a unified representation of authentication claims across OIDC and local JWTs.
type TokenClaims struct {
	Subject      string   `json:"sub"`
	Email        string   `json:"email"`
	Scopes       []string `json:"scp"`
	OriginDomain string   `json:"origin_domain"`
	AuthMethod   string   `json:"auth_method"` // "oidc" or "wallet"
}

// TokenVerifier defines the interface for OIDC or Platform JWT token verification.
type TokenVerifier interface {
	Verify(ctx context.Context, token string) (*TokenClaims, error)
}

// RealVerifier is the production implementation using go-oidc.
type RealVerifier struct {
	verifier *oidc.IDTokenVerifier
}

func (v *RealVerifier) Verify(ctx context.Context, token string) (*TokenClaims, error) {
	idToken, err := v.verifier.Verify(ctx, token)
	if err != nil {
		return nil, err
	}
	var claims struct {
		Email        string   `json:"email"`
		Scopes       []string `json:"scp"`
		OriginDomain string   `json:"origin_domain"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return nil, err
	}
	return &TokenClaims{
		Subject:      idToken.Subject,
		Email:        claims.Email,
		Scopes:       claims.Scopes,
		OriginDomain: claims.OriginDomain,
		AuthMethod:   "oidc",
	}, nil
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

// PlatformJWTVerifier implements TokenVerifier for local wallet sessions.
type PlatformJWTVerifier struct {
	secret []byte
}

func NewPlatformJWTVerifier(secret []byte) *PlatformJWTVerifier {
	return &PlatformJWTVerifier{secret: secret}
}

func (v *PlatformJWTVerifier) Verify(ctx context.Context, token string) (*TokenClaims, error) {
	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.secret, nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		sub, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)
		originDomain, _ := claims["origin_domain"].(string)
		authMethod, _ := claims["auth_method"].(string)
		if authMethod == "" {
			authMethod = "wallet"
		}
		
		var scopes []string
		if scpRaw, exists := claims["scp"]; exists {
			if scpSlice, ok := scpRaw.([]interface{}); ok {
				for _, s := range scpSlice {
					if str, ok := s.(string); ok {
						scopes = append(scopes, str)
					}
				}
			}
		}

		return &TokenClaims{
			Subject:      sub,
			Email:        email,
			Scopes:       scopes,
			OriginDomain: originDomain,
			AuthMethod:   authMethod,
		}, nil
	}

	return nil, fmt.Errorf("invalid platform token")
}

// UnifiedTokenVerifier handles routing token verification between Okta and Local Platform JWTs.
type UnifiedTokenVerifier struct {
	oktaVerifier  TokenVerifier
	localVerifier *PlatformJWTVerifier
}

func NewUnifiedTokenVerifier(oktaVerifier TokenVerifier, localSecret []byte) *UnifiedTokenVerifier {
	return &UnifiedTokenVerifier{
		oktaVerifier:  oktaVerifier,
		localVerifier: NewPlatformJWTVerifier(localSecret),
	}
}

func (u *UnifiedTokenVerifier) Verify(ctx context.Context, tokenStr string) (*TokenClaims, error) {
	// High-Assurance: Check if it is our local platform JWT first to prevent slow network lookups
	if claims, err := u.localVerifier.Verify(ctx, tokenStr); err == nil {
		return claims, nil
	}

	if u.oktaVerifier == nil {
		return nil, fmt.Errorf("okta verifier is uninitialized")
	}

	return u.oktaVerifier.Verify(ctx, tokenStr)
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
				strings.HasPrefix(r.URL.Path, "/api/v1/auth/nonce") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/auth/wallet/verify") ||
				strings.HasPrefix(r.URL.Path, "/swagger") ||
				strings.HasPrefix(r.URL.Path, "/api/v1/invites/token/") {

				// Special Case: In dev bypass mode, we still want to inject the Dev User
				// if they provided the header, so they can 'Claim' an invitation token.
				if isDevBypass {
					devUser := r.Header.Get("X-Dev-User")
					if devUser != "" {
						scopes := []string{ScopeEscrowRead, ScopeEscrowWrite, ScopeEscrowAccept}
						// High-Assurance: Check if already an email to avoid double-suffix
						email := devUser
						if !strings.Contains(devUser, "@") {
							email = devUser + "@dev.local"
						}

						ctx := context.WithValue(r.Context(), AuthSubKey, devUser)
						ctx = context.WithValue(ctx, EmailKey, email)
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

				// High-Assurance: Handle full emails correctly
				email := devUser
				if !strings.Contains(devUser, "@") {
					email = devUser + "@dev.local"
				}

				ctx := context.WithValue(r.Context(), AuthSubKey, devUser)
				ctx = context.WithValue(ctx, EmailKey, email)
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

			// Verify the token using the injected verifier.
			claims, err := verifier.Verify(r.Context(), tokenStr)
			if err != nil {
				logger.Warn("failed to verify token", zap.Error(err))
				http.Error(w, "unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), AuthSubKey, claims.Subject)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			ctx = context.WithValue(ctx, ScopesKey, claims.Scopes)
			ctx = context.WithValue(ctx, AuthMethodKey, claims.AuthMethod)

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
