package model

import "time"

type ForecastRequest struct {
	SeriesID     string `json:"seriesId"`
	Model        string `json:"model"`
	Steps        int    `json:"steps"`
	HistoryLimit int    `json:"historyLimit"`

	MetricName    string  `json:"metricName,omitempty"`
	ResourceKind  string  `json:"resourceKind,omitempty"`
	NodeName      *string `json:"nodeName,omitempty"`
	Namespace     *string `json:"namespace,omitempty"`
	PodName       *string `json:"podName,omitempty"`
	ContainerName *string `json:"containerName,omitempty"`
}

type ForecastPoint struct {
	CollectedAt time.Time `json:"collectedAt"`
	Value       float64   `json:"value"`
}

type ForecastModelInfo struct {
	Name           string `json:"name"`
	TrainingPoints int    `json:"trainingPoints"`
	StepSeconds    int64  `json:"stepSeconds"`
}

type ForecastResponse struct {
	Series   MetricSeries        `json:"series"`
	Actual   []MetricSeriesPoint `json:"actual"`
	Forecast []ForecastPoint     `json:"forecast"`
	Model    ForecastModelInfo   `json:"model"`
}
