package model

import "time"

type AgentCluster struct {
	AgentID       string    `json:"agentId"`
	ClusterUID    string    `json:"clusterUid"`
	ClusterName   string    `json:"clusterName"`
	KubeVersion   string    `json:"kubeVersion"`
	Distribution  string    `json:"distribution"`
	APIServerHost string    `json:"apiServerHost"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}
