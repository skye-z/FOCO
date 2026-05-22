package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	authpkg "foco/backend/api/internal/auth"
)

type TokenVerifier interface {
	VerifyToken(ctx context.Context, token string) (*authpkg.Claims, error)
}

type contextKey string

const claimsContextKey contextKey = "claims"

func AuthMiddleware(verifier TokenVerifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			if verifier == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := verifier.VerifyToken(r.Context(), token)
			if err != nil || claims == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func ClaimsFromContext(ctx context.Context) *authpkg.Claims {
	claims, _ := ctx.Value(claimsContextKey).(*authpkg.Claims)
	return claims
}

func ContextWithClaims(ctx context.Context, claims *authpkg.Claims) context.Context {
	return context.WithValue(ctx, claimsContextKey, claims)
}

var ErrUnauthorized = errors.New("unauthorized")
