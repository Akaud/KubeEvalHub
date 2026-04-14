-- +goose Up
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS metric_series (
    id UUID PRIMARY KEY,
    cluster_id UUID NOT NULL REFERENCES agent_clusters(id) ON DELETE CASCADE,
    metric_name TEXT NOT NULL,
    metric_type TEXT NOT NULL,
    unit TEXT NOT NULL,
    resource_kind TEXT NOT NULL,

    node_name TEXT,
    namespace TEXT,
    pod_name TEXT,
    pod_uid TEXT,
    container_name TEXT,

    controller_uid TEXT,
    controller_kind TEXT,
    controller_name TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS metric_series_identity_uidx
ON metric_series (
    cluster_id,
    metric_name,
    resource_kind,
    node_name,
    namespace,
    pod_name,
    pod_uid,
    container_name,
    controller_uid,
    controller_kind,
    controller_name
)
NULLS NOT DISTINCT;

CREATE INDEX IF NOT EXISTS metric_series_cluster_id_idx
    ON metric_series (cluster_id);

CREATE INDEX IF NOT EXISTS metric_series_cluster_metric_idx
    ON metric_series (cluster_id, metric_name);

CREATE INDEX IF NOT EXISTS metric_series_cluster_controller_idx
    ON metric_series (cluster_id, controller_uid);

CREATE INDEX IF NOT EXISTS metric_series_cluster_namespace_idx
    ON metric_series (cluster_id, namespace);

CREATE TABLE IF NOT EXISTS metric_samples (
    id UUID PRIMARY KEY,
    series_id UUID NOT NULL REFERENCES metric_series(id) ON DELETE CASCADE,
    collected_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL,
    value_double DOUBLE PRECISION NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS metric_samples_series_collected_at_uidx
    ON metric_samples (series_id, collected_at);

CREATE INDEX IF NOT EXISTS metric_samples_series_time_idx
    ON metric_samples (series_id, collected_at DESC);

CREATE INDEX IF NOT EXISTS metric_samples_collected_at_idx
    ON metric_samples (collected_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS metric_samples_collected_at_idx;
DROP INDEX IF EXISTS metric_samples_series_time_idx;
DROP INDEX IF EXISTS metric_samples_series_collected_at_uidx;
DROP TABLE IF EXISTS metric_samples;

DROP INDEX IF EXISTS metric_series_cluster_namespace_idx;
DROP INDEX IF EXISTS metric_series_cluster_controller_idx;
DROP INDEX IF EXISTS metric_series_cluster_metric_idx;
DROP INDEX IF EXISTS metric_series_cluster_id_idx;
DROP INDEX IF EXISTS metric_series_identity_uidx;
DROP TABLE IF EXISTS metric_series;

-- +goose StatementEnd