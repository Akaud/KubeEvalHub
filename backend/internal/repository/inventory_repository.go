package repository

import (
	"context"
	"errors"

	"backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrInventorySnapshotNotFound = errors.New("inventory snapshot not found")

type InventoryRepository interface {
	InsertSnapshot(ctx context.Context, snapshot *model.InventorySnapshot) error
	InsertNamespaces(ctx context.Context, items []model.NamespaceInventory) error
	InsertNodes(ctx context.Context, items []model.NodeInventory) error
	InsertWorkloads(ctx context.Context, items []model.WorkloadInventory) error
	InsertPods(ctx context.Context, items []model.PodInventory) error
	InsertContainers(ctx context.Context, items []model.ContainerInventory) error
	InsertContainerStatuses(ctx context.Context, items []model.ContainerStatusInventory) error

	GetLatestSnapshotByClusterID(
		ctx context.Context,
		clusterID string,
	) (*model.InventorySnapshot, error)

	GetNamespacesBySnapshotID(
		ctx context.Context,
		snapshotID string,
	) ([]model.NamespaceInventory, error)

	GetNodesBySnapshotID(
		ctx context.Context,
		snapshotID string,
	) ([]model.NodeInventory, error)

	GetWorkloadsBySnapshotID(
		ctx context.Context,
		snapshotID string,
	) ([]model.WorkloadInventory, error)

	GetPodsBySnapshotID(
		ctx context.Context,
		snapshotID string,
	) ([]model.PodInventory, error)

	GetContainersBySnapshotID(
		ctx context.Context,
		snapshotID string,
	) ([]model.ContainerInventory, error)

	GetContainerStatusesBySnapshotID(
		ctx context.Context,
		snapshotID string,
	) ([]model.ContainerStatusInventory, error)
}

type inventoryRepository struct {
	pool *pgxpool.Pool
}

func NewInventoryRepository(pool *pgxpool.Pool) InventoryRepository {
	return &inventoryRepository{pool: pool}
}

func (r *inventoryRepository) InsertSnapshot(ctx context.Context, snapshot *model.InventorySnapshot) error {
	query := `
		INSERT INTO inventory_snapshots (
			id,
			cluster_id,
			collected_at,
			received_at,
			created_at,
			revision_hash
		)
		VALUES ($1,$2,$3,$4,$5,$6)
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		snapshot.ID,
		snapshot.ClusterID,
		snapshot.CollectedAt,
		snapshot.ReceivedAt,
		snapshot.CreatedAt,
		snapshot.RevisionHash,
	)
	return err
}

func (r *inventoryRepository) InsertNamespaces(ctx context.Context, items []model.NamespaceInventory) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO inventory_namespaces (
			id,
			snapshot_id,
			uid,
			name,
			labels_json
		)
		VALUES ($1,$2,$3,$4,$5::jsonb)
	`

	for _, item := range items {
		v := item
		batch.Queue(query, v.ID, v.SnapshotID, v.UID, v.Name, v.LabelsJSON)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *inventoryRepository) InsertNodes(ctx context.Context, items []model.NodeInventory) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO inventory_nodes (
			id,
			snapshot_id,
			uid,
			name,
			labels_json,
			kubelet_version,
			container_runtime_version,
			operating_system,
			architecture,
			kernel_version,
			os_image,
			cpu_capacity_millicores,
			memory_capacity_bytes,
			cpu_allocatable_millicores,
			memory_allocatable_bytes,
			pod_capacity,
			pod_allocatable
		)
		VALUES ($1,$2,$3,$4,$5::jsonb,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17)
	`

	for _, item := range items {
		v := item
		batch.Queue(
			query,
			v.ID,
			v.SnapshotID,
			v.UID,
			v.Name,
			v.LabelsJSON,
			v.KubeletVersion,
			v.ContainerRuntime,
			v.OperatingSystem,
			v.Architecture,
			v.KernelVersion,
			v.OSImage,
			v.CPUCapacityMillicores,
			v.MemoryCapacityBytes,
			v.CPUAllocatableMillicores,
			v.MemoryAllocatableBytes,
			v.PodCapacity,
			v.PodAllocatable,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *inventoryRepository) InsertWorkloads(ctx context.Context, items []model.WorkloadInventory) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO inventory_workloads (
			id,
			snapshot_id,
			kind,
			uid,
			name,
			namespace,
			replicas,
			labels_json
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)
	`

	for _, item := range items {
		v := item
		batch.Queue(
			query,
			v.ID,
			v.SnapshotID,
			v.Kind,
			v.UID,
			v.Name,
			v.Namespace,
			v.Replicas,
			v.LabelsJSON,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *inventoryRepository) InsertPods(ctx context.Context, items []model.PodInventory) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO inventory_pods (
			id,
			snapshot_id,
			uid,
			name,
			namespace,
			node_name,
			phase,
			controller_uid,
			controller_kind,
			controller_name,
			labels_json
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11::jsonb)
	`

	for _, item := range items {
		v := item
		batch.Queue(
			query,
			v.ID,
			v.SnapshotID,
			v.UID,
			v.Name,
			v.Namespace,
			v.NodeName,
			v.Phase,
			v.ControllerUID,
			v.ControllerKind,
			v.ControllerName,
			v.LabelsJSON,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *inventoryRepository) InsertContainers(ctx context.Context, items []model.ContainerInventory) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO inventory_containers (
			id,
			snapshot_id,
			parent_kind,
			parent_ref_id,
			name,
			image,
			cpu_request_millicores,
			cpu_limit_millicores,
			memory_request_bytes,
			memory_limit_bytes,
			is_init_container
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`

	for _, item := range items {
		v := item
		batch.Queue(
			query,
			v.ID,
			v.SnapshotID,
			v.ParentKind,
			v.ParentRefID,
			v.Name,
			v.Image,
			v.CPURequestMillicores,
			v.CPULimitMillicores,
			v.MemoryRequestBytes,
			v.MemoryLimitBytes,
			v.IsInitContainer,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *inventoryRepository) InsertContainerStatuses(ctx context.Context, items []model.ContainerStatusInventory) error {
	if len(items) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO inventory_container_statuses (
			id,
			snapshot_id,
			pod_ref_id,
			name,
			container_id,
			restart_count,
			ready,
			started,
			state,
			last_termination_reason,
			last_termination_exit_code,
			oom_killed,
			is_init_container
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`

	for _, item := range items {
		v := item
		batch.Queue(
			query,
			v.ID,
			v.SnapshotID,
			v.PodRefID,
			v.Name,
			v.ContainerID,
			v.RestartCount,
			v.Ready,
			v.Started,
			v.State,
			v.LastTerminationReason,
			v.LastTerminationExitCode,
			v.OOMKilled,
			v.IsInitContainer,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *inventoryRepository) GetLatestSnapshotByClusterID(
	ctx context.Context,
	clusterID string,
) (*model.InventorySnapshot, error) {
	query := `
		SELECT
			id,
			cluster_id,
			collected_at,
			received_at,
			created_at,
			revision_hash
		FROM inventory_snapshots
		WHERE cluster_id = $1
		ORDER BY collected_at DESC, created_at DESC
		LIMIT 1
	`

	var snapshot model.InventorySnapshot
	err := r.pool.QueryRow(ctx, query, clusterID).Scan(
		&snapshot.ID,
		&snapshot.ClusterID,
		&snapshot.CollectedAt,
		&snapshot.ReceivedAt,
		&snapshot.CreatedAt,
		&snapshot.RevisionHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrInventorySnapshotNotFound
		}
		return nil, err
	}

	return &snapshot, nil
}

func (r *inventoryRepository) GetNamespacesBySnapshotID(
	ctx context.Context,
	snapshotID string,
) ([]model.NamespaceInventory, error) {
	query := `
		SELECT
			id,
			snapshot_id,
			uid,
			name,
			labels_json::text
		FROM inventory_namespaces
		WHERE snapshot_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.NamespaceInventory, 0)
	for rows.Next() {
		var item model.NamespaceInventory
		if err := rows.Scan(
			&item.ID,
			&item.SnapshotID,
			&item.UID,
			&item.Name,
			&item.LabelsJSON,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *inventoryRepository) GetNodesBySnapshotID(
	ctx context.Context,
	snapshotID string,
) ([]model.NodeInventory, error) {
	query := `
		SELECT
			id,
			snapshot_id,
			uid,
			name,
			labels_json::text,
			kubelet_version,
			container_runtime_version,
			operating_system,
			architecture,
			kernel_version,
			os_image,
			cpu_capacity_millicores,
			memory_capacity_bytes,
			cpu_allocatable_millicores,
			memory_allocatable_bytes,
			pod_capacity,
			pod_allocatable
		FROM inventory_nodes
		WHERE snapshot_id = $1
		ORDER BY name ASC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.NodeInventory, 0)
	for rows.Next() {
		var item model.NodeInventory
		if err := rows.Scan(
			&item.ID,
			&item.SnapshotID,
			&item.UID,
			&item.Name,
			&item.LabelsJSON,
			&item.KubeletVersion,
			&item.ContainerRuntime,
			&item.OperatingSystem,
			&item.Architecture,
			&item.KernelVersion,
			&item.OSImage,
			&item.CPUCapacityMillicores,
			&item.MemoryCapacityBytes,
			&item.CPUAllocatableMillicores,
			&item.MemoryAllocatableBytes,
			&item.PodCapacity,
			&item.PodAllocatable,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *inventoryRepository) GetWorkloadsBySnapshotID(
	ctx context.Context,
	snapshotID string,
) ([]model.WorkloadInventory, error) {
	query := `
		SELECT
			id,
			snapshot_id,
			kind,
			uid,
			name,
			namespace,
			replicas,
			labels_json::text
		FROM inventory_workloads
		WHERE snapshot_id = $1
		ORDER BY kind ASC, namespace ASC NULLS FIRST, name ASC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.WorkloadInventory, 0)
	for rows.Next() {
		var item model.WorkloadInventory
		if err := rows.Scan(
			&item.ID,
			&item.SnapshotID,
			&item.Kind,
			&item.UID,
			&item.Name,
			&item.Namespace,
			&item.Replicas,
			&item.LabelsJSON,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *inventoryRepository) GetPodsBySnapshotID(
	ctx context.Context,
	snapshotID string,
) ([]model.PodInventory, error) {
	query := `
		SELECT
			id,
			snapshot_id,
			uid,
			name,
			namespace,
			node_name,
			phase,
			controller_uid,
			controller_kind,
			controller_name,
			labels_json::text
		FROM inventory_pods
		WHERE snapshot_id = $1
		ORDER BY namespace ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.PodInventory, 0)
	for rows.Next() {
		var item model.PodInventory
		if err := rows.Scan(
			&item.ID,
			&item.SnapshotID,
			&item.UID,
			&item.Name,
			&item.Namespace,
			&item.NodeName,
			&item.Phase,
			&item.ControllerUID,
			&item.ControllerKind,
			&item.ControllerName,
			&item.LabelsJSON,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *inventoryRepository) GetContainersBySnapshotID(
	ctx context.Context,
	snapshotID string,
) ([]model.ContainerInventory, error) {
	query := `
		SELECT
			id,
			snapshot_id,
			parent_kind,
			parent_ref_id,
			name,
			image,
			cpu_request_millicores,
			cpu_limit_millicores,
			memory_request_bytes,
			memory_limit_bytes,
			is_init_container
		FROM inventory_containers
		WHERE snapshot_id = $1
		ORDER BY parent_kind ASC, parent_ref_id ASC, is_init_container ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.ContainerInventory, 0)
	for rows.Next() {
		var item model.ContainerInventory
		if err := rows.Scan(
			&item.ID,
			&item.SnapshotID,
			&item.ParentKind,
			&item.ParentRefID,
			&item.Name,
			&item.Image,
			&item.CPURequestMillicores,
			&item.CPULimitMillicores,
			&item.MemoryRequestBytes,
			&item.MemoryLimitBytes,
			&item.IsInitContainer,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *inventoryRepository) GetContainerStatusesBySnapshotID(
	ctx context.Context,
	snapshotID string,
) ([]model.ContainerStatusInventory, error) {
	query := `
		SELECT
			id,
			snapshot_id,
			pod_ref_id,
			name,
			container_id,
			restart_count,
			ready,
			started,
			state,
			last_termination_reason,
			last_termination_exit_code,
			oom_killed,
			is_init_container
		FROM inventory_container_statuses
		WHERE snapshot_id = $1
		ORDER BY pod_ref_id ASC, is_init_container ASC, name ASC
	`

	rows, err := r.pool.Query(ctx, query, snapshotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.ContainerStatusInventory, 0)
	for rows.Next() {
		var item model.ContainerStatusInventory
		if err := rows.Scan(
			&item.ID,
			&item.SnapshotID,
			&item.PodRefID,
			&item.Name,
			&item.ContainerID,
			&item.RestartCount,
			&item.Ready,
			&item.Started,
			&item.State,
			&item.LastTerminationReason,
			&item.LastTerminationExitCode,
			&item.OOMKilled,
			&item.IsInitContainer,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}
