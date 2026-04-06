package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

var (
	ErrInvalidMetricsPayload = errors.New("invalid metrics payload")
	ErrClusterUIDMismatch    = errors.New("cluster uid mismatch")
)

type MetricService interface {
	IngestMetrics(ctx context.Context, agentID string, req *model.PushMetricsRequest) error
}

type metricService struct {
	clusterRepo repository.ClusterRepository
	metricRepo  repository.MetricRepository
}

func NewMetricService(
	clusterRepo repository.ClusterRepository,
	metricRepo repository.MetricRepository,
) MetricService {
	return &metricService{
		clusterRepo: clusterRepo,
		metricRepo:  metricRepo,
	}
}

func (s *metricService) IngestMetrics(ctx context.Context, agentID string, req *model.PushMetricsRequest) error {
	if req == nil {
		return ErrInvalidMetricsPayload
	}
	if strings.TrimSpace(agentID) == "" {
		return ErrInvalidMetricsPayload
	}
	if strings.TrimSpace(req.Cluster.ClusterUID) == "" ||
		strings.TrimSpace(req.Cluster.ClusterName) == "" ||
		strings.TrimSpace(req.Cluster.KubeVersion) == "" {
		return ErrInvalidMetricsPayload
	}
	if len(req.Samples) == 0 {
		return ErrInvalidMetricsPayload
	}

	now := time.Now().UTC()

	existingCluster, err := s.clusterRepo.GetByAgentID(ctx, agentID)
	if err != nil && !errors.Is(err, repository.ErrClusterNotFound) {
		return err
	}
	if existingCluster != nil && existingCluster.ClusterUID != req.Cluster.ClusterUID {
		return ErrClusterUIDMismatch
	}

	distribution := strings.TrimSpace(req.Cluster.Distribution)
	if distribution == "" {
		distribution = "unknown"
	}

	cluster := &model.AgentCluster{
		AgentID:       agentID,
		ClusterUID:    strings.TrimSpace(req.Cluster.ClusterUID),
		ClusterName:   strings.TrimSpace(req.Cluster.ClusterName),
		KubeVersion:   strings.TrimSpace(req.Cluster.KubeVersion),
		Distribution:  distribution,
		APIServerHost: strings.TrimSpace(req.Cluster.APIServerHost),
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if existingCluster != nil {
		cluster.CreatedAt = existingCluster.CreatedAt
	}

	if err := s.clusterRepo.Upsert(ctx, cluster); err != nil {
		return err
	}

	samples := make([]model.MetricSample, 0, len(req.Samples))

	for _, point := range req.Samples {
		if err := validatePoint(point); err != nil {
			return err
		}

		labelsHash, err := hashLabels(point.Labels)
		if err != nil {
			return err
		}

		nodeName := stringPtrOrNil(point.NodeName)
		namespace := stringPtrOrNil(point.Namespace)
		podName := stringPtrOrNil(point.PodName)
		containerName := stringPtrOrNil(point.ContainerName)

		series := &model.MetricSeries{
			ID:            uuid.NewString(),
			AgentID:       agentID,
			MetricName:    strings.TrimSpace(point.MetricName),
			MetricType:    strings.TrimSpace(point.MetricType),
			Unit:          strings.TrimSpace(point.Unit),
			ResourceKind:  strings.TrimSpace(point.ResourceKind),
			NodeName:      nodeName,
			Namespace:     namespace,
			PodName:       podName,
			ContainerName: containerName,
			LabelsHash:    labelsHash,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		seriesID, err := s.metricRepo.UpsertSeries(ctx, series)
		if err != nil {
			return err
		}

		samples = append(samples, model.MetricSample{
			ID:          uuid.NewString(),
			SeriesID:    seriesID,
			CollectedAt: point.CollectedAt,
			ReceivedAt:  now,
			Value:       point.Value,
		})
	}

	return s.metricRepo.InsertSamples(ctx, samples)
}

func validatePoint(point model.MetricPointPayload) error {
	if strings.TrimSpace(point.MetricName) == "" {
		return ErrInvalidMetricsPayload
	}
	if strings.TrimSpace(point.MetricType) == "" {
		return ErrInvalidMetricsPayload
	}
	if strings.TrimSpace(point.Unit) == "" {
		return ErrInvalidMetricsPayload
	}
	if strings.TrimSpace(point.ResourceKind) == "" {
		return ErrInvalidMetricsPayload
	}
	if point.CollectedAt.IsZero() {
		return ErrInvalidMetricsPayload
	}
	if math.IsNaN(point.Value) || math.IsInf(point.Value, 0) {
		return ErrInvalidMetricsPayload
	}

	return nil
}

func stringPtrOrNil(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}

func hashLabels(labels map[string]string) (string, error) {
	if len(labels) == 0 {
		sum := sha256.Sum256([]byte("{}"))
		return hex.EncodeToString(sum[:]), nil
	}

	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	normalized := make(map[string]string, len(labels))
	for _, k := range keys {
		normalized[k] = labels[k]
	}

	b, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}
