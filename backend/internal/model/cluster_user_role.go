package model

import "time"

type ClusterUserRole struct {
	ClusterID string      `json:"clusterId"`
	UserID    int64       `json:"userId"`
	Role      ClusterRole `json:"role"`
	CreatedAt time.Time   `json:"createdAt"`
	UpdatedAt time.Time   `json:"updatedAt"`
}
