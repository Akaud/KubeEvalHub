package model

import "time"

type OverProvisionThresholds struct {
	CPUSafetyMargin          float64 `json:"cpuSafetyMargin"`
	MemorySafetyMargin       float64 `json:"memorySafetyMargin"`
	CPUOverprovisionRatio    float64 `json:"cpuOverprovisionRatio"`
	MemoryOverprovisionRatio float64 `json:"memoryOverprovisionRatio"`
	MinReclaimCPUCores       float64 `json:"minReclaimCpuCores"`
	MinReclaimMemoryBytes    int64   `json:"minReclaimMemoryBytes"`
}

type OverProvisionFinding struct {
	Namespace      string `json:"namespace"`
	ControllerUID  string `json:"controllerUid"`
	ControllerKind string `json:"controllerKind"`
	ControllerName string `json:"controllerName"`

	WindowFrom time.Time `json:"windowFrom"`
	WindowTo   time.Time `json:"windowTo"`

	CurrentCPURequestCores     float64 `json:"currentCpuRequestCores"`
	ObservedCPUP95Cores        float64 `json:"observedCpuP95Cores"`
	RecommendedCPURequestCores float64 `json:"recommendedCpuRequestCores"`
	ReclaimableCPUCores        float64 `json:"reclaimableCpuCores"`

	CurrentMemoryRequestBytes     int64 `json:"currentMemoryRequestBytes"`
	ObservedMemoryP95Bytes        int64 `json:"observedMemoryP95Bytes"`
	RecommendedMemoryRequestBytes int64 `json:"recommendedMemoryRequestBytes"`
	ReclaimableMemoryBytes        int64 `json:"reclaimableMemoryBytes"`

	CPUOverProvisioned    bool `json:"cpuOverProvisioned"`
	MemoryOverProvisioned bool `json:"memoryOverProvisioned"`
}

type OverProvisionResponse struct {
	ClusterID  string                  `json:"clusterId"`
	From       time.Time               `json:"from"`
	To         time.Time               `json:"to"`
	Thresholds OverProvisionThresholds `json:"thresholds"`
	Items      []OverProvisionFinding  `json:"items"`
}
