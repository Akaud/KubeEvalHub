package repository

import (
	"context"
	"errors"
	"time"

	"backend/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrMetricSeriesNotFound = errors.New("metric series not found")

type MetricRepository interface {
	UpsertSeries(ctx context.Context, series *model.MetricSeries) (string, error)
	InsertSample(ctx context.Context, sample *model.MetricSample) error
	InsertSamples(ctx context.Context, samples []model.MetricSample) error
}

type metricRepository struct {
	pool *pgxpool.Pool
}

func NewMetricRepository(pool *pgxpool.Pool) MetricRepository {
	return &metricRepository{pool: pool}
}

func (r *metricRepository) UpsertSeries(ctx context.Context, series *model.MetricSeries) (string, error) {
	query := `
		INSERT INTO metric_series (
			id,
			agent_id,
			metric_name,
			metric_type,
			unit,
			resource_kind,
			node_name,
			namespace,
			pod_name,
			container_name,
			labels_hash,
			created_at,
			updated_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (
			agent_id,
			metric_name,
			resource_kind,
			node_name,
			namespace,
			pod_name,
			container_name,
			labels_hash
		)
		DO UPDATE SET
			metric_type = EXCLUDED.metric_type,
			unit = EXCLUDED.unit,
			updated_at = EXCLUDED.updated_at
		RETURNING id
	`

	var id string
	err := r.pool.QueryRow(
		ctx,
		query,
		series.ID,
		series.AgentID,
		series.MetricName,
		series.MetricType,
		series.Unit,
		series.ResourceKind,
		series.NodeName,
		series.Namespace,
		series.PodName,
		series.ContainerName,
		series.LabelsHash,
		series.CreatedAt,
		series.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *metricRepository) InsertSample(ctx context.Context, sample *model.MetricSample) error {
	query := `
		INSERT INTO metric_samples (
			id,
			series_id,
			collected_at,
			received_at,
			value_double
		)
		VALUES ($1,$2,$3,$4,$5)
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		sample.ID,
		sample.SeriesID,
		sample.CollectedAt,
		sample.ReceivedAt,
		sample.Value,
	)

	return err
}

func (r *metricRepository) InsertSamples(ctx context.Context, samples []model.MetricSample) error {
	if len(samples) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO metric_samples (
			id,
			series_id,
			collected_at,
			received_at,
			value_double
		)
		VALUES ($1,$2,$3,$4,$5)
	`

	for _, sample := range samples {
		s := sample
		batch.Queue(query,
			s.ID,
			s.SeriesID,
			s.CollectedAt,
			s.ReceivedAt,
			s.Value,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range samples {
		_, err := br.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func nowUTCMetric() time.Time {
	return time.Now().UTC()
}
