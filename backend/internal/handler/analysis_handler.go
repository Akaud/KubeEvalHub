package handler

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type AnalysisHandler struct {
	analysisService service.AnalysisService
}

func NewAnalysisHandler(analysisService service.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		analysisService: analysisService,
	}
}

func (h *AnalysisHandler) GetWorkloadUtilization(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusterID := chi.URLParam(r, "id")
	if clusterID == "" {
		writeError(w, http.StatusBadRequest, "invalid cluster id")
		return
	}

	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	if fromStr == "" || toStr == "" {
		writeError(w, http.StatusBadRequest, "from and to are required")
		return
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid from")
		return
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid to")
		return
	}

	resp, err := h.analysisService.GetWorkloadUtilization(r.Context(), ownerID, clusterID, from, to)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAnalysisQuery):
			writeError(w, http.StatusBadRequest, "invalid analysis query")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to compute utilization")
			return
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
