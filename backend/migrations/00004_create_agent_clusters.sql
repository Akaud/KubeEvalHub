-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS agent_clusters (
    id UUID PRIMARY KEY,
    owner_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    agent_id UUID NULL UNIQUE REFERENCES agents(id) ON DELETE SET NULL,

    cluster_uid TEXT NULL,
    cluster_name TEXT NULL,
    kube_version TEXT,
    distribution TEXT,
    api_server_host TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_agent_clusters_owner_cluster_uid_nonnull
    ON agent_clusters (owner_id, cluster_uid)
    WHERE cluster_uid IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_agent_clusters_owner_id
    ON agent_clusters (owner_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_agent_clusters_owner_id;
DROP INDEX IF EXISTS uq_agent_clusters_owner_cluster_uid_nonnull;
DROP TABLE IF EXISTS agent_clusters;

-- +goose StatementEnd