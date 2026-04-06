package model

import "time"

type MetricSample struct {
	ID          string    `json:"id"`
	SeriesID    string    `json:"seriesId"`
	CollectedAt time.Time `json:"collectedAt"`
	ReceivedAt  time.Time `json:"receivedAt"`
	Value       float64   `json:"value"`
}
