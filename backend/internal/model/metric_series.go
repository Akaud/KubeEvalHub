package model

import "time"

type MetricSeries struct {
	ID           string `json:"id"`
	ClusterID    string `json:"clusterId"`
	MetricName   string `json:"metricName"`
	MetricType   string `json:"metricType"`
	Unit         string `json:"unit"`
	ResourceKind string `json:"resourceKind"`

	NodeName      *string `json:"nodeName,omitempty"`
	Namespace     *string `json:"namespace,omitempty"`
	PodName       *string `json:"podName,omitempty"`
	PodUID        *string `json:"podUid,omitempty"`
	ContainerName *string `json:"containerName,omitempty"`

	ControllerUID  *string `json:"controllerUid,omitempty"`
	ControllerKind *string `json:"controllerKind,omitempty"`
	ControllerName *string `json:"controllerName,omitempty"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
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
