-- +goose Up
-- +goose StatementBegin

CREATE TABLE agent_clusters (
    agent_id uuid PRIMARY KEY REFERENCES agents(id) ON DELETE CASCADE,
    cluster_uid text NOT NULL UNIQUE,
    cluster_name text NOT NULL,
    kube_version text NOT NULL,
    distribution text NOT NULL,
    api_server_host text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE INDEX agent_clusters_cluster_uid_idx
ON agent_clusters (cluster_uid);

CREATE TABLE metric_series (
    id uuid PRIMARY KEY,
    agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    metric_name text NOT NULL,
    metric_type text NOT NULL,
    unit text NOT NULL,
    resource_kind text NOT NULL,

    node_name text,
    namespace text,
    pod_name text,
    pod_uid text,
    container_name text,

    controller_uid text,
    controller_kind text,
    controller_name text,

    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL
);

CREATE UNIQUE INDEX metric_series_identity_uidx
ON metric_series (
    agent_id,
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

CREATE INDEX metric_series_agent_id_idx
ON metric_series (agent_id);

CREATE INDEX metric_series_agent_metric_idx
ON metric_series (agent_id, metric_name);

CREATE INDEX metric_series_agent_controller_idx
ON metric_series (agent_id, controller_uid);

CREATE INDEX metric_series_agent_namespace_idx
ON metric_series (agent_id, namespace);

CREATE TABLE metric_samples (
    id uuid PRIMARY KEY,
    series_id uuid NOT NULL REFERENCES metric_series(id) ON DELETE CASCADE,
    collected_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL,
    value_double double precision NOT NULL
);

CREATE UNIQUE INDEX metric_samples_series_collected_at_uidx
ON metric_samples (series_id, collected_at);

CREATE INDEX metric_samples_series_time_idx
ON metric_samples (series_id, collected_at DESC);

CREATE INDEX metric_samples_collected_at_idx
ON metric_samples (collected_at DESC);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS metric_samples_collected_at_idx;
DROP INDEX IF EXISTS metric_samples_series_time_idx;
DROP INDEX IF EXISTS metric_samples_series_collected_at_uidx;
DROP TABLE IF EXISTS metric_samples;

DROP INDEX IF EXISTS metric_series_agent_namespace_idx;
DROP INDEX IF EXISTS metric_series_agent_controller_idx;
DROP INDEX IF EXISTS metric_series_agent_metric_idx;
DROP INDEX IF EXISTS metric_series_agent_id_idx;
DROP INDEX IF EXISTS metric_series_identity_uidx;
DROP TABLE IF EXISTS metric_series;

DROP INDEX IF EXISTS agent_clusters_cluster_uid_idx;
DROP TABLE IF EXISTS agent_clusters;

-- +goose StatementEnd