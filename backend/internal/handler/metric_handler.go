package handler

import (
	"errors"
	"net/http"

	"backend/internal/model"
	"backend/internal/service"
)

type MetricHandler struct {
	metricService service.MetricService
}

func NewMetricHandler(metricService service.MetricService) *MetricHandler {
	return &MetricHandler{
		metricService: metricService,
	}
}

type ingestMetricsResponse struct {
	Message string `json:"message"`
}

func (h *MetricHandler) IngestMetrics(w http.ResponseWriter, r *http.Request) {
	agentID, ok := getAgentIDFromContext(r)
	if !ok || agentID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.PushMetricsRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.metricService.IngestMetrics(r.Context(), agentID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMetricsPayload):
			writeError(w, http.StatusBadRequest, "invalid metrics payload")
			return
		case errors.Is(err, service.ErrClusterUIDMismatch):
			writeError(w, http.StatusConflict, "cluster uid does not match bound cluster")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to ingest metrics")
			return
		}
	}

	writeJSON(w, http.StatusAccepted, ingestMetricsResponse{
		Message: "metrics accepted",
	})
}
