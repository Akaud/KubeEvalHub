package handler

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDContextKey contextKey = "userID"

// define minimal interface
type tokenValidator interface {
	ValidateToken(ctx context.Context, token string) (int64, error)
}

func JWTAuthMiddleware(userService tokenValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				writeError(w, http.StatusUnauthorized, "invalid authorization header")
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, prefix))
			if token == "" {
				writeError(w, http.StatusUnauthorized, "missing token")
				return
			}

			userID, err := userService.ValidateToken(r.Context(), token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getAuthenticatedUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDContextKey).(int64)
	return userID, ok
}
