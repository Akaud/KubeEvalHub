package model

import "time"

type RecommendationThresholds struct {
	CPUSafetyMargin    float64 `json:"cpuSafetyMargin"`
	MemorySafetyMargin float64 `json:"memorySafetyMargin"`

	MinCPURequestCores    float64 `json:"minCpuRequestCores"`
	MinMemoryRequestBytes int64   `json:"minMemoryRequestBytes"`

	MinSampleCount int `json:"minSampleCount"`
}

type RightSizingRecommendation struct {
	Namespace      string `json:"namespace"`
	ControllerUID  string `json:"controllerUid"`
	ControllerKind string `json:"controllerKind"`
	ControllerName string `json:"controllerName"`

	WindowFrom time.Time `json:"windowFrom"`
	WindowTo   time.Time `json:"windowTo"`

	Eligible bool   `json:"eligible"`
	Reason   string `json:"reason,omitempty"`

	CurrentCPURequestCores     float64 `json:"currentCpuRequestCores"`
	ObservedCPUP95Cores        float64 `json:"observedCpuP95Cores"`
	RecommendedCPURequestCores float64 `json:"recommendedCpuRequestCores"`

	CurrentMemoryRequestBytes     int64 `json:"currentMemoryRequestBytes"`
	ObservedMemoryP95Bytes        int64 `json:"observedMemoryP95Bytes"`
	RecommendedMemoryRequestBytes int64 `json:"recommendedMemoryRequestBytes"`

	CPUSafetyMarginUsed    float64 `json:"cpuSafetyMarginUsed"`
	MemorySafetyMarginUsed float64 `json:"memorySafetyMarginUsed"`

	CPUUnderProvisioned    bool `json:"cpuUnderProvisioned"`
	MemoryUnderProvisioned bool `json:"memoryUnderProvisioned"`
	MemoryOOMDetected      bool `json:"memoryOomDetected"`
}

type RecommendationResponse struct {
	ClusterID  string                      `json:"clusterId"`
	From       time.Time                   `json:"from"`
	To         time.Time                   `json:"to"`
	Thresholds RecommendationThresholds    `json:"thresholds"`
	Items      []RightSizingRecommendation `json:"items"`
}
