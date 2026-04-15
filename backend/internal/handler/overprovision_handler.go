package handler

import (
	"log"
	"net/http"
	"time"

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
		log.Printf("GetOverProvisionedWorkloads parseAnalysisRequestContext error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to detect over-provisioned workloads")
		return
	}

	thresholds, err := parseOverProvisionThresholds(r)
	if err != nil {
		log.Printf("GetOverProvisionedWorkloads parseOverProvisionThresholds error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to detect over-provisioned workloads")
		return
	}

	log.Printf(
		"GetOverProvisionedWorkloads owner=%d cluster=%s from=%s to=%s thresholds=%+v",
		req.OwnerID,
		req.ClusterID,
		req.From.Format(time.RFC3339),
		req.To.Format(time.RFC3339),
		thresholds,
	)

	resp, err := h.analysisService.GetOverProvisionedWorkloads(
		r.Context(),
		req.ClusterID,
		req.From,
		req.To,
		thresholds,
	)
	if err != nil {
		log.Printf("GetOverProvisionedWorkloads service error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to detect over-provisioned workloads")
		return
	}

	log.Printf("GetOverProvisionedWorkloads response: %+v", resp)
	writeJSON(w, http.StatusOK, resp)
}
