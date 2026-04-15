package handler

import (
	"errors"
	"net/http"

	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

func ClusterRBACMiddleware(
	clusterAuthService service.ClusterAuthService,
	required service.EffectiveClusterRole,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := getAuthenticatedUserID(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			clusterID := chi.URLParam(r, "id")
			if clusterID == "" {
				writeError(w, http.StatusBadRequest, "missing cluster id")
				return
			}

			err := clusterAuthService.RequireRole(r.Context(), clusterID, userID, required)
			if err != nil {
				switch {
				case errors.Is(err, service.ErrClusterAccessDenied):
					writeError(w, http.StatusForbidden, "forbidden")
				case errors.Is(err, service.ErrClusterNotFound):
					writeError(w, http.StatusNotFound, "cluster not found")
				default:
					writeError(w, http.StatusInternalServerError, "internal server error")
				}
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
