package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var ErrInvalidInventoryPayload = errors.New("invalid inventory payload")

type InventoryService interface {
	IngestInventory(ctx context.Context, agentID string, req *model.PushInventoryRequest) error
	GetLatestInventory(ctx context.Context, ownerID int64, agentID string) (*model.InventorySnapshotResponse, error)
}

type inventoryService struct {
	clusterRepo   repository.ClusterRepository
	inventoryRepo repository.InventoryRepository
}

func NewInventoryService(
	clusterRepo repository.ClusterRepository,
	inventoryRepo repository.InventoryRepository,
) InventoryService {
	return &inventoryService{
		clusterRepo:   clusterRepo,
		inventoryRepo: inventoryRepo,
	}
}

func (s *inventoryService) IngestInventory(ctx context.Context, agentID string, req *model.PushInventoryRequest) error {
	if req == nil {
		return ErrInvalidInventoryPayload
	}
	if strings.TrimSpace(agentID) == "" {
		return ErrInvalidInventoryPayload
	}
	if req.CollectedAt.IsZero() {
		return ErrInvalidInventoryPayload
	}
	if strings.TrimSpace(req.Cluster.ClusterUID) == "" ||
		strings.TrimSpace(req.Cluster.ClusterName) == "" ||
		strings.TrimSpace(req.Cluster.KubeVersion) == "" {
		return ErrInvalidInventoryPayload
	}

	now := time.Now().UTC()

	existingCluster, err := s.clusterRepo.GetByAgentID(ctx, agentID)
	if err != nil && !errors.Is(err, repository.ErrClusterNotFound) {
		return err
	}
	if existingCluster != nil && existingCluster.ClusterUID != req.Cluster.ClusterUID {
		return ErrClusterUIDMismatch
	}

	distribution := strings.TrimSpace(req.Cluster.Distribution)
	if distribution == "" {
		distribution = "unknown"
	}

	cluster := &model.AgentCluster{
		AgentID:       agentID,
		ClusterUID:    strings.TrimSpace(req.Cluster.ClusterUID),
		ClusterName:   strings.TrimSpace(req.Cluster.ClusterName),
		KubeVersion:   strings.TrimSpace(req.Cluster.KubeVersion),
		Distribution:  distribution,
		APIServerHost: strings.TrimSpace(req.Cluster.APIServerHost),
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if existingCluster != nil {
		cluster.CreatedAt = existingCluster.CreatedAt
	}

	if err := s.clusterRepo.Upsert(ctx, cluster); err != nil {
		return err
	}

	snapshot := &model.InventorySnapshot{
		ID:           uuid.NewString(),
		AgentID:      agentID,
		CollectedAt:  req.CollectedAt,
		ReceivedAt:   now,
		CreatedAt:    now,
		RevisionHash: stringPtrOrNil(req.RevisionHash),
	}
	if err := s.inventoryRepo.InsertSnapshot(ctx, snapshot); err != nil {
		return err
	}

	namespaces := make([]model.NamespaceInventory, 0, len(req.Inventory.Namespaces))
	for _, item := range req.Inventory.Namespaces {
		if strings.TrimSpace(item.UID) == "" || strings.TrimSpace(item.Name) == "" {
			return ErrInvalidInventoryPayload
		}
		namespaces = append(namespaces, model.NamespaceInventory{
			ID:         uuid.NewString(),
			SnapshotID: snapshot.ID,
			UID:        strings.TrimSpace(item.UID),
			Name:       strings.TrimSpace(item.Name),
			LabelsJSON: mustJSONPtr(item.Labels),
		})
	}
	if err := s.inventoryRepo.InsertNamespaces(ctx, namespaces); err != nil {
		return err
	}

	nodes := make([]model.NodeInventory, 0, len(req.Inventory.Nodes))
	for _, item := range req.Inventory.Nodes {
		if strings.TrimSpace(item.UID) == "" || strings.TrimSpace(item.Name) == "" {
			return ErrInvalidInventoryPayload
		}
		nodes = append(nodes, model.NodeInventory{
			ID:                       uuid.NewString(),
			SnapshotID:               snapshot.ID,
			UID:                      strings.TrimSpace(item.UID),
			Name:                     strings.TrimSpace(item.Name),
			LabelsJSON:               mustJSONPtr(item.Labels),
			KubeletVersion:           stringPtrOrNil(item.KubeletVersion),
			ContainerRuntime:         stringPtrOrNil(item.ContainerRuntime),
			OperatingSystem:          stringPtrOrNil(item.OperatingSystem),
			Architecture:             stringPtrOrNil(item.Architecture),
			KernelVersion:            stringPtrOrNil(item.KernelVersion),
			OSImage:                  stringPtrOrNil(item.OSImage),
			CPUCapacityMillicores:    item.CPUCapacityMillicores,
			MemoryCapacityBytes:      item.MemoryCapacityBytes,
			CPUAllocatableMillicores: item.CPUAllocatableMillicores,
			MemoryAllocatableBytes:   item.MemoryAllocatableBytes,
			PodCapacity:              item.PodCapacity,
			PodAllocatable:           item.PodAllocatable,
		})
	}
	if err := s.inventoryRepo.InsertNodes(ctx, nodes); err != nil {
		return err
	}

	workloads := make([]model.WorkloadInventory, 0,
		len(req.Inventory.Deployments)+len(req.Inventory.StatefulSets)+len(req.Inventory.DaemonSets),
	)
	containers := make([]model.ContainerInventory, 0)
	containerStatuses := make([]model.ContainerStatusInventory, 0)

	for _, item := range req.Inventory.Deployments {
		if strings.TrimSpace(item.UID) == "" || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Namespace) == "" {
			return ErrInvalidInventoryPayload
		}
		workloadID := uuid.NewString()
		workloads = append(workloads, model.WorkloadInventory{
			ID:         workloadID,
			SnapshotID: snapshot.ID,
			Kind:       "deployment",
			UID:        strings.TrimSpace(item.UID),
			Name:       strings.TrimSpace(item.Name),
			Namespace:  stringPtrOrNil(item.Namespace),
			Replicas:   int32Ptr(item.Replicas),
			LabelsJSON: mustJSONPtr(item.Labels),
		})
		containers = append(containers, toContainerInventory(snapshot.ID, "deployment", workloadID, item.Containers)...)
	}

	for _, item := range req.Inventory.StatefulSets {
		if strings.TrimSpace(item.UID) == "" || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Namespace) == "" {
			return ErrInvalidInventoryPayload
		}
		workloadID := uuid.NewString()
		workloads = append(workloads, model.WorkloadInventory{
			ID:         workloadID,
			SnapshotID: snapshot.ID,
			Kind:       "statefulset",
			UID:        strings.TrimSpace(item.UID),
			Name:       strings.TrimSpace(item.Name),
			Namespace:  stringPtrOrNil(item.Namespace),
			Replicas:   int32Ptr(item.Replicas),
			LabelsJSON: mustJSONPtr(item.Labels),
		})
		containers = append(containers, toContainerInventory(snapshot.ID, "statefulset", workloadID, item.Containers)...)
	}

	for _, item := range req.Inventory.DaemonSets {
		if strings.TrimSpace(item.UID) == "" || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Namespace) == "" {
			return ErrInvalidInventoryPayload
		}
		workloadID := uuid.NewString()
		workloads = append(workloads, model.WorkloadInventory{
			ID:         workloadID,
			SnapshotID: snapshot.ID,
			Kind:       "daemonset",
			UID:        strings.TrimSpace(item.UID),
			Name:       strings.TrimSpace(item.Name),
			Namespace:  stringPtrOrNil(item.Namespace),
			Replicas:   nil,
			LabelsJSON: mustJSONPtr(item.Labels),
		})
		containers = append(containers, toContainerInventory(snapshot.ID, "daemonset", workloadID, item.Containers)...)
	}

	if err := s.inventoryRepo.InsertWorkloads(ctx, workloads); err != nil {
		return err
	}

	pods := make([]model.PodInventory, 0, len(req.Inventory.Pods))
	for _, item := range req.Inventory.Pods {
		if strings.TrimSpace(item.UID) == "" || strings.TrimSpace(item.Name) == "" || strings.TrimSpace(item.Namespace) == "" {
			return ErrInvalidInventoryPayload
		}

		podID := uuid.NewString()
		pods = append(pods, model.PodInventory{
			ID:             podID,
			SnapshotID:     snapshot.ID,
			UID:            strings.TrimSpace(item.UID),
			Name:           strings.TrimSpace(item.Name),
			Namespace:      strings.TrimSpace(item.Namespace),
			NodeName:       stringPtrOrNil(item.NodeName),
			Phase:          stringPtrOrNil(item.Phase),
			ControllerUID:  stringPtrOrNil(item.ControllerUID),
			ControllerKind: stringPtrOrNil(item.ControllerKind),
			ControllerName: stringPtrOrNil(item.ControllerName),
			LabelsJSON:     mustJSONPtr(item.Labels),
		})

		containers = append(containers, toContainerInventory(snapshot.ID, "pod", podID, item.Containers)...)
		containerStatuses = append(containerStatuses, toContainerStatusInventory(snapshot.ID, podID, item.ContainerStatuses)...)
	}

	if err := s.inventoryRepo.InsertPods(ctx, pods); err != nil {
		return err
	}

	if err := s.inventoryRepo.InsertContainers(ctx, containers); err != nil {
		return err
	}

	if err := s.inventoryRepo.InsertContainerStatuses(ctx, containerStatuses); err != nil {
		return err
	}

	return nil
}

