package model

import "time"

type UnderProvisionThresholds struct {
	CPUPressureThreshold             float64 `json:"cpuPressureThreshold"`
	MemoryPressureThreshold          float64 `json:"memoryPressureThreshold"`
	CPUPressureFrequencyThreshold    float64 `json:"cpuPressureFrequencyThreshold"`
	MemoryPressureFrequencyThreshold float64 `json:"memoryPressureFrequencyThreshold"`
}

type UnderProvisionFinding struct {
	Namespace      string `json:"namespace"`
	ControllerUID  string `json:"controllerUid"`
	ControllerKind string `json:"controllerKind"`
	ControllerName string `json:"controllerName"`

	WindowFrom time.Time `json:"windowFrom"`
	WindowTo   time.Time `json:"windowTo"`

	CurrentCPULimitCores    float64 `json:"currentCpuLimitCores"`
	CurrentMemoryLimitBytes int64   `json:"currentMemoryLimitBytes"`

	CPUPressureFrequency    float64 `json:"cpuPressureFrequency"`
	MemoryPressureFrequency float64 `json:"memoryPressureFrequency"`

	CPUUnderProvisioned    bool `json:"cpuUnderProvisioned"`
	MemoryUnderProvisioned bool `json:"memoryUnderProvisioned"`

	MemoryOOMDetected bool  `json:"memoryOomDetected"`
	OOMContainerCount int   `json:"oomContainerCount"`
	RestartCount      int32 `json:"restartCount"`

	Reason string `json:"reason,omitempty"`
}

type UnderProvisionResponse struct {
	ClusterID  string                   `json:"clusterId"`
	From       time.Time                `json:"from"`
	To         time.Time                `json:"to"`
	Thresholds UnderProvisionThresholds `json:"thresholds"`
	Items      []UnderProvisionFinding  `json:"items"`
}
