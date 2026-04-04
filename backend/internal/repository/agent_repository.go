package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrAgentNotFound = errors.New("agent not found")

type AgentRepository interface {
	Create(ctx context.Context, agent *model.Agent) error
	GetByIDAndOwnerID(ctx context.Context, id string, ownerID int64) (*model.Agent, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (*model.Agent, error)
	ListByOwnerID(ctx context.Context, ownerID int64) ([]model.Agent, error)
	UpdateEnabled(ctx context.Context, id string, ownerID int64, enabled bool, updatedAt time.Time) error
	UpdateHeartbeat(ctx context.Context, id string, active bool, heartbeatAt, updatedAt time.Time) error
}

type agentRepository struct {
	pool *pgxpool.Pool
}

func NewAgentRepository(pool *pgxpool.Pool) AgentRepository {
	return &agentRepository{pool: pool}
}

func (r *agentRepository) Create(ctx context.Context, agent *model.Agent) error {
	query := `
		INSERT INTO agents (
			id,
			owner_id,
			name,
			enabled,
			active,
			token_hash,
			last_heartbeat_at,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`

	_, err := r.pool.Exec(ctx, query,
		agent.ID,
		agent.OwnerID,
		agent.Name,
		agent.Enabled,
		agent.Active,
		agent.TokenHash,
		agent.LastHeartbeatAt,
		agent.CreatedAt,
		agent.UpdatedAt,
	)

	return err
}

func (r *agentRepository) GetByIDAndOwnerID(ctx context.Context, id string, ownerID int64) (*model.Agent, error) {
	query := `
		SELECT
			id,
			owner_id,
			name,
			enabled,
			active,
			token_hash,
			last_heartbeat_at,
			created_at,
			updated_at
		FROM agents
		WHERE id = $1 AND owner_id = $2
	`

	return r.scanOne(ctx, query, id, ownerID)
}

func (r *agentRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*model.Agent, error) {
	query := `
		SELECT
			id,
			owner_id,
			name,
			enabled,
			active,
			token_hash,
			last_heartbeat_at,
			created_at,
			updated_at
		FROM agents
		WHERE token_hash = $1
	`

	return r.scanOne(ctx, query, tokenHash)
}

func (r *agentRepository) ListByOwnerID(ctx context.Context, ownerID int64) ([]model.Agent, error) {
	query := `
		SELECT
			id,
			owner_id,
			name,
			enabled,
			active,
			token_hash,
			last_heartbeat_at,
			created_at,
			updated_at
		FROM agents
		WHERE owner_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []model.Agent

	for rows.Next() {
		var a model.Agent
		err := rows.Scan(
			&a.ID,
			&a.OwnerID,
			&a.Name,
			&a.Enabled,
			&a.Active,
			&a.TokenHash,
			&a.LastHeartbeatAt,
			&a.CreatedAt,
			&a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		agents = append(agents, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return agents, nil
}

func (r *agentRepository) UpdateEnabled(ctx context.Context, id string, ownerID int64, enabled bool, updatedAt time.Time) error {
	query := `
		UPDATE agents
		SET enabled = $1,
		    updated_at = $2
		WHERE id = $3 AND owner_id = $4
	`

	cmd, err := r.pool.Exec(ctx, query, enabled, updatedAt, id, ownerID)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrAgentNotFound
	}

	return nil
}

func (r *agentRepository) UpdateHeartbeat(ctx context.Context, id string, active bool, heartbeatAt, updatedAt time.Time) error {
	query := `
		UPDATE agents
		SET active = $1,
		    last_heartbeat_at = $2,
		    updated_at = $3
		WHERE id = $4
	`

	cmd, err := r.pool.Exec(ctx, query, active, heartbeatAt, updatedAt, id)
	if err != nil {
		return err
	}

	if cmd.RowsAffected() == 0 {
		return ErrAgentNotFound
	}

	return nil
}

func (r *agentRepository) scanOne(ctx context.Context, query string, args ...any) (*model.Agent, error) {
	var a model.Agent

	err := r.pool.QueryRow(ctx, query, args...).Scan(
		&a.ID,
		&a.OwnerID,
		&a.Name,
		&a.Enabled,
		&a.Active,
		&a.TokenHash,
		&a.LastHeartbeatAt,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		return nil, ErrAgentNotFound
	}

	return &a, nil
}
