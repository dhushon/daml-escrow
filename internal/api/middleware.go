package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"daml-escrow/internal/services"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type contextKey string

const (
	AuthSubKey contextKey = "auth_sub"
	EmailKey   contextKey = "email"
	ScopesKey  contextKey = "scopes"
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

func AuthMiddleware(issuer, clientId, audience string, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass auth for health, swagger, and anonymous token endpoints
			if strings.HasPrefix(r.URL.Path, "/api/v1/health") || 
			   strings.HasPrefix(r.URL.Path, "/api/v1/auth/discover") ||
			   strings.HasPrefix(r.URL.Path, "/swagger") ||
			   strings.HasPrefix(r.URL.Path, "/api/v1/invites/token/") {
				// Special Case: In AUTH_DEV_MODE, we still want to inject the Dev User 
				// if they provided the header, so they can 'Claim' the token.
				if os.Getenv("AUTH_DEV_MODE") == "true" {
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
			if os.Getenv("AUTH_DEV_MODE") == "true" {
				devUser := r.Header.Get("X-Dev-User")
				if devUser == "" {
					devUser = "Buyer"
				}

				scopes := []string{ScopeEscrowRead}
				if devUser == "Buyer" || devUser == "Seller" || devUser == "EscrowMediator" || devUser == "CentralBank" {
					scopes = append(scopes, ScopeEscrowWrite, ScopeEscrowAccept)
				}
				if devUser == "CentralBank" {
					scopes = append(scopes, ScopeSystemAdmin)
				}

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

			// Token parsing (validation skipped for prototype walkthrough)
			token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
			if err != nil {
				logger.Error("failed to parse token", zap.Error(err))
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid claims", http.StatusUnauthorized)
				return
			}

			sub, _ := claims["sub"].(string)
			email, _ := claims["email"].(string)
			
			var scopes []string
			if scp, ok := claims["scp"].([]interface{}); ok {
				for _, s := range scp {
					scopes = append(scopes, s.(string))
				}
			}

			ctx := context.WithValue(r.Context(), AuthSubKey, sub)
			ctx = context.WithValue(ctx, EmailKey, email)
			ctx = context.WithValue(ctx, ScopesKey, scopes)
			
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireScope(ctx context.Context, required string) bool {
	scopes, ok := ctx.Value(ScopesKey).([]string)
	if !ok {
		return false
	}
	// For debugging during walkthrough
	if os.Getenv("AUTH_DEV_MODE") == "true" {
		fmt.Printf("DEBUG: RequireScope checking '%s' against user scopes: %v\n", required, scopes)
	}
	for _, s := range scopes {
		if s == required {
			if os.Getenv("AUTH_DEV_MODE") == "true" {
				fmt.Printf("DEBUG: RequireScope MATCHED '%s'\n", required)
			}
			return true
		}
	}
	if os.Getenv("AUTH_DEV_MODE") == "true" {
		fmt.Printf("DEBUG: RequireScope FAILED to find '%s'\n", required)
	}
	return false
}
