package model

import "time"

type PushMetricsRequest struct {
	Cluster ClusterPayload       `json:"cluster"`
	Samples []MetricPointPayload `json:"samples"`
}

type ClusterPayload struct {
	ClusterUID    string `json:"clusterUid"`
	ClusterName   string `json:"clusterName"`
	KubeVersion   string `json:"kubeVersion"`
	Distribution  string `json:"distribution"`
	APIServerHost string `json:"apiServerHost"`
}

type MetricPointPayload struct {
	MetricName    string            `json:"metricName"`
	MetricType    string            `json:"metricType"`
	Unit          string            `json:"unit"`
	ResourceKind  string            `json:"resourceKind"`
	NodeName      string            `json:"nodeName,omitempty"`
	Namespace     string            `json:"namespace,omitempty"`
	PodName       string            `json:"podName,omitempty"`
	ContainerName string            `json:"containerName,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	CollectedAt   time.Time         `json:"collectedAt"`
	Value         float64           `json:"value"`
}
