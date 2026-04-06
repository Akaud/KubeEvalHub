package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrClusterNotFound = errors.New("cluster not found")

type ClusterWithAgentState struct {
	AgentID         string
	ClusterUID      string
	ClusterName     string
	KubeVersion     string
	Distribution    string
	APIServerHost   string
	Enabled         bool
	LastHeartbeatAt *time.Time
}

type ClusterRepository interface {
	Upsert(ctx context.Context, cluster *model.AgentCluster) error
	GetByAgentID(ctx context.Context, agentID string) (*model.AgentCluster, error)
	ListByOwnerID(ctx context.Context, ownerID int64) ([]ClusterWithAgentState, error)
}

type clusterRepository struct {
	pool *pgxpool.Pool
}

func NewClusterRepository(pool *pgxpool.Pool) ClusterRepository {
	return &clusterRepository{pool: pool}
}

func (r *clusterRepository) Upsert(ctx context.Context, cluster *model.AgentCluster) error {
	query := `
		INSERT INTO agent_clusters (
			agent_id,
			cluster_uid,
			cluster_name,
			kube_version,
			distribution,
			api_server_host,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (agent_id)
		DO UPDATE SET
			cluster_uid = EXCLUDED.cluster_uid,
			cluster_name = EXCLUDED.cluster_name,
			kube_version = EXCLUDED.kube_version,
			distribution = EXCLUDED.distribution,
			api_server_host = EXCLUDED.api_server_host,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.pool.Exec(
		ctx,
		query,
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
		return nil, ErrClusterNotFound
	}

	return &c, nil
}

func (r *clusterRepository) ListByOwnerID(ctx context.Context, ownerID int64) ([]ClusterWithAgentState, error) {
	query := `
		SELECT
			ac.agent_id,
			ac.cluster_uid,
			ac.cluster_name,
			ac.kube_version,
			ac.distribution,
			ac.api_server_host,
			a.enabled,
			a.last_heartbeat_at
		FROM agent_clusters ac
		INNER JOIN agents a ON a.id = ac.agent_id
		WHERE a.owner_id = $1
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
