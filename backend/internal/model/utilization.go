package model

import "time"

type UtilizationQuery struct {
	From time.Time `json:"from"`
	To   time.Time `json:"to"`
}

type ResourceStats struct {
	Average     float64 `json:"average"`
	Maximum     float64 `json:"maximum"`
	P50         float64 `json:"p50"`
	P95         float64 `json:"p95"`
	P99         float64 `json:"p99"`
	SampleCount int     `json:"sampleCount"`
	Unit        string  `json:"unit"`
}

type WorkloadUtilizationSummary struct {
	Namespace      string `json:"namespace"`
	ControllerUID  string `json:"controllerUid"`
	ControllerKind string `json:"controllerKind"`
	ControllerName string `json:"controllerName"`

	WindowFrom time.Time `json:"windowFrom"`
	WindowTo   time.Time `json:"windowTo"`

	CPU    *ResourceStats `json:"cpu,omitempty"`
	Memory *ResourceStats `json:"memory,omitempty"`
}

type ClusterUtilizationResponse struct {
	ClusterID string                       `json:"clusterId"`
	From      time.Time                    `json:"from"`
	To        time.Time                    `json:"to"`
	Items     []WorkloadUtilizationSummary `json:"items"`
}
