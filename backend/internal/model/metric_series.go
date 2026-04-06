package model

import "time"

type MetricSeries struct {
	ID            string    `json:"id"`
	AgentID       string    `json:"agentId"`
	MetricName    string    `json:"metricName"`
	MetricType    string    `json:"metricType"`
	Unit          string    `json:"unit"`
	ResourceKind  string    `json:"resourceKind"`
	NodeName      *string   `json:"nodeName,omitempty"`
	Namespace     *string   `json:"namespace,omitempty"`
	PodName       *string   `json:"podName,omitempty"`
	ContainerName *string   `json:"containerName,omitempty"`
	LabelsHash    string    `json:"labelsHash"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type MetricSeriesPoint struct {
	CollectedAt time.Time `json:"collectedAt"`
	Value       float64   `json:"value"`
}

type MetricSeriesWithSamples struct {
	Series  MetricSeries        `json:"series"`
	Samples []MetricSeriesPoint `json:"samples"`
}

type ClusterMetricsResponse struct {
	ClusterID string                    `json:"clusterId"`
	From      time.Time                 `json:"from"`
	To        time.Time                 `json:"to"`
	Items     []MetricSeriesWithSamples `json:"items"`
}
