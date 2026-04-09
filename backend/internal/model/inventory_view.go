package model

import "time"

type InventorySnapshotResponse struct {
	ClusterID   string           `json:"clusterId"`
	CollectedAt time.Time        `json:"collectedAt"`
	Inventory   InventoryPayload `json:"inventory"`
}
