package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type OverProvisionHandler struct {
	analysisService service.AnalysisService
}

func NewOverProvisionHandler(analysisService service.AnalysisService) *OverProvisionHandler {
	return &OverProvisionHandler{
		analysisService: analysisService,
	}
}

func (h *OverProvisionHandler) GetOverProvisionedWorkloads(w http.ResponseWriter, r *http.Request) {
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

	thresholds := model.OverProvisionThresholds{
		CPUSafetyMargin:          parseFloat64Default(r.URL.Query().Get("cpuSafetyMargin"), 1.20),
		MemorySafetyMargin:       parseFloat64Default(r.URL.Query().Get("memorySafetyMargin"), 1.15),
		CPUOverprovisionRatio:    parseFloat64Default(r.URL.Query().Get("cpuOverprovisionRatio"), 1.50),
		MemoryOverprovisionRatio: parseFloat64Default(r.URL.Query().Get("memoryOverprovisionRatio"), 1.50),
		MinReclaimCPUCores:       parseFloat64Default(r.URL.Query().Get("minReclaimCpuCores"), 0.10),
		MinReclaimMemoryBytes:    parseInt64Default(r.URL.Query().Get("minReclaimMemoryBytes"), 128*1024*1024),
	}

	resp, err := h.analysisService.GetOverProvisionedWorkloads(
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
			writeError(w, http.StatusInternalServerError, "failed to detect over-provisioned workloads")
			return
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func parseFloat64Default(raw string, fallback float64) float64 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return v
}

func parseInt64Default(raw string, fallback int64) int64 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return fallback
	}
	return v
}
