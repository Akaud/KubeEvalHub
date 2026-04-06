package model

import "time"

type MetricSeries struct {
	ID            string    `json:"id"`
	AgentID       string    `json:"agentId"`
	MetricName    string    `json:"metricName"`
	MetricType    string    `json:"metricType"`   // e.g. "gauge"
	Unit          string    `json:"unit"`         // e.g. "cores", "bytes"
	ResourceKind  string    `json:"resourceKind"` // e.g. "node", "pod"
	NodeName      *string   `json:"nodeName,omitempty"`
	Namespace     *string   `json:"namespace,omitempty"`
	PodName       *string   `json:"podName,omitempty"`
	ContainerName *string   `json:"containerName,omitempty"`
	LabelsHash    string    `json:"labelsHash"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
