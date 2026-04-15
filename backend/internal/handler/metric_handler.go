package handler

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/model"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
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
	if err := decodeAgentJSON(w, r, &req); err != nil {
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

func (h *MetricHandler) GetClusterMetrics(w http.ResponseWriter, r *http.Request) {
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

	resp, err := h.metricService.GetClusterMetrics(r.Context(), clusterID, from, to)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMetricsPayload):
			writeError(w, http.StatusBadRequest, "invalid metrics query")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to fetch metrics")
			return
		}
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *MetricHandler) ForecastMetric(w http.ResponseWriter, r *http.Request) {
	clusterID := chi.URLParam(r, "id")
	if clusterID == "" {
		writeError(w, http.StatusBadRequest, "invalid cluster id")
		return
	}

	var req model.ForecastRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

	resp, err := h.metricService.ForecastClusterMetric(r.Context(), clusterID, &req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidMetricsPayload):
			writeError(w, http.StatusBadRequest, "invalid metrics payload")
			return
		case errors.Is(err, service.ErrAgentClusterNotAssigned):
			writeError(w, http.StatusConflict, "agent is not assigned to a cluster")
			return
		case errors.Is(err, service.ErrClusterUIDMismatch):
			writeError(w, http.StatusConflict, "cluster uid does not match bound cluster")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to ingest metrics")
			return
		}
	}

	writeJSON(w, http.StatusOK, resp)
}
