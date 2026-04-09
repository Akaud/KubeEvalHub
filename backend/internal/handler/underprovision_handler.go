package handler

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type UnderProvisionHandler struct {
	analysisService service.AnalysisService
}

func NewUnderProvisionHandler(analysisService service.AnalysisService) *UnderProvisionHandler {
	return &UnderProvisionHandler{
		analysisService: analysisService,
	}
}

func (h *UnderProvisionHandler) GetUnderProvisionedWorkloads(w http.ResponseWriter, r *http.Request) {
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

	thresholds := model.UnderProvisionThresholds{
		CPUPressureThreshold:             parseFloat64Default(r.URL.Query().Get("cpuPressureThreshold"), 0.90),
		MemoryPressureThreshold:          parseFloat64Default(r.URL.Query().Get("memoryPressureThreshold"), 0.90),
		CPUPressureFrequencyThreshold:    parseFloat64Default(r.URL.Query().Get("cpuPressureFrequencyThreshold"), 0.10),
		MemoryPressureFrequencyThreshold: parseFloat64Default(r.URL.Query().Get("memoryPressureFrequencyThreshold"), 0.10),
	}

	resp, err := h.analysisService.GetUnderProvisionedWorkloads(
		r.Context(),
		ownerID,
		clusterID,
		from,
		to,
		thresholds,
	)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAnalysisQuery):
			writeError(w, http.StatusBadRequest, "invalid analysis query")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to detect under-provisioned workloads")
			return
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
