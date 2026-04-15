package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/go-chi/chi/v5"
)

type ClusterHandler struct {
	clusterService service.ClusterService
}

func NewClusterHandler(clusterService service.ClusterService) *ClusterHandler {
	return &ClusterHandler{
		clusterService: clusterService,
	}
}

func (h *ClusterHandler) CreateCluster(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req model.CreateClusterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cluster, err := h.clusterService.CreateCluster(r.Context(), ownerID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwnerID),
			errors.Is(err, service.ErrInvalidClusterName):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, repository.ErrClusterAlreadyExists):
			writeError(w, http.StatusConflict, "cluster already exists")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to create cluster")
			return
		}
	}

	writeJSON(w, http.StatusCreated, cluster)
}

func (h *ClusterHandler) ListClusters(w http.ResponseWriter, r *http.Request) {
	userID, ok := getAuthenticatedUserID(r)
	if !ok || userID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusters, err := h.clusterService.ListAccessibleClusters(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list clusters")
		return
	}

	writeJSON(w, http.StatusOK, clusters)
}

func (h *ClusterHandler) AssignAgentToCluster(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusterID := chi.URLParam(r, "id")
	if strings.TrimSpace(clusterID) == "" {
		writeError(w, http.StatusBadRequest, "cluster id is required")
		return
	}

	var req model.AssignAgentRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.clusterService.AssignAgentToCluster(r.Context(), ownerID, clusterID, req.AgentID); err != nil {
		switch {
		case errors.Is(err, repository.ErrClusterNotFound):
			writeError(w, http.StatusNotFound, "cluster not found")
			return
		case errors.Is(err, repository.ErrAgentNotFound):
			writeError(w, http.StatusNotFound, "agent not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to assign agent")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "agent assigned",
	})
}

func (h *ClusterHandler) DeleteCluster(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusterID := chi.URLParam(r, "id")
	if strings.TrimSpace(clusterID) == "" {
		writeError(w, http.StatusBadRequest, "cluster id is required")
		return
	}

	err := h.clusterService.DeleteCluster(r.Context(), ownerID, clusterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwnerID):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, repository.ErrClusterNotFound):
			writeError(w, http.StatusNotFound, "cluster not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete cluster")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClusterHandler) UpsertClusterUserRole(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusterID := chi.URLParam(r, "id")
	if strings.TrimSpace(clusterID) == "" {
		writeError(w, http.StatusBadRequest, "cluster id is required")
		return
	}

	var req model.UpsertClusterUserRoleRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.clusterService.UpsertClusterUserRole(r.Context(), ownerID, clusterID, req.Email, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwnerID),
			errors.Is(err, service.ErrInvalidTargetEmail),
			errors.Is(err, service.ErrInvalidClusterRole),
			errors.Is(err, service.ErrClusterRoleSelfAssign):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, repository.ErrClusterNotFound):
			writeError(w, http.StatusNotFound, "cluster not found")
			return
		case errors.Is(err, repository.ErrUserNotFound):
			writeError(w, http.StatusNotFound, "user not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to upsert cluster user role")
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "cluster user role updated",
	})
}

func (h *ClusterHandler) DeleteClusterUserRole(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusterID := chi.URLParam(r, "id")
	if strings.TrimSpace(clusterID) == "" {
		writeError(w, http.StatusBadRequest, "cluster id is required")
		return
	}

	targetUserIDRaw := chi.URLParam(r, "userId")
	targetUserID, err := strconv.ParseInt(strings.TrimSpace(targetUserIDRaw), 10, 64)
	if err != nil || targetUserID <= 0 {
		writeError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	err = h.clusterService.DeleteClusterUserRole(r.Context(), ownerID, clusterID, targetUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwnerID),
			errors.Is(err, service.ErrInvalidTargetUserID),
			errors.Is(err, service.ErrCannotChangeOwnerRole):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, repository.ErrClusterNotFound):
			writeError(w, http.StatusNotFound, "cluster not found")
			return
		case errors.Is(err, repository.ErrClusterUserRoleNotFound):
			writeError(w, http.StatusNotFound, "cluster user role not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to delete cluster user role")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *ClusterHandler) ListClusterUserRoles(w http.ResponseWriter, r *http.Request) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	clusterID := chi.URLParam(r, "id")
	if strings.TrimSpace(clusterID) == "" {
		writeError(w, http.StatusBadRequest, "cluster id is required")
		return
	}

	roles, err := h.clusterService.ListClusterUserRoles(r.Context(), ownerID, clusterID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidOwnerID):
			writeError(w, http.StatusBadRequest, err.Error())
			return
		case errors.Is(err, repository.ErrClusterNotFound):
			writeError(w, http.StatusNotFound, "cluster not found")
			return
		default:
			writeError(w, http.StatusInternalServerError, "failed to list cluster user roles")
			return
		}
	}

	writeJSON(w, http.StatusOK, roles)
}
