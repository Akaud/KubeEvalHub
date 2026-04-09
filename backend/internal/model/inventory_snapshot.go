package model

import "time"

type InventorySnapshot struct {
	ID           string    `json:"id"`
	AgentID      string    `json:"agentId"`
	CollectedAt  time.Time `json:"collectedAt"`
	ReceivedAt   time.Time `json:"receivedAt"`
	CreatedAt    time.Time `json:"createdAt"`
	RevisionHash *string   `json:"revisionHash,omitempty"`
}

type NamespaceInventory struct {
	ID         string  `json:"id"`
	SnapshotID string  `json:"snapshotId"`
	UID        string  `json:"uid"`
	Name       string  `json:"name"`
	LabelsJSON *string `json:"labelsJson,omitempty"`
}

type NodeInventory struct {
	ID         string  `json:"id"`
	SnapshotID string  `json:"snapshotId"`
	UID        string  `json:"uid"`
	Name       string  `json:"name"`
	LabelsJSON *string `json:"labelsJson,omitempty"`

	KubeletVersion   *string `json:"kubeletVersion,omitempty"`
	ContainerRuntime *string `json:"containerRuntimeVersion,omitempty"`
	OperatingSystem  *string `json:"operatingSystem,omitempty"`
	Architecture     *string `json:"architecture,omitempty"`
	KernelVersion    *string `json:"kernelVersion,omitempty"`
	OSImage          *string `json:"osImage,omitempty"`

	CPUCapacityMillicores    *int64 `json:"cpuCapacityMillicores,omitempty"`
	MemoryCapacityBytes      *int64 `json:"memoryCapacityBytes,omitempty"`
	CPUAllocatableMillicores *int64 `json:"cpuAllocatableMillicores,omitempty"`
	MemoryAllocatableBytes   *int64 `json:"memoryAllocatableBytes,omitempty"`
	PodCapacity              *int64 `json:"podCapacity,omitempty"`
	PodAllocatable           *int64 `json:"podAllocatable,omitempty"`
}

type WorkloadInventory struct {
	ID         string  `json:"id"`
	SnapshotID string  `json:"snapshotId"`
	Kind       string  `json:"kind"`
	UID        string  `json:"uid"`
	Name       string  `json:"name"`
	Namespace  *string `json:"namespace,omitempty"`
	Replicas   *int32  `json:"replicas,omitempty"`
	LabelsJSON *string `json:"labelsJson,omitempty"`
}

type PodInventory struct {
	ID         string  `json:"id"`
	SnapshotID string  `json:"snapshotId"`
	UID        string  `json:"uid"`
	Name       string  `json:"name"`
	Namespace  string  `json:"namespace"`
	NodeName   *string `json:"nodeName,omitempty"`
	Phase      *string `json:"phase,omitempty"`

	ControllerUID  *string `json:"controllerUid,omitempty"`
	ControllerKind *string `json:"controllerKind,omitempty"`
	ControllerName *string `json:"controllerName,omitempty"`

	LabelsJSON *string `json:"labelsJson,omitempty"`
}

type ContainerInventory struct {
	ID                   string  `json:"id"`
	SnapshotID           string  `json:"snapshotId"`
	ParentKind           string  `json:"parentKind"`
	ParentRefID          string  `json:"parentRefId"`
	Name                 string  `json:"name"`
	Image                *string `json:"image,omitempty"`
	CPURequestMillicores *int64  `json:"cpuRequestMillicores,omitempty"`
	CPULimitMillicores   *int64  `json:"cpuLimitMillicores,omitempty"`
	MemoryRequestBytes   *int64  `json:"memoryRequestBytes,omitempty"`
	MemoryLimitBytes     *int64  `json:"memoryLimitBytes,omitempty"`
	IsInitContainer      bool    `json:"isInitContainer"`
}

type ContainerStatusInventory struct {
	ID         string `json:"id"`
	SnapshotID string `json:"snapshotId"`
	PodRefID   string `json:"podRefId"`
	Name       string `json:"name"`

	ContainerID  *string `json:"containerId,omitempty"`
	RestartCount int32   `json:"restartCount"`
	Ready        bool    `json:"ready"`
	Started      *bool   `json:"started,omitempty"`
	State        *string `json:"state,omitempty"`

	LastTerminationReason   *string `json:"lastTerminationReason,omitempty"`
	LastTerminationExitCode *int32  `json:"lastTerminationExitCode,omitempty"`
	OOMKilled               bool    `json:"oomKilled"`

	IsInitContainer bool `json:"isInitContainer"`
}
