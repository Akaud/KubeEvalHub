package model

import "time"

type PushInventoryRequest struct {
	Cluster     AgentClusterPayload `json:"cluster"`
	CollectedAt time.Time           `json:"collectedAt"`
	Inventory   InventoryPayload    `json:"inventory"`
}

type InventoryPayload struct {
	Namespaces   []NamespacePayload   `json:"namespaces"`
	Nodes        []NodePayload        `json:"nodes"`
	Deployments  []DeploymentPayload  `json:"deployments"`
	StatefulSets []StatefulSetPayload `json:"statefulSets"`
	DaemonSets   []DaemonSetPayload   `json:"daemonSets"`
	Pods         []PodPayload         `json:"pods"`
}

type NamespacePayload struct {
	UID    string            `json:"uid"`
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`
}

type NodePayload struct {
	UID              string            `json:"uid"`
	Name             string            `json:"name"`
	Labels           map[string]string `json:"labels,omitempty"`
	KubeletVersion   string            `json:"kubeletVersion,omitempty"`
	ContainerRuntime string            `json:"containerRuntimeVersion,omitempty"`
	OperatingSystem  string            `json:"operatingSystem,omitempty"`
	Architecture     string            `json:"architecture,omitempty"`
	KernelVersion    string            `json:"kernelVersion,omitempty"`
	OSImage          string            `json:"osImage,omitempty"`
}

type DeploymentPayload struct {
	UID        string                 `json:"uid"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Replicas   int32                  `json:"replicas"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Containers []ContainerSpecPayload `json:"containers,omitempty"`
}

type StatefulSetPayload struct {
	UID        string                 `json:"uid"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Replicas   int32                  `json:"replicas"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Containers []ContainerSpecPayload `json:"containers,omitempty"`
}

type DaemonSetPayload struct {
	UID        string                 `json:"uid"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	Labels     map[string]string      `json:"labels,omitempty"`
	Containers []ContainerSpecPayload `json:"containers,omitempty"`
}

type PodPayload struct {
	UID        string                 `json:"uid"`
	Name       string                 `json:"name"`
	Namespace  string                 `json:"namespace"`
	NodeName   string                 `json:"nodeName,omitempty"`
	Phase      string                 `json:"phase,omitempty"`
	Labels     map[string]string      `json:"labels,omitempty"`
	OwnerKind  string                 `json:"ownerKind,omitempty"`
	OwnerName  string                 `json:"ownerName,omitempty"`
	Containers []ContainerSpecPayload `json:"containers"`
}

type ContainerSpecPayload struct {
	Name                 string `json:"name"`
	Image                string `json:"image,omitempty"`
	CPURequestMillicores *int64 `json:"cpuRequestMillicores,omitempty"`
	CPULimitMillicores   *int64 `json:"cpuLimitMillicores,omitempty"`
	MemoryRequestBytes   *int64 `json:"memoryRequestBytes,omitempty"`
	MemoryLimitBytes     *int64 `json:"memoryLimitBytes,omitempty"`
}
