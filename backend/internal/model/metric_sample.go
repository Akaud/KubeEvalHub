package model

import "time"

type MetricSample struct {
	ID          string    `json:"id"`
	SeriesID    string    `json:"seriesId"`
	CollectedAt time.Time `json:"collectedAt"`
	ReceivedAt  time.Time `json:"receivedAt"`
	Value       float64   `json:"value"`
}

type MetricSampleRow struct {
	SeriesID  string
	ClusterID string

	MetricName   string
	MetricType   string
	Unit         string
	ResourceKind string

	NodeName      *string
	Namespace     *string
	PodName       *string
	PodUID        *string
	ContainerName *string

	ControllerUID  *string
	ControllerKind *string
	ControllerName *string

	CollectedAt time.Time
	Value       float64
}
