package handler

import (
	"net/http"

	"backend/internal/service"
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
	req, err := parseAnalysisRequestContext(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to detect over-provisioned workloads")
		return
	}

	thresholds, err := parseOverProvisionThresholds(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to detect over-provisioned workloads")
		return
	}

	resp, err := h.analysisService.GetOverProvisionedWorkloads(
		r.Context(),
		req.OwnerID,
		req.ClusterID,
		req.From,
		req.To,
		thresholds,
	)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to detect over-provisioned workloads")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
