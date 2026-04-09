-- +goose Up
-- +goose StatementBegin

CREATE TABLE inventory_snapshots (
    id uuid PRIMARY KEY,
    agent_id uuid NOT NULL REFERENCES agents(id) ON DELETE CASCADE,
    collected_at timestamptz NOT NULL,
    received_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL,
    revision_hash text
);

CREATE INDEX inventory_snapshots_agent_id_collected_at_idx
ON inventory_snapshots (agent_id, collected_at DESC);

CREATE INDEX inventory_snapshots_agent_revision_idx
ON inventory_snapshots (agent_id, revision_hash);

CREATE TABLE inventory_namespaces (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES inventory_snapshots(id) ON DELETE CASCADE,
    uid text NOT NULL,
    name text NOT NULL,
    labels_json jsonb
);

CREATE UNIQUE INDEX inventory_namespaces_snapshot_uid_uidx
ON inventory_namespaces (snapshot_id, uid);

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
    os_image text,
    cpu_capacity_millicores bigint,
    memory_capacity_bytes bigint,
    cpu_allocatable_millicores bigint,
    memory_allocatable_bytes bigint,
    pod_capacity bigint,
    pod_allocatable bigint
);

CREATE UNIQUE INDEX inventory_nodes_snapshot_uid_uidx
ON inventory_nodes (snapshot_id, uid);

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

CREATE UNIQUE INDEX inventory_workloads_snapshot_uid_uidx
ON inventory_workloads (snapshot_id, uid);

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
    controller_uid text,
    controller_kind text,
    controller_name text,
    labels_json jsonb
);

CREATE UNIQUE INDEX inventory_pods_snapshot_uid_uidx
ON inventory_pods (snapshot_id, uid);

CREATE INDEX inventory_pods_snapshot_id_idx
ON inventory_pods (snapshot_id);

CREATE INDEX inventory_pods_snapshot_controller_idx
ON inventory_pods (snapshot_id, controller_uid);

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
    memory_limit_bytes bigint,
    is_init_container boolean NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX inventory_containers_parent_name_init_uidx
ON inventory_containers (parent_kind, parent_ref_id, name, is_init_container);

CREATE INDEX inventory_containers_snapshot_id_idx
ON inventory_containers (snapshot_id);

CREATE INDEX inventory_containers_parent_idx
ON inventory_containers (parent_kind, parent_ref_id);

CREATE TABLE inventory_container_statuses (
    id uuid PRIMARY KEY,
    snapshot_id uuid NOT NULL REFERENCES inventory_snapshots(id) ON DELETE CASCADE,
    pod_ref_id uuid NOT NULL,
    name text NOT NULL,
    container_id text,
    restart_count integer NOT NULL DEFAULT 0,
    ready boolean NOT NULL DEFAULT FALSE,
    started boolean,
    state text,
    last_termination_reason text,
    last_termination_exit_code integer,
    oom_killed boolean NOT NULL DEFAULT FALSE,
    is_init_container boolean NOT NULL DEFAULT FALSE
);

CREATE UNIQUE INDEX inventory_container_statuses_pod_name_init_uidx
ON inventory_container_statuses (pod_ref_id, name, is_init_container);

CREATE INDEX inventory_container_statuses_snapshot_id_idx
ON inventory_container_statuses (snapshot_id);

CREATE INDEX inventory_container_statuses_pod_ref_idx
ON inventory_container_statuses (pod_ref_id);

CREATE INDEX inventory_container_statuses_oom_idx
ON inventory_container_statuses (oom_killed);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS inventory_container_statuses_oom_idx;
DROP INDEX IF EXISTS inventory_container_statuses_pod_ref_idx;
DROP INDEX IF EXISTS inventory_container_statuses_snapshot_id_idx;
DROP INDEX IF EXISTS inventory_container_statuses_pod_name_init_uidx;
DROP TABLE IF EXISTS inventory_container_statuses;

DROP INDEX IF EXISTS inventory_containers_parent_idx;
DROP INDEX IF EXISTS inventory_containers_snapshot_id_idx;
DROP INDEX IF EXISTS inventory_containers_parent_name_init_uidx;
DROP TABLE IF EXISTS inventory_containers;

DROP INDEX IF EXISTS inventory_pods_snapshot_controller_idx;
DROP INDEX IF EXISTS inventory_pods_snapshot_id_idx;
DROP INDEX IF EXISTS inventory_pods_snapshot_uid_uidx;
DROP TABLE IF EXISTS inventory_pods;

DROP INDEX IF EXISTS inventory_workloads_snapshot_kind_idx;
DROP INDEX IF EXISTS inventory_workloads_snapshot_id_idx;
DROP INDEX IF EXISTS inventory_workloads_snapshot_uid_uidx;
DROP TABLE IF EXISTS inventory_workloads;

DROP INDEX IF EXISTS inventory_nodes_snapshot_id_idx;
DROP INDEX IF EXISTS inventory_nodes_snapshot_uid_uidx;
DROP TABLE IF EXISTS inventory_nodes;

DROP INDEX IF EXISTS inventory_namespaces_snapshot_id_idx;
DROP INDEX IF EXISTS inventory_namespaces_snapshot_uid_uidx;
DROP TABLE IF EXISTS inventory_namespaces;

DROP INDEX IF EXISTS inventory_snapshots_agent_revision_idx;
DROP INDEX IF EXISTS inventory_snapshots_agent_id_collected_at_idx;
DROP TABLE IF EXISTS inventory_snapshots;

-- +goose StatementEnd