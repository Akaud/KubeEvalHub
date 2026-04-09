-- +goose Up
-- +goose StatementBegin

CREATE TABLE inventory_snapshots (
    id uuid PRIMARY KEY,
    agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    collected_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL
);

CREATE INDEX inventory_snapshots_agent_id_collected_at_idx
ON inventory_snapshots (agent_id, collected_at DESC);

CREATE TABLE inventory_namespaces (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES inventory_snapshots(id) ON DELETE CASCADE,
    uid text NOT NULL,
    name text NOT NULL,
    labels_json jsonb
);

CREATE INDEX inventory_namespaces_snapshot_id_idx
ON inventory_namespaces (snapshot_id);

CREATE TABLE inventory_nodes (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES inventory_snapshots(id) ON DELETE CASCADE,
    uid text NOT NULL,
    name text NOT NULL,
    labels_json jsonb,
    kubelet_version text,
    container_runtime_version text,
    operating_system text,
    architecture text,
    kernel_version text,
    os_image text
);

CREATE INDEX inventory_nodes_snapshot_id_idx
ON inventory_nodes (snapshot_id);

CREATE TABLE inventory_workloads (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES inventory_snapshots(id) ON DELETE CASCADE,
    kind text NOT NULL,
    uid text NOT NULL,
    name text NOT NULL,
    namespace text,
    replicas integer,
    labels_json jsonb
);

CREATE INDEX inventory_workloads_snapshot_id_idx
ON inventory_workloads (snapshot_id);

CREATE INDEX inventory_workloads_snapshot_kind_idx
ON inventory_workloads (snapshot_id, kind);

CREATE TABLE inventory_pods (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES inventory_snapshots(id) ON DELETE CASCADE,
    uid text NOT NULL,
    name text NOT NULL,
    namespace text NOT NULL,
    node_name text,
    phase text,
    owner_kind text,
    owner_name text,
    labels_json jsonb
);

CREATE INDEX inventory_pods_snapshot_id_idx
ON inventory_pods (snapshot_id);

CREATE TABLE inventory_containers (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES inventory_snapshots(id) ON DELETE CASCADE,
    parent_kind text NOT NULL,
    parent_ref_id uuid NOT NULL,
    name text NOT NULL,
    image text,
    cpu_request_millicores bigint,
    cpu_limit_millicores bigint,
    memory_request_bytes bigint,
    memory_limit_bytes bigint
);

CREATE INDEX inventory_containers_snapshot_id_idx
ON inventory_containers (snapshot_id);

CREATE INDEX inventory_containers_parent_idx
ON inventory_containers (parent_kind, parent_ref_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS inventory_containers_parent_idx;
DROP INDEX IF EXISTS inventory_containers_snapshot_id_idx;
DROP TABLE IF EXISTS inventory_containers;

DROP INDEX IF EXISTS inventory_pods_snapshot_id_idx;
DROP TABLE IF EXISTS inventory_pods;

DROP INDEX IF EXISTS inventory_workloads_snapshot_kind_idx;
DROP INDEX IF EXISTS inventory_workloads_snapshot_id_idx;
DROP TABLE IF EXISTS inventory_workloads;

DROP INDEX IF EXISTS inventory_nodes_snapshot_id_idx;
DROP TABLE IF EXISTS inventory_nodes;

DROP INDEX IF EXISTS inventory_namespaces_snapshot_id_idx;
DROP TABLE IF EXISTS inventory_namespaces;

DROP INDEX IF EXISTS inventory_snapshots_agent_id_collected_at_idx;
DROP TABLE IF EXISTS inventory_snapshots;

-- +goose StatementEnd