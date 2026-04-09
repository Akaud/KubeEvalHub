package model

import "time"

type PushMetricsRequest struct {
	Cluster          ClusterPayload       `json:"cluster"`
	BatchCollectedAt time.Time            `json:"batchCollectedAt"`
	Source           string               `json:"source"` // e.g. "metrics-server"
	Samples          []MetricPointPayload `json:"samples"`
}

type PushInventoryRequest struct {
	Cluster      ClusterPayload   `json:"cluster"`
	CollectedAt  time.Time        `json:"collectedAt"`
	RevisionHash string           `json:"revisionHash,omitempty"` // stable hash of normalized inventory content
	Inventory    InventoryPayload `json:"inventory"`
}

type ClusterPayload struct {
	ClusterUID    string `json:"clusterUid"`
	ClusterName   string `json:"clusterName"`
	KubeVersion   string `json:"kubeVersion"`
	Distribution  string `json:"distribution"`
	APIServerHost string `json:"apiServerHost"`
}

type MetricPointPayload struct {
	MetricName string `json:"metricName"`
	MetricType string `json:"metricType"` // gauge
	Unit       string `json:"unit"`       // cores, bytes
	// node, pod, container
	ResourceKind string `json:"resourceKind"`

	NodeName string `json:"nodeName,omitempty"`

	Namespace     string `json:"namespace,omitempty"`
	PodName       string `json:"podName,omitempty"`
	PodUID        string `json:"podUid,omitempty"`
	ContainerName string `json:"containerName,omitempty"`

	ControllerUID  string `json:"controllerUid,omitempty"`
	ControllerKind string `json:"controllerKind,omitempty"`
	ControllerName string `json:"controllerName,omitempty"`

	CollectedAt time.Time `json:"collectedAt"`
	Value       float64   `json:"value"`
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
	UID    string            `json:"uid"`
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels,omitempty"`

	KubeletVersion   string `json:"kubeletVersion,omitempty"`
	ContainerRuntime string `json:"containerRuntimeVersion,omitempty"`
	OperatingSystem  string `json:"operatingSystem,omitempty"`
	Architecture     string `json:"architecture,omitempty"`
	KernelVersion    string `json:"kernelVersion,omitempty"`
	OSImage          string `json:"osImage,omitempty"`

	CPUCapacityMillicores    *int64 `json:"cpuCapacityMillicores,omitempty"`
	MemoryCapacityBytes      *int64 `json:"memoryCapacityBytes,omitempty"`
	CPUAllocatableMillicores *int64 `json:"cpuAllocatableMillicores,omitempty"`
	MemoryAllocatableBytes   *int64 `json:"memoryAllocatableBytes,omitempty"`
	PodCapacity              *int64 `json:"podCapacity,omitempty"`
	PodAllocatable           *int64 `json:"podAllocatable,omitempty"`
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
	UID       string            `json:"uid"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	NodeName  string            `json:"nodeName,omitempty"`
	Phase     string            `json:"phase,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`

	ControllerUID  string `json:"controllerUid,omitempty"`
	ControllerKind string `json:"controllerKind,omitempty"`
	ControllerName string `json:"controllerName,omitempty"`

	Containers        []ContainerSpecPayload   `json:"containers,omitempty"`
	ContainerStatuses []ContainerStatusPayload `json:"containerStatuses,omitempty"`
}

type ContainerSpecPayload struct {
	Name  string `json:"name"`
	Image string `json:"image,omitempty"`

	CPURequestMillicores *int64 `json:"cpuRequestMillicores,omitempty"`
	CPULimitMillicores   *int64 `json:"cpuLimitMillicores,omitempty"`
	MemoryRequestBytes   *int64 `json:"memoryRequestBytes,omitempty"`
	MemoryLimitBytes     *int64 `json:"memoryLimitBytes,omitempty"`

	IsInitContainer bool `json:"isInitContainer,omitempty"`
}

type ContainerStatusPayload struct {
	Name        string `json:"name"`
	ContainerID string `json:"containerId,omitempty"`

	RestartCount int32 `json:"restartCount"`
	Ready        bool  `json:"ready"`
	Started      *bool `json:"started,omitempty"`

	// running, waiting, terminated
	State string `json:"state,omitempty"`

	LastTerminationReason   string `json:"lastTerminationReason,omitempty"`
	LastTerminationExitCode *int32 `json:"lastTerminationExitCode,omitempty"`
	OOMKilled               bool   `json:"oomKilled,omitempty"`

	IsInitContainer bool `json:"isInitContainer,omitempty"`
}
