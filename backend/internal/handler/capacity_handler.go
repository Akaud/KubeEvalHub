package handler

import (
	"net/http"

	"backend/internal/service"
)

type CapacityHandler struct {
	analysisService service.AnalysisService
}

func NewCapacityHandler(s service.AnalysisService) *CapacityHandler {
	return &CapacityHandler{analysisService: s}
}

func (h *CapacityHandler) GetClusterCapacity(w http.ResponseWriter, r *http.Request) {
	req, err := parseAnalysisRequestContext(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to compute capacity")
		return
	}

	recThresholds, err := parseCapacityRecommendationThresholds(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to compute capacity")
		return
	}

	underThresholds, err := parseUnderProvisionThresholds(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to compute capacity")
		return
	}

	resp, err := h.analysisService.GetClusterCapacity(
		r.Context(),
		req.OwnerID,
		req.ClusterID,
		req.From,
		req.To,
		recThresholds,
		underThresholds,
	)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to compute capacity")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
