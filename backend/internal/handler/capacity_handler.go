package handler

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type CapacityHandler struct {
	analysisService service.AnalysisService
}

func NewCapacityHandler(s service.AnalysisService) *CapacityHandler {
	return &CapacityHandler{analysisService: s}
}

func (h *CapacityHandler) GetClusterCapacity(w http.ResponseWriter, r *http.Request) {
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

	recThresholds := model.RecommendationThresholds{
		CPUSafetyMargin:    parseFloat64Default(r.URL.Query().Get("cpuSafetyMargin"), 1.20),
		MemorySafetyMargin: parseFloat64Default(r.URL.Query().Get("memorySafetyMargin"), 1.15),
	}

	underThresholds := model.UnderProvisionThresholds{
		CPUPressureThreshold:             parseFloat64Default(r.URL.Query().Get("cpuPressureThreshold"), 0.90),
		MemoryPressureThreshold:          parseFloat64Default(r.URL.Query().Get("memoryPressureThreshold"), 0.90),
		CPUPressureFrequencyThreshold:    parseFloat64Default(r.URL.Query().Get("cpuPressureFrequencyThreshold"), 0.10),
		MemoryPressureFrequencyThreshold: parseFloat64Default(r.URL.Query().Get("memoryPressureFrequencyThreshold"), 0.10),
	}

	resp, err := h.analysisService.GetClusterCapacity(
		r.Context(),
		ownerID,
		clusterID,
		from,
		to,
		recThresholds,
		underThresholds,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAnalysisQuery):
			writeError(w, http.StatusBadRequest, "invalid query")
		default:
			writeError(w, http.StatusInternalServerError, "failed to compute capacity")
		}
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
