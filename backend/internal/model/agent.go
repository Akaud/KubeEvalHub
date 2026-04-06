package model

import "time"

type AgentStatus string

const (
	AgentStatusDisabled       AgentStatus = "disabled"
	AgentStatusOnline         AgentStatus = "online"
	AgentStatusOffline        AgentStatus = "offline"
	AgentStatusNeverConnected AgentStatus = "never_connected"
)

type Agent struct {
	ID              string     `json:"id"`
	OwnerID         int64      `json:"ownerId"`
	Name            string     `json:"name"`
	Enabled         bool       `json:"enabled"`
	Active          bool       `json:"active"`
	TokenHash       string     `json:"-"`
	LastHeartbeatAt *time.Time `json:"lastHeartbeatAt,omitempty"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
}

type AgentView struct {
	ID              string      `json:"id"`
	OwnerID         int64       `json:"ownerId"`
	Name            string      `json:"name"`
	Enabled         bool        `json:"enabled"`
	Active          bool        `json:"active"`
	Status          AgentStatus `json:"status"`
	LastHeartbeatAt *time.Time  `json:"lastHeartbeatAt,omitempty"`
	CreatedAt       time.Time   `json:"createdAt"`
	UpdatedAt       time.Time   `json:"updatedAt"`
}
