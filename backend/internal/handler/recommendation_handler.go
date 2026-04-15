package handler

import (
	"net/http"

	"backend/internal/service"
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
	req, err := parseAnalysisRequestContext(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to generate recommendations")
		return
	}

	recThresholds, err := parseRecommendationThresholds(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to generate recommendations")
		return
	}

	underThresholds, err := parseUnderProvisionThresholds(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to generate recommendations")
		return
	}

	resp, err := h.analysisService.GetRightSizingRecommendations(
		r.Context(),
		req.ClusterID,
		req.From,
		req.To,
		recThresholds,
		underThresholds,
	)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to generate recommendations")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
