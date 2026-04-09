package model

import "time"

type PushMetricsRequest struct {
	Cluster          AgentClusterPayload  `json:"cluster"`
	BatchCollectedAt time.Time            `json:"batchCollectedAt"`
	Source           string               `json:"source"`
	Samples          []MetricPointPayload `json:"samples"`
}

type AgentClusterPayload struct {
	ClusterUID    string `json:"clusterUid"`
	ClusterName   string `json:"clusterName"`
	KubeVersion   string `json:"kubeVersion"`
	Distribution  string `json:"distribution"`
	APIServerHost string `json:"apiServerHost"`
}

type MetricPointPayload struct {
	MetricName   string `json:"metricName"`
	MetricType   string `json:"metricType"`
	Unit         string `json:"unit"`
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
