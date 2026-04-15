package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrClusterUserRoleNotFound = errors.New("cluster user role not found")

type ClusterUserRoleRepository interface {
	GetByClusterAndUser(ctx context.Context, clusterID string, userID int64) (*model.ClusterUserRole, error)
	Upsert(ctx context.Context, clusterID string, userID int64, role model.ClusterRole, now time.Time) error
	Delete(ctx context.Context, clusterID string, userID int64) error
	ListByCluster(ctx context.Context, clusterID string) ([]model.ClusterUserRole, error)
}

type clusterUserRoleRepository struct {
	pool *pgxpool.Pool
}

func NewClusterUserRoleRepository(pool *pgxpool.Pool) ClusterUserRoleRepository {
	return &clusterUserRoleRepository{pool: pool}
}

func (r *clusterUserRoleRepository) GetByClusterAndUser(
	ctx context.Context,
	clusterID string,
	userID int64,
) (*model.ClusterUserRole, error) {
	query := `
		SELECT
			cluster_id,
			user_id,
			role,
			created_at,
			updated_at
		FROM cluster_user_roles
		WHERE cluster_id = $1
		  AND user_id = $2
	`

	var clusterUserRole model.ClusterUserRole

	err := r.pool.QueryRow(ctx, query, clusterID, userID).Scan(
		&clusterUserRole.ClusterID,
		&clusterUserRole.UserID,
		&clusterUserRole.Role,
		&clusterUserRole.CreatedAt,
		&clusterUserRole.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClusterUserRoleNotFound
		}
		return nil, err
	}

	return &clusterUserRole, nil
}

func (r *clusterUserRoleRepository) Upsert(
	ctx context.Context,
	clusterID string,
	userID int64,
	role model.ClusterRole,
	now time.Time,
) error {
	query := `
		INSERT INTO cluster_user_roles (
			cluster_id,
			user_id,
			role,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (cluster_id, user_id)
		DO UPDATE SET
			role = EXCLUDED.role,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(ctx, query, clusterID, userID, role, now, now)
	return err
}

func (r *clusterUserRoleRepository) Delete(
	ctx context.Context,
	clusterID string,
	userID int64,
) error {
	query := `
		DELETE FROM cluster_user_roles
		WHERE cluster_id = $1
		  AND user_id = $2
	`

	cmd, err := r.pool.Exec(ctx, query, clusterID, userID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrClusterUserRoleNotFound
	}

	return nil
}

func (r *clusterUserRoleRepository) ListByCluster(
	ctx context.Context,
	clusterID string,
) ([]model.ClusterUserRole, error) {
	query := `
		SELECT
			cluster_id,
			user_id,
			role,
			created_at,
			updated_at
		FROM cluster_user_roles
		WHERE cluster_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, clusterID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	roles := make([]model.ClusterUserRole, 0)

	for rows.Next() {
		var clusterUserRole model.ClusterUserRole

		err := rows.Scan(
			&clusterUserRole.ClusterID,
			&clusterUserRole.UserID,
			&clusterUserRole.Role,
			&clusterUserRole.CreatedAt,
			&clusterUserRole.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		roles = append(roles, clusterUserRole)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}
