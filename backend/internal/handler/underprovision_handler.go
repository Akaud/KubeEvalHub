package handler

import (
	"log"
	"net/http"
	"time"

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
		log.Printf("GetUnderProvisionedWorkloads parseAnalysisRequestContext error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to detect under-provisioned workloads")
		return
	}

	thresholds, err := parseUnderProvisionThresholds(r)
	if err != nil {
		log.Printf("GetUnderProvisionedWorkloads parseUnderProvisionThresholds error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to detect under-provisioned workloads")
		return
	}

	log.Printf(
		"GetUnderProvisionedWorkloads owner=%d cluster=%s from=%s to=%s thresholds=%+v",
		req.OwnerID,
		req.ClusterID,
		req.From.Format(time.RFC3339),
		req.To.Format(time.RFC3339),
		thresholds,
	)

	resp, err := h.analysisService.GetUnderProvisionedWorkloads(
		r.Context(),
		req.ClusterID,
		req.From,
		req.To,
		thresholds,
	)
	if err != nil {
		log.Printf("GetUnderProvisionedWorkloads service error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to detect under-provisioned workloads")
		return
	}

	log.Printf("GetUnderProvisionedWorkloads response: %+v", resp)
	writeJSON(w, http.StatusOK, resp)
}
