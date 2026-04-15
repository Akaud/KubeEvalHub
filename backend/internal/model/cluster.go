package model

import "time"

type AgentCluster struct {
	ID            string    `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	AgentID       *string   `json:"agentId,omitempty"`
	ClusterUID    string    `json:"clusterUid"`
	ClusterName   string    `json:"clusterName"`
	KubeVersion   string    `json:"kubeVersion"`
	Distribution  string    `json:"distribution"`
	APIServerHost string    `json:"apiServerHost"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

type ClusterView struct {
	ID              string      `json:"id"`
	OwnerID         int64       `json:"ownerId"`
	AgentID         *string     `json:"agentId,omitempty"`
	ClusterUID      string      `json:"clusterUid"`
	ClusterName     string      `json:"clusterName"`
	KubeVersion     string      `json:"kubeVersion"`
	Distribution    string      `json:"distribution"`
	APIServerHost   string      `json:"apiServerHost"`
	LastHeartbeatAt *time.Time  `json:"lastHeartbeatAt,omitempty"`
	Status          AgentStatus `json:"status"`
	MyRole          string      `json:"myRole"`
}

type CreateClusterRequest struct {
	ClusterName string `json:"clusterName"`
}

type AssignAgentRequest struct {
	AgentID string `json:"agentId"`
}

type ClusterPayload struct {
	ClusterUID    string `json:"clusterUid"`
	ClusterName   string `json:"clusterName"`
	KubeVersion   string `json:"kubeVersion"`
	Distribution  string `json:"distribution"`
	APIServerHost string `json:"apiServerHost"`
}

type UpsertClusterUserRoleRequest struct {
	Email string      `json:"email"`
	Role  ClusterRole `json:"role"`
}