func toContainerInventory(
	snapshotID string,
	parentKind string,
	parentRefID string,
	items []model.ContainerSpecPayload,
) []model.ContainerInventory {
	out := make([]model.ContainerInventory, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		out = append(out, model.ContainerInventory{
			ID:                   uuid.NewString(),
			SnapshotID:           snapshotID,
			ParentKind:           parentKind,
			ParentRefID:          parentRefID,
			Name:                 strings.TrimSpace(item.Name),
			Image:                stringPtrOrNil(item.Image),
			CPURequestMillicores: item.CPURequestMillicores,
			CPULimitMillicores:   item.CPULimitMillicores,
			MemoryRequestBytes:   item.MemoryRequestBytes,
			MemoryLimitBytes:     item.MemoryLimitBytes,
			IsInitContainer:      item.IsInitContainer,
		})
	}
	return out
}

func toContainerStatusInventory(
	snapshotID string,
	podRefID string,
	items []model.ContainerStatusPayload,
) []model.ContainerStatusInventory {
	out := make([]model.ContainerStatusInventory, 0, len(items))
	for _, item := range items {
		if strings.TrimSpace(item.Name) == "" {
			continue
		}
		out = append(out, model.ContainerStatusInventory{
			ID:                      uuid.NewString(),
			SnapshotID:              snapshotID,
			PodRefID:                podRefID,
			Name:                    strings.TrimSpace(item.Name),
			ContainerID:             stringPtrOrNil(item.ContainerID),
			RestartCount:            item.RestartCount,
			Ready:                   item.Ready,
			Started:                 item.Started,
			State:                   stringPtrOrNil(item.State),
			LastTerminationReason:   stringPtrOrNil(item.LastTerminationReason),
			LastTerminationExitCode: item.LastTerminationExitCode,
			OOMKilled:               item.OOMKilled,
			IsInitContainer:         item.IsInitContainer,
		})
	}
	return out
}

