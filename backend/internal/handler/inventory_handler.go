package handler

import (
	"errors"
	"net/http"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type InventoryHandler struct {
	inventoryService service.InventoryService
}

func NewInventoryHandler(inventoryService service.InventoryService) *InventoryHandler {
	return &InventoryHandler{
		inventoryService: inventoryService,
	}
}

type ingestInventoryResponse struct {
	Message string `json:"message"`
}

func (h *InventoryHandler) IngestInventory(w http.ResponseWriter, r *http.Request) {
	agentID, ok := getAgentIDFromContext(r)
	if !ok || agentID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.PushInventoryRequest
	if err := decodeAgentJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.inventoryService.IngestInventory(r.Context(), agentID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidInventoryPayload):
			writeError(w, http.StatusBadRequest, "invalid inventory payload")
			return
		case errors.Is(err, service.ErrClusterUIDMismatch):
			writeError(w, http.StatusConflict, "cluster uid does not match bound cluster")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to ingest inventory")
			return
		}
	}

	writeJSON(w, http.StatusAccepted, ingestInventoryResponse{
		Message: "inventory accepted",
	})
}

func (h *InventoryHandler) GetLatestInventory(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "id")
	if clusterID == "" {
		writeError(w, http.StatusBadRequest, "invalid cluster id")
		return
	}

	resp, err := h.inventoryService.GetLatestInventory(r.Context(), clusterID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInventorySnapshotNotFound):
			writeError(w, http.StatusNotFound, "inventory not found")
			return
		case errors.Is(err, service.ErrInvalidInventoryPayload):
			writeError(w, http.StatusBadRequest, "invalid inventory query")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch inventory")
			return
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
