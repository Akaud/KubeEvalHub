package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrClusterNotFound = errors.New("cluster not found")
var ErrClusterAlreadyExists = errors.New("cluster already exists")

type ClusterWithAgentState struct {
	ClusterID       string
	OwnerID         int64
	AgentID         *string
	ClusterUID      string
	ClusterName     string
	KubeVersion     string
	Distribution    string
	APIServerHost   string
	Enabled         *bool
	LastHeartbeatAt *time.Time
}

type ClusterRepository interface {
	Create(ctx context.Context, cluster *model.AgentCluster) error
	Upsert(ctx context.Context, cluster *model.AgentCluster) error
	GetByAgentID(ctx context.Context, agentID string) (*model.AgentCluster, error)
	GetByIDAndOwnerID(ctx context.Context, id string, ownerID int64) (*model.AgentCluster, error)
	AssignAgent(ctx context.Context, clusterID string, ownerID int64, agentID string, updatedAt time.Time) error
	UpdateMetadataByAgentID(ctx context.Context, agentID string, clusterUID, clusterName, kubeVersion, distribution, apiServerHost string, updatedAt time.Time) error
	ListByOwnerID(ctx context.Context, ownerID int64) ([]ClusterWithAgentState, error)
	Delete(ctx context.Context, id string, ownerID int64) error
}

type clusterRepository struct {
	pool *pgxpool.Pool
}

func NewClusterRepository(pool *pgxpool.Pool) ClusterRepository {
	return &clusterRepository{pool: pool}
}
func (r *clusterRepository) UpdateMetadataByAgentID(
	ctx context.Context,
	agentID string,
	clusterUID, clusterName, kubeVersion, distribution, apiServerHost string,
	updatedAt time.Time,
) error {
	query := `
		UPDATE agent_clusters
		SET cluster_uid = $1,
		    cluster_name = $2,
		    kube_version = $3,
		    distribution = $4,
		    api_server_host = $5,
		    updated_at = $6
		WHERE agent_id = $7
	`

	cmd, err := r.pool.Exec(
		ctx,
		query,
		clusterUID,
		clusterName,
		kubeVersion,
		distribution,
		apiServerHost,
		updatedAt,
		agentID,
	)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrClusterNotFound
	}

	return nil
}

func (r *clusterRepository) Delete(ctx context.Context, id string, ownerID int64) error {
	query := `
		DELETE FROM agent_clusters
		WHERE id = $1 AND owner_id = $2
	`

	cmd, err := r.pool.Exec(ctx, query, id, ownerID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrClusterNotFound
	}

	return nil
}

func (r *clusterRepository) AssignAgent(ctx context.Context, clusterID string, ownerID int64, agentID string, updatedAt time.Time) error {
	query := `
		UPDATE agent_clusters
		SET agent_id = $1,
		    updated_at = $2
		WHERE id = $3 AND owner_id = $4
	`

	cmd, err := r.pool.Exec(ctx, query, agentID, updatedAt, clusterID, ownerID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrClusterNotFound
	}

	return nil
}

func (r *clusterRepository) GetByIDAndOwnerID(ctx context.Context, id string, ownerID int64) (*model.AgentCluster, error) {
	query := `
		SELECT
			id,
			owner_id,
			agent_id,
			cluster_uid,
			cluster_name,
			kube_version,
			distribution,
			api_server_host,
			created_at,
			updated_at
		FROM agent_clusters
		WHERE id = $1 AND owner_id = $2
	`

	var c model.AgentCluster

	err := r.pool.QueryRow(ctx, query, id, ownerID).Scan(
		&c.ID,
		&c.OwnerID,
		&c.AgentID,
		&c.ClusterUID,
		&c.ClusterName,
		&c.KubeVersion,
		&c.Distribution,
		&c.APIServerHost,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClusterNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *clusterRepository) Create(ctx context.Context, cluster *model.AgentCluster) error {
	query := `
		INSERT INTO agent_clusters (
			id,
			owner_id,
			agent_id,
			cluster_uid,
			cluster_name,
			kube_version,
			distribution,
			api_server_host,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		cluster.ID,
		cluster.OwnerID,
		cluster.AgentID,
		cluster.ClusterUID,
		cluster.ClusterName,
		cluster.KubeVersion,
		cluster.Distribution,
		cluster.APIServerHost,
		cluster.CreatedAt,
		cluster.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrClusterAlreadyExists
		}
		return err
	}

	return nil
}

func (r *clusterRepository) Upsert(ctx context.Context, cluster *model.AgentCluster) error {
	query := `
		INSERT INTO agent_clusters (
			id,
			owner_id,
			agent_id,
			cluster_uid,
			cluster_name,
			kube_version,
			distribution,
			api_server_host,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (cluster_uid)
		DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			agent_id = EXCLUDED.agent_id,
			cluster_name = EXCLUDED.cluster_name,
			kube_version = EXCLUDED.kube_version,
			distribution = EXCLUDED.distribution,
			api_server_host = EXCLUDED.api_server_host,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		cluster.ID,
		cluster.OwnerID,
		cluster.AgentID,
		cluster.ClusterUID,
		cluster.ClusterName,
		cluster.KubeVersion,
		cluster.Distribution,
		cluster.APIServerHost,
		cluster.CreatedAt,
		cluster.UpdatedAt,
	)

	return err
}

func (r *clusterRepository) GetByAgentID(ctx context.Context, agentID string) (*model.AgentCluster, error) {
	query := `
		SELECT
			id,
			owner_id,
			agent_id,
			cluster_uid,
			cluster_name,
			kube_version,
			distribution,
			api_server_host,
			created_at,
			updated_at
		FROM agent_clusters
		WHERE agent_id = $1
	`

	var c model.AgentCluster

	err := r.pool.QueryRow(ctx, query, agentID).Scan(
		&c.ID,
		&c.OwnerID,
		&c.AgentID,
		&c.ClusterUID,
		&c.ClusterName,
		&c.KubeVersion,
		&c.Distribution,
		&c.APIServerHost,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrClusterNotFound
		}
		return nil, err
	}

	return &c, nil
}

func (r *clusterRepository) ListByOwnerID(ctx context.Context, ownerID int64) ([]ClusterWithAgentState, error) {
	query := `
		SELECT
			ac.id,
			ac.owner_id,
			ac.agent_id,
			ac.cluster_uid,
			ac.cluster_name,
			ac.kube_version,
			ac.distribution,
			ac.api_server_host,
			a.enabled,
			a.last_heartbeat_at
		FROM agent_clusters ac
		LEFT JOIN agents a ON a.id = ac.agent_id
		WHERE ac.owner_id = $1
		ORDER BY ac.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	clusters := make([]ClusterWithAgentState, 0)

	for rows.Next() {
		var c ClusterWithAgentState

		err := rows.Scan(
			&c.ClusterID,
			&c.OwnerID,
			&c.AgentID,
			&c.ClusterUID,
			&c.ClusterName,
			&c.KubeVersion,
			&c.Distribution,
			&c.APIServerHost,
			&c.Enabled,
			&c.LastHeartbeatAt,
		)
		if err != nil {
			return nil, err
		}

		clusters = append(clusters, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return clusters, nil
}
