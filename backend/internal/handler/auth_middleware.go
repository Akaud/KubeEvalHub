package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"backend/internal/service"
)

type contextKey string

const (
	userIDContextKey  contextKey = "userID"
	agentIDContextKey contextKey = "agentID"
)

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

func AgentAuthMiddleware(agentService service.AgentService) func(http.Handler) http.Handler {
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

			agent, err := agentService.AuthenticateByToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, service.ErrAgentDisabled) {
					writeError(w, http.StatusForbidden, "agent is disabled")
					return
				}

				writeError(w, http.StatusUnauthorized, "invalid agent token")
				return
			}

			ctx := context.WithValue(r.Context(), agentIDContextKey, agent.ID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func getAuthenticatedUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDContextKey).(int64)
	return userID, ok
}

func getAgentIDFromContext(r *http.Request) (string, bool) {
	agentID, ok := r.Context().Value(agentIDContextKey).(string)
	return agentID, ok
}
