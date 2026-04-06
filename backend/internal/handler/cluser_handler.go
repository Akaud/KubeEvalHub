package handler

import (
	"net/http"

	"backend/internal/service"
)

type ClusterHandler struct {
	clusterService service.ClusterService
}

func NewClusterHandler(clusterService service.ClusterService) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterService,
	}
}

func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusters, err := h.clusterService.ListClustersByOwnerID(r.Context(), ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list clusters")
		return
	}

	writeJSON(w, http.StatusOK, clusters)
}