func mustJSONPtr(v map[string]string) *string {
	if len(v) == 0 {
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	s := string(b)
	return &s
}

func int32Ptr(v int32) *int32 {
	return &v
}

func (s *inventoryService) GetLatestInventory(
	ctx context.Context,
	ownerID int64,
	agentID string,
) (*model.InventorySnapshotResponse, error) {
	if ownerID <= 0 || strings.TrimSpace(agentID) == "" {
		return nil, ErrInvalidInventoryPayload
	}

	snapshot, err := s.inventoryRepo.GetLatestSnapshotForOwner(ctx, ownerID, agentID)
	if err != nil {
		return nil, err
	}

	namespaces, err := s.inventoryRepo.GetNamespacesBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	nodes, err := s.inventoryRepo.GetNodesBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	workloads, err := s.inventoryRepo.GetWorkloadsBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	pods, err := s.inventoryRepo.GetPodsBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	containers, err := s.inventoryRepo.GetContainersBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	containerStatuses, err := s.inventoryRepo.GetContainerStatusesBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	containerMap := make(map[string][]model.ContainerInventory)
	for _, c := range containers {
		containerMap[c.ParentRefID] = append(containerMap[c.ParentRefID], c)
	}

	containerStatusMap := make(map[string][]model.ContainerStatusInventory)
	for _, st := range containerStatuses {
		containerStatusMap[st.PodRefID] = append(containerStatusMap[st.PodRefID], st)
	}

	resp := &model.InventorySnapshotResponse{
		ClusterID:   agentID,
		CollectedAt: snapshot.CollectedAt,
		Inventory: model.InventoryPayload{
			Namespaces:   make([]model.NamespacePayload, 0, len(namespaces)),
			Nodes:        make([]model.NodePayload, 0, len(nodes)),
			Deployments:  make([]model.DeploymentPayload, 0),
			StatefulSets: make([]model.StatefulSetPayload, 0),
			DaemonSets:   make([]model.DaemonSetPayload, 0),
			Pods:         make([]model.PodPayload, 0, len(pods)),
		},
	}

	for _, item := range namespaces {
		resp.Inventory.Namespaces = append(resp.Inventory.Namespaces, model.NamespacePayload{
			UID:    item.UID,
			Name:   item.Name,
			Labels: parseLabelsJSON(item.LabelsJSON),
		})
	}

	for _, item := range nodes {
		resp.Inventory.Nodes = append(resp.Inventory.Nodes, model.NodePayload{
			UID:                      item.UID,
			Name:                     item.Name,
			Labels:                   parseLabelsJSON(item.LabelsJSON),
			KubeletVersion:           derefString(item.KubeletVersion),
			ContainerRuntime:         derefString(item.ContainerRuntime),
			OperatingSystem:          derefString(item.OperatingSystem),
			Architecture:             derefString(item.Architecture),
			KernelVersion:            derefString(item.KernelVersion),
			OSImage:                  derefString(item.OSImage),
			CPUCapacityMillicores:    item.CPUCapacityMillicores,
			MemoryCapacityBytes:      item.MemoryCapacityBytes,
			CPUAllocatableMillicores: item.CPUAllocatableMillicores,
			MemoryAllocatableBytes:   item.MemoryAllocatableBytes,
			PodCapacity:              item.PodCapacity,
			PodAllocatable:           item.PodAllocatable,
		})
	}

	for _, item := range workloads {
		containers := toContainerSpecPayloads(containerMap[item.ID])

		switch strings.ToLower(strings.TrimSpace(item.Kind)) {
		case "deployment":
			resp.Inventory.Deployments = append(resp.Inventory.Deployments, model.DeploymentPayload{
				UID:        item.UID,
				Name:       item.Name,
				Namespace:  derefString(item.Namespace),
				Replicas:   derefInt32(item.Replicas),
				Labels:     parseLabelsJSON(item.LabelsJSON),
				Containers: containers,
			})
		case "statefulset":
			resp.Inventory.StatefulSets = append(resp.Inventory.StatefulSets, model.StatefulSetPayload{
				UID:        item.UID,
				Name:       item.Name,
				Namespace:  derefString(item.Namespace),
				Replicas:   derefInt32(item.Replicas),
				Labels:     parseLabelsJSON(item.LabelsJSON),
				Containers: containers,
			})
		case "daemonset":
			resp.Inventory.DaemonSets = append(resp.Inventory.DaemonSets, model.DaemonSetPayload{
				UID:        item.UID,
				Name:       item.Name,
				Namespace:  derefString(item.Namespace),
				Labels:     parseLabelsJSON(item.LabelsJSON),
				Containers: containers,
			})
		}
	}

	for _, item := range pods {
		resp.Inventory.Pods = append(resp.Inventory.Pods, model.PodPayload{
			UID:               item.UID,
			Name:              item.Name,
			Namespace:         item.Namespace,
			NodeName:          derefString(item.NodeName),
			Phase:             derefString(item.Phase),
			Labels:            parseLabelsJSON(item.LabelsJSON),
			ControllerUID:     derefString(item.ControllerUID),
			ControllerKind:    derefString(item.ControllerKind),
			ControllerName:    derefString(item.ControllerName),
			Containers:        toContainerSpecPayloads(containerMap[item.ID]),
			ContainerStatuses: toContainerStatusPayloads(containerStatusMap[item.ID]),
		})
	}

	return resp, nil
}

func parseLabelsJSON(raw *string) map[string]string {
	if raw == nil || strings.TrimSpace(*raw) == "" {
		return nil
	}

	var out map[string]string
	if err := json.Unmarshal([]byte(*raw), &out); err != nil {
		return nil
	}
	return out
}

func toContainerSpecPayloads(items []model.ContainerInventory) []model.ContainerSpecPayload {
	if len(items) == 0 {
		return nil
	}

	out := make([]model.ContainerSpecPayload, 0, len(items))
	for _, item := range items {
		out = append(out, model.ContainerSpecPayload{
			Name:                 item.Name,
			Image:                derefString(item.Image),
			CPURequestMillicores: item.CPURequestMillicores,
			CPULimitMillicores:   item.CPULimitMillicores,
			MemoryRequestBytes:   item.MemoryRequestBytes,
			MemoryLimitBytes:     item.MemoryLimitBytes,
			IsInitContainer:      item.IsInitContainer,
		})
	}
	return out
}

func toContainerStatusPayloads(items []model.ContainerStatusInventory) []model.ContainerStatusPayload {
	if len(items) == 0 {
		return nil
	}

	out := make([]model.ContainerStatusPayload, 0, len(items))
	for _, item := range items {
		out = append(out, model.ContainerStatusPayload{
			Name:                    item.Name,
			ContainerID:             derefString(item.ContainerID),
			RestartCount:            item.RestartCount,
			Ready:                   item.Ready,
			Started:                 item.Started,
			State:                   derefString(item.State),
			LastTerminationReason:   derefString(item.LastTerminationReason),
			LastTerminationExitCode: item.LastTerminationExitCode,
			OOMKilled:               item.OOMKilled,
			IsInitContainer:         item.IsInitContainer,
		})
	}
	return out
}

func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}
