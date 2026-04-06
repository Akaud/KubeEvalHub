package handler

import (
	"errors"
	"net/http"
	"strings"

	"backend/internal/repository"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type AgentHandler struct {
	agentService service.AgentService
}

func NewAgentHandler(agentService service.AgentService) *AgentHandler {
	return &AgentHandler{
		agentService: agentService,
	}
}

type createAgentRequest struct {
	Name string `json:"name"`
}

type updateAgentEnabledRequest struct {
	Enabled bool `json:"enabled"`
}

type heartbeatResponse struct {
	Message string `json:"message"`
}

func (h *AgentHandler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req createAgentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	result, err := h.agentService.CreateAgent(r.Context(), ownerID, req.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create agent")
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *AgentHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	agents, err := h.agentService.ListAgentsByOwnerID(r.Context(), ownerID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}

	writeJSON(w, http.StatusOK, agents)
}

func (h *AgentHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	agentID := chi.URLParam(r, "id")
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	agent, err := h.agentService.GetAgentByIDAndOwnerID(r.Context(), agentID, ownerID)
	if err != nil {
		if errors.Is(err, repository.ErrAgentNotFound) {
			writeError(w, http.StatusNotFound, "agent not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to get agent")
		return
	}

	writeJSON(w, http.StatusOK, agent)
}

func (h *AgentHandler) UpdateAgentEnabled(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	agentID := chi.URLParam(r, "id")
	if agentID == "" {
		writeError(w, http.StatusBadRequest, "agent id is required")
		return
	}

	var req updateAgentEnabledRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.agentService.SetEnabled(r.Context(), agentID, ownerID, req.Enabled)
	if err != nil {
		if errors.Is(err, repository.ErrAgentNotFound) {
			writeError(w, http.StatusNotFound, "agent not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to update agent")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "agent updated",
	})
}

func (h *AgentHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	agentID, ok := getAgentIDFromContext(r)
	if !ok || agentID == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	if err := h.agentService.Heartbeat(r.Context(), agentID); err != nil {
		if errors.Is(err, repository.ErrAgentNotFound) {
			writeError(w, http.StatusNotFound, "agent not found")
			return
		}

		writeError(w, http.StatusInternalServerError, "failed to process heartbeat")
		return
	}

	writeJSON(w, http.StatusOK, heartbeatResponse{
		Message: "heartbeat accepted",
	})
}
