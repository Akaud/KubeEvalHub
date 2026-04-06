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
	GetClusterMetricSamples(
		ctx context.Context,
		ownerID int64,
		agentID string,
		from time.Time,
		to time.Time,
	) ([]model.MetricSampleRow, error)
	FindSeriesByIdentity(
		ctx context.Context,
		ownerID int64,
		agentID string,
		metricName string,
		resourceKind string,
		nodeName *string,
		namespace *string,
		podName *string,
		containerName *string,
	) (*model.MetricSeries, error)
	GetSeriesByIDForOwner(
		ctx context.Context,
		ownerID int64,
		agentID string,
		seriesID string,
	) (*model.MetricSeries, error)
	GetSeriesSamples(
		ctx context.Context,
		seriesID string,
		limit int,
	) ([]model.MetricSeriesPoint, error)
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

func (r *metricRepository) GetClusterMetricSamples(
	ctx context.Context,
	ownerID int64,
	agentID string,
	from time.Time,
	to time.Time,
) ([]model.MetricSampleRow, error) {
	query := `
		SELECT
			msr.id,
			msr.agent_id,
			msr.metric_name,
			msr.metric_type,
			msr.unit,
			msr.resource_kind,
			msr.node_name,
			msr.namespace,
			msr.pod_name,
			msr.container_name,
			msr.labels_hash,
			mss.collected_at,
			mss.value_double
		FROM agent_clusters ac
		JOIN agents a
			ON a.id = ac.agent_id
		JOIN metric_series msr
			ON msr.agent_id = a.id
		JOIN metric_samples mss
			ON mss.series_id = msr.id
		WHERE ac.agent_id = $1
		  AND a.owner_id = $2
		  AND mss.collected_at >= $3
		  AND mss.collected_at <= $4
		ORDER BY
			msr.metric_name,
			msr.resource_kind,
			msr.node_name NULLS FIRST,
			msr.namespace NULLS FIRST,
			msr.pod_name NULLS FIRST,
			msr.container_name NULLS FIRST,
			mss.collected_at ASC
	`

	rows, err := r.pool.Query(ctx, query, agentID, ownerID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]model.MetricSampleRow, 0)
	for rows.Next() {
		var row model.MetricSampleRow
		if err := rows.Scan(
			&row.SeriesID,
			&row.AgentID,
			&row.MetricName,
			&row.MetricType,
			&row.Unit,
			&row.ResourceKind,
			&row.NodeName,
			&row.Namespace,
			&row.PodName,
			&row.ContainerName,
			&row.LabelsHash,
			&row.CollectedAt,
			&row.Value,
		); err != nil {
			return nil, err
		}
		out = append(out, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *metricRepository) FindSeriesByIdentity(
	ctx context.Context,
	ownerID int64,
	agentID string,
	metricName string,
	resourceKind string,
	nodeName *string,
	namespace *string,
	podName *string,
	containerName *string,
) (*model.MetricSeries, error) {
	query := `
		SELECT
			msr.id,
			msr.agent_id,
			msr.metric_name,
			msr.metric_type,
			msr.unit,
			msr.resource_kind,
			msr.node_name,
			msr.namespace,
			msr.pod_name,
			msr.container_name,
			msr.labels_hash,
			msr.created_at,
			msr.updated_at
		FROM metric_series msr
		JOIN agents a
			ON a.id = msr.agent_id
		WHERE msr.agent_id = $1
		  AND a.owner_id = $2
		  AND msr.metric_name = $3
		  AND msr.resource_kind = $4
		  AND msr.node_name IS NOT DISTINCT FROM $5
		  AND msr.namespace IS NOT DISTINCT FROM $6
		  AND msr.pod_name IS NOT DISTINCT FROM $7
		  AND msr.container_name IS NOT DISTINCT FROM $8
		ORDER BY msr.updated_at DESC
		LIMIT 1
	`

	var series model.MetricSeries
	err := r.pool.QueryRow(
		ctx,
		query,
		agentID,
		ownerID,
		metricName,
		resourceKind,
		nodeName,
		namespace,
		podName,
		containerName,
	).Scan(
		&series.ID,
		&series.AgentID,
		&series.MetricName,
		&series.MetricType,
		&series.Unit,
		&series.ResourceKind,
		&series.NodeName,
		&series.Namespace,
		&series.PodName,
		&series.ContainerName,
		&series.LabelsHash,
		&series.CreatedAt,
		&series.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMetricSeriesNotFound
		}
		return nil, err
	}

	return &series, nil
}

func (r *metricRepository) GetSeriesByIDForOwner(
	ctx context.Context,
	ownerID int64,
	agentID string,
	seriesID string,
) (*model.MetricSeries, error) {
	query := `
		SELECT
			msr.id,
			msr.agent_id,
			msr.metric_name,
			msr.metric_type,
			msr.unit,
			msr.resource_kind,
			msr.node_name,
			msr.namespace,
			msr.pod_name,
			msr.container_name,
			msr.labels_hash,
			msr.created_at,
			msr.updated_at
		FROM metric_series msr
		JOIN agents a
			ON a.id = msr.agent_id
		WHERE msr.id = $1
		  AND msr.agent_id = $2
		  AND a.owner_id = $3
		LIMIT 1
	`

	var series model.MetricSeries
	err := r.pool.QueryRow(ctx, query, seriesID, agentID, ownerID).Scan(
		&series.ID,
		&series.AgentID,
		&series.MetricName,
		&series.MetricType,
		&series.Unit,
		&series.ResourceKind,
		&series.NodeName,
		&series.Namespace,
		&series.PodName,
		&series.ContainerName,
		&series.LabelsHash,
		&series.CreatedAt,
		&series.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMetricSeriesNotFound
		}
		return nil, err
	}

	return &series, nil
}

func (r *metricRepository) GetSeriesSamples(
	ctx context.Context,
	seriesID string,
	limit int,
) ([]model.MetricSeriesPoint, error) {
	if limit <= 0 {
		limit = 120
	}

	query := `
		SELECT collected_at, value_double
		FROM metric_samples
		WHERE series_id = $1
		ORDER BY collected_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, seriesID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]model.MetricSeriesPoint, 0, limit)
	for rows.Next() {
		var point model.MetricSeriesPoint
		if err := rows.Scan(&point.CollectedAt, &point.Value); err != nil {
			return nil, err
		}
		points = append(points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	return points, nil
}
