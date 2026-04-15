package service

import (
	"context"
	"errors"

	"backend/internal/model"
	"backend/internal/repository"
)

type EffectiveClusterRole string

const (
	EffectiveClusterRoleNone     EffectiveClusterRole = "none"
	EffectiveClusterRoleViewer   EffectiveClusterRole = "viewer"
	EffectiveClusterRoleOperator EffectiveClusterRole = "operator"
	EffectiveClusterRoleAdmin    EffectiveClusterRole = "admin"
)

type ClusterAuthService interface {
	GetEffectiveRole(ctx context.Context, clusterID string, userID int64) (EffectiveClusterRole, error)
	RequireRole(ctx context.Context, clusterID string, userID int64, required EffectiveClusterRole) error
}

type clusterAuthService struct {
	clusterRepo         repository.ClusterRepository
	clusterUserRoleRepo repository.ClusterUserRoleRepository
}

var ErrClusterAccessDenied = errors.New("cluster access denied")
var ErrClusterNotFound = errors.New("cluster not found")

func NewClusterAuthService(
	clusterRepo repository.ClusterRepository,
	clusterUserRoleRepo repository.ClusterUserRoleRepository,
) ClusterAuthService {
	return &clusterAuthService{
		clusterRepo:         clusterRepo,
		clusterUserRoleRepo: clusterUserRoleRepo,
	}
}

func (s *clusterAuthService) GetEffectiveRole(
	ctx context.Context,
	clusterID string,
	userID int64,
) (EffectiveClusterRole, error) {
	cluster, err := s.clusterRepo.GetByID(ctx, clusterID)
	if err != nil {
		if errors.Is(err, repository.ErrClusterNotFound) {
			return EffectiveClusterRoleNone, ErrClusterNotFound
		}
		return EffectiveClusterRoleNone, err
	}

	if cluster.OwnerID == userID {
		return EffectiveClusterRoleAdmin, nil
	}

	clusterUserRole, err := s.clusterUserRoleRepo.GetByClusterAndUser(ctx, clusterID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrClusterUserRoleNotFound) {
			return EffectiveClusterRoleNone, nil
		}
		return EffectiveClusterRoleNone, err
	}

	switch clusterUserRole.Role {
	case model.ClusterRoleOperator:
		return EffectiveClusterRoleOperator, nil
	case model.ClusterRoleViewer:
		return EffectiveClusterRoleViewer, nil
	default:
		return EffectiveClusterRoleNone, nil
	}
}

func (s *clusterAuthService) RequireRole(
	ctx context.Context,
	clusterID string,
	userID int64,
	required EffectiveClusterRole,
) error {
	actual, err := s.GetEffectiveRole(ctx, clusterID, userID)
	if err != nil {
		return err
	}

	if !actual.Allows(required) {
		return ErrClusterAccessDenied
	}

	return nil
}

func (r EffectiveClusterRole) Allows(required EffectiveClusterRole) bool {
	return roleRank(r) >= roleRank(required)
}

func roleRank(role EffectiveClusterRole) int {
	switch role {
	case EffectiveClusterRoleAdmin:
		return 3
	case EffectiveClusterRoleOperator:
		return 2
	case EffectiveClusterRoleViewer:
		return 1
	default:
		return 0
	}
}
