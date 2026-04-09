package handler

import (
	"net/http"

	"backend/internal/service"
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
	req, err := parseAnalysisRequestContext(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to compute utilization")
		return
	}

	resp, err := h.analysisService.GetWorkloadUtilization(
		r.Context(),
		req.OwnerID,
		req.ClusterID,
		req.From,
		req.To,
	)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to compute utilization")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
