package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrInvalidOwnerID     = errors.New("invalid owner id")
	ErrInvalidClusterName = errors.New("cluster name is required")
)

type ClusterService interface {
	CreateCluster(ctx context.Context, ownerID int64, req model.CreateClusterRequest) (*model.ClusterView, error)
	DeleteCluster(ctx context.Context, ownerID int64, clusterID string) error
	AssignAgentToCluster(ctx context.Context, ownerID int64, clusterID string, agentID string) error
	UpdateClusterMetadataFromAgent(ctx context.Context, agentID string, payload model.ClusterPayload) error
	ListClustersByOwnerID(ctx context.Context, ownerID int64) ([]model.ClusterView, error)
}

type clusterService struct {
	clusterRepo repository.ClusterRepository
	agentRepo   repository.AgentRepository
}

func NewClusterService(clusterRepo repository.ClusterRepository, agentRepo repository.AgentRepository) ClusterService {
	return &clusterService{
		clusterRepo: clusterRepo,
		agentRepo:   agentRepo,
	}
}

func (s *clusterService) UpdateClusterMetadataFromAgent(ctx context.Context, agentID string, payload model.ClusterPayload) error {
	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return errors.New("agent id is required")
	}

	return s.clusterRepo.UpdateMetadataByAgentID(
		ctx,
		agentID,
		strings.TrimSpace(payload.ClusterUID),
		strings.TrimSpace(payload.ClusterName),
		strings.TrimSpace(payload.KubeVersion),
		strings.TrimSpace(payload.Distribution),
		strings.TrimSpace(payload.APIServerHost),
		time.Now().UTC(),
	)
}

func (s *clusterService) DeleteCluster(ctx context.Context, ownerID int64, clusterID string) error {
	clusterID = strings.TrimSpace(clusterID)

	if ownerID <= 0 {
		return ErrInvalidOwnerID
	}
	if clusterID == "" {
		return errors.New("cluster id is required")
	}

	return s.clusterRepo.Delete(ctx, clusterID, ownerID)
}

func (s *clusterService) CreateCluster(ctx context.Context, ownerID int64, req model.CreateClusterRequest) (*model.ClusterView, error) {
	req.ClusterName = strings.TrimSpace(req.ClusterName)

	switch {
	case ownerID <= 0:
		return nil, ErrInvalidOwnerID
	case req.ClusterName == "":
		return nil, ErrInvalidClusterName
	}

	now := time.Now().UTC()

	cluster := &model.AgentCluster{
		ID:            uuid.NewString(),
		OwnerID:       ownerID,
		AgentID:       nil,
		ClusterUID:    "",
		ClusterName:   req.ClusterName,
		KubeVersion:   "",
		Distribution:  "manual",
		APIServerHost: "",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.clusterRepo.Create(ctx, cluster); err != nil {
		return nil, err
	}

	return &model.ClusterView{
		ID:              cluster.ID,
		OwnerID:         cluster.OwnerID,
		AgentID:         cluster.AgentID,
		ClusterUID:      cluster.ClusterUID,
		ClusterName:     cluster.ClusterName,
		KubeVersion:     cluster.KubeVersion,
		Distribution:    cluster.Distribution,
		APIServerHost:   cluster.APIServerHost,
		LastHeartbeatAt: nil,
		Status:          model.AgentStatusNeverConnected,
	}, nil
}

func (s *clusterService) AssignAgentToCluster(ctx context.Context, ownerID int64, clusterID string, agentID string) error {
	clusterID = strings.TrimSpace(clusterID)
	agentID = strings.TrimSpace(agentID)

	if ownerID <= 0 {
		return ErrInvalidOwnerID
	}
	if clusterID == "" {
		return errors.New("cluster id is required")
	}
	if agentID == "" {
		return errors.New("agent id is required")
	}

	agent, err := s.agentRepo.GetByIDAndOwnerID(ctx, agentID, ownerID)
	if err != nil {
		return err
	}

	_ = agent

	return s.clusterRepo.AssignAgent(ctx, clusterID, ownerID, agentID, time.Now().UTC())
}

func (s *clusterService) ListClustersByOwnerID(ctx context.Context, ownerID int64) ([]model.ClusterView, error) {
	rows, err := s.clusterRepo.ListByOwnerID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	clusters := make([]model.ClusterView, 0, len(rows))

	for _, row := range rows {
		status := model.AgentStatusNeverConnected

		switch {
		case row.AgentID == nil:
			status = model.AgentStatusNeverConnected
		case row.Enabled != nil && !*row.Enabled:
			status = model.AgentStatusDisabled
		case row.LastHeartbeatAt == nil:
			status = model.AgentStatusNeverConnected
		case now.Sub(*row.LastHeartbeatAt) <= 90*time.Second:
			status = model.AgentStatusOnline
		default:
			status = model.AgentStatusOffline
		}

		clusters = append(clusters, model.ClusterView{
			ID:              row.ClusterID,
			OwnerID:         row.OwnerID,
			AgentID:         row.AgentID,
			ClusterUID:      row.ClusterUID,
			ClusterName:     row.ClusterName,
			KubeVersion:     row.KubeVersion,
			Distribution:    row.Distribution,
			APIServerHost:   row.APIServerHost,
			LastHeartbeatAt: row.LastHeartbeatAt,
			Status:          status,
		})
	}

	return clusters, nil
}
