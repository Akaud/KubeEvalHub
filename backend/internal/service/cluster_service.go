package service

import (
	"context"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
)

type ClusterService interface {
	ListClustersByOwnerID(ctx context.Context, ownerID int64) ([]model.ClusterView, error)
}

type clusterService struct {
	clusterRepo repository.ClusterRepository
}

func NewClusterService(clusterRepo repository.ClusterRepository) ClusterService {
	return &clusterService{
		clusterRepo: clusterRepo,
	}
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

		if !row.Enabled {
			status = model.AgentStatusDisabled
		} else if row.LastHeartbeatAt == nil {
			status = model.AgentStatusNeverConnected
		} else if now.Sub(*row.LastHeartbeatAt) <= 90*time.Second {
			status = model.AgentStatusOnline
		} else {
			status = model.AgentStatusOffline
		}

		clusters = append(clusters, model.ClusterView{
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
