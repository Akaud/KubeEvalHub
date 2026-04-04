-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT FALSE,
    active BOOLEAN NOT NULL DEFAULT FALSE,
    token_hash TEXT NOT NULL UNIQUE,
    last_heartbeat_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_agents_owner_id_name UNIQUE (owner_id, name)
);

CREATE INDEX IF NOT EXISTS idx_agents_owner_id
    ON agents (owner_id);

CREATE INDEX IF NOT EXISTS idx_agents_enabled
    ON agents (enabled);

CREATE INDEX IF NOT EXISTS idx_agents_active
    ON agents (active);

CREATE INDEX IF NOT EXISTS idx_agents_owner_id_enabled
    ON agents (owner_id, enabled);

CREATE INDEX IF NOT EXISTS idx_agents_owner_id_active
    ON agents (owner_id, active);

CREATE INDEX IF NOT EXISTS idx_agents_last_heartbeat_at
    ON agents (last_heartbeat_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_agents_last_heartbeat_at;
DROP INDEX IF EXISTS idx_agents_owner_id_active;
DROP INDEX IF EXISTS idx_agents_owner_id_enabled;
DROP INDEX IF EXISTS idx_agents_active;
DROP INDEX IF EXISTS idx_agents_enabled;
DROP INDEX IF EXISTS idx_agents_owner_id;

DROP TABLE IF EXISTS agents;
-- +goose StatementEnd