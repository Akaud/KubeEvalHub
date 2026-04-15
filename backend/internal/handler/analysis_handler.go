package handler

import (
	"log"
	"net/http"
	"time"

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
		log.Printf("GetWorkloadUtilization parseAnalysisRequestContext error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to compute utilization")
		return
	}

	log.Printf(
		"GetWorkloadUtilization owner=%d cluster=%s from=%s to=%s",
		req.OwnerID,
		req.ClusterID,
		req.From.Format(time.RFC3339),
		req.To.Format(time.RFC3339),
	)

	resp, err := h.analysisService.GetWorkloadUtilization(
		r.Context(),
		req.ClusterID,
		req.From,
		req.To,
	)
	if err != nil {
		log.Printf("GetWorkloadUtilization service error: %v", err)
		writeAnalysisHandlerError(w, err, "failed to compute utilization")
		return
	}

	log.Printf("GetWorkloadUtilization response: %+v", resp)
	writeJSON(w, http.StatusOK, resp)
}
