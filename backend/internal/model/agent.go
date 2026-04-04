package model

import "time"

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
