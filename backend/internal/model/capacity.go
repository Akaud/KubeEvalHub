package model

import "time"

type NamespaceCapacity struct {
	Namespace string `json:"namespace"`

	TotalCPURequestCores    float64 `json:"totalCpuRequestCores"`
	TotalMemoryRequestBytes int64   `json:"totalMemoryRequestBytes"`

	ReclaimableCPUCores    float64 `json:"reclaimableCpuCores"`
	ReclaimableMemoryBytes int64   `json:"reclaimableMemoryBytes"`

	WorkloadCount         int `json:"workloadCount"`
	EligibleWorkloadCount int `json:"eligibleWorkloadCount"`
}

type ClusterCapacityResponse struct {
	ClusterID string    `json:"clusterId"`
	From      time.Time `json:"from"`
	To        time.Time `json:"to"`

	Namespaces []NamespaceCapacity `json:"namespaces"`

	TotalCPURequestCores    float64 `json:"totalCpuRequestCores"`
	TotalMemoryRequestBytes int64   `json:"totalMemoryRequestBytes"`

	ReclaimableCPUCores    float64 `json:"reclaimableCpuCores"`
	ReclaimableMemoryBytes int64   `json:"reclaimableMemoryBytes"`

	EligibleWorkloads int `json:"eligibleWorkloads"`
	TotalWorkloads    int `json:"totalWorkloads"`
}
