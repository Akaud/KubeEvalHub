package handler

import (
	"net/http"

	"backend/internal/service"
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
	req, err := parseAnalysisRequestContext(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to detect under-provisioned workloads")
		return
	}

	thresholds, err := parseUnderProvisionThresholds(r)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to detect under-provisioned workloads")
		return
	}

	resp, err := h.analysisService.GetUnderProvisionedWorkloads(
		r.Context(),
		req.OwnerID,
		req.ClusterID,
		req.From,
		req.To,
		thresholds,
	)
	if err != nil {
		writeAnalysisHandlerError(w, err, "failed to detect under-provisioned workloads")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}
