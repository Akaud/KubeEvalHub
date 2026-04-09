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

type RecommendationHandler struct {
	analysisService service.AnalysisService
}

func NewRecommendationHandler(analysisService service.AnalysisService) *RecommendationHandler {
	return &RecommendationHandler{
		analysisService: analysisService,
	}
}

func (h *RecommendationHandler) GetRightSizingRecommendations(w http.ResponseWriter, r *http.Request) {
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

	recThresholds := model.RecommendationThresholds{
		CPUSafetyMargin:       parseFloat64Default(r.URL.Query().Get("cpuSafetyMargin"), 1.20),
		MemorySafetyMargin:    parseFloat64Default(r.URL.Query().Get("memorySafetyMargin"), 1.15),
		MinCPURequestCores:    parseFloat64Default(r.URL.Query().Get("minCpuRequestCores"), 0.05),
		MinMemoryRequestBytes: parseInt64Default(r.URL.Query().Get("minMemoryRequestBytes"), 64*1024*1024),
		MinSampleCount:        parseIntDefault(r.URL.Query().Get("minSampleCount"), 5),
	}

	underThresholds := model.UnderProvisionThresholds{
		CPUPressureThreshold:             parseFloat64Default(r.URL.Query().Get("cpuPressureThreshold"), 0.90),
		MemoryPressureThreshold:          parseFloat64Default(r.URL.Query().Get("memoryPressureThreshold"), 0.90),
		CPUPressureFrequencyThreshold:    parseFloat64Default(r.URL.Query().Get("cpuPressureFrequencyThreshold"), 0.10),
		MemoryPressureFrequencyThreshold: parseFloat64Default(r.URL.Query().Get("memoryPressureFrequencyThreshold"), 0.10),
	}

	resp, err := h.analysisService.GetRightSizingRecommendations(
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
			writeError(w, http.StatusBadRequest, "invalid analysis query")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to generate recommendations")
			return
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func parseIntDefault(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
