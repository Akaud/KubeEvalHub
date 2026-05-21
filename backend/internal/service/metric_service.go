package service

import (
	"context"
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
	ErrInvalidMetricsPayload    = errors.New("invalid metrics payload")
	ErrClusterUIDMismatch       = errors.New("cluster uid mismatch")
	ErrInsufficientForecastData = errors.New("insufficient forecast data")
	ErrAgentClusterNotAssigned  = errors.New("agent is not assigned to a cluster")
)

type MetricService interface {
	IngestMetrics(ctx context.Context, agentID string, req *model.PushMetricsRequest) error
	GetClusterMetrics(
		ctx context.Context,
		clusterID string,
		from time.Time,
		to time.Time,
	) (*model.ClusterMetricsResponse, error)
	ForecastClusterMetric(
		ctx context.Context,
		clusterID string,
		req *model.ForecastRequest,
	) (*model.ForecastResponse, error)
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

	agentID = strings.TrimSpace(agentID)
	if agentID == "" {
		return ErrInvalidMetricsPayload
	}

	if strings.TrimSpace(req.Cluster.ClusterUID) == "" ||
		strings.TrimSpace(req.Cluster.KubeVersion) == "" {
		return ErrInvalidMetricsPayload
	}

	if len(req.Samples) == 0 {
		return ErrInvalidMetricsPayload
	}

	now := time.Now().UTC()

	cluster, err := s.clusterRepo.GetByAgentID(ctx, agentID)
	if err != nil {
		if errors.Is(err, repository.ErrClusterNotFound) {
			return ErrAgentClusterNotAssigned
		}
		return err
	}

	incomingClusterUID := strings.TrimSpace(req.Cluster.ClusterUID)

	if strings.TrimSpace(cluster.ClusterUID) != "" && cluster.ClusterUID != incomingClusterUID {
		return ErrClusterUIDMismatch
	}

	distribution := strings.TrimSpace(req.Cluster.Distribution)
	if distribution == "" {
		distribution = "unknown"
	}

	if err := s.clusterRepo.UpdateMetadataByAgentID(
		ctx,
		agentID,
		incomingClusterUID,
		cluster.ClusterName,
		strings.TrimSpace(req.Cluster.KubeVersion),
		distribution,
		strings.TrimSpace(req.Cluster.APIServerHost),
		now,
	); err != nil {
		return err
	}

	samples := make([]model.MetricSample, 0, len(req.Samples))

	for _, point := range req.Samples {
		if err := validatePoint(point); err != nil {
			return err
		}

		series := &model.MetricSeries{
			ID:             uuid.NewString(),
			ClusterID:      cluster.ID,
			MetricName:     strings.TrimSpace(point.MetricName),
			MetricType:     strings.TrimSpace(point.MetricType),
			Unit:           strings.TrimSpace(point.Unit),
			ResourceKind:   strings.TrimSpace(point.ResourceKind),
			NodeName:       stringPtrOrNil(point.NodeName),
			Namespace:      stringPtrOrNil(point.Namespace),
			PodName:        stringPtrOrNil(point.PodName),
			PodUID:         stringPtrOrNil(point.PodUID),
			ContainerName:  stringPtrOrNil(point.ContainerName),
			ControllerUID:  stringPtrOrNil(point.ControllerUID),
			ControllerKind: stringPtrOrNil(point.ControllerKind),
			ControllerName: stringPtrOrNil(point.ControllerName),
			CreatedAt:      now,
			UpdatedAt:      now,
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

func (s *metricService) GetClusterMetrics(
	ctx context.Context,
	clusterID string,
	from time.Time,
	to time.Time,
) (*model.ClusterMetricsResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, ErrInvalidMetricsPayload
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, ErrInvalidMetricsPayload
	}

	rows, err := s.metricRepo.GetClusterMetricSamples(ctx, clusterID, from, to)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string]*model.MetricSeriesWithSamples)
	order := make([]string, 0)

	for _, row := range rows {
		item, exists := grouped[row.SeriesID]
		if !exists {
			item = &model.MetricSeriesWithSamples{
				Series: model.MetricSeries{
					ID:             row.SeriesID,
					ClusterID:      row.ClusterID,
					MetricName:     row.MetricName,
					MetricType:     row.MetricType,
					Unit:           row.Unit,
					ResourceKind:   row.ResourceKind,
					NodeName:       row.NodeName,
					Namespace:      row.Namespace,
					PodName:        row.PodName,
					PodUID:         row.PodUID,
					ContainerName:  row.ContainerName,
					ControllerUID:  row.ControllerUID,
					ControllerKind: row.ControllerKind,
					ControllerName: row.ControllerName,
				},
				Samples: make([]model.MetricSeriesPoint, 0),
			}
			grouped[row.SeriesID] = item
			order = append(order, row.SeriesID)
		}

		item.Samples = append(item.Samples, model.MetricSeriesPoint{
			CollectedAt: row.CollectedAt,
			Value:       row.Value,
		})
	}

	items := make([]model.MetricSeriesWithSamples, 0, len(order))
	for _, id := range order {
		items = append(items, *grouped[id])
	}

	return &model.ClusterMetricsResponse{
		ClusterID: clusterID,
		From:      from,
		To:        to,
		Items:     items,
	}, nil
}

func (s *metricService) ForecastClusterMetric(
	ctx context.Context,
	clusterID string,
	req *model.ForecastRequest,
) (*model.ForecastResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" || req == nil {
		return nil, ErrInvalidMetricsPayload
	}

	if req.Steps <= 0 {
		req.Steps = 8
	}
	if req.HistoryLimit <= 0 {
		req.HistoryLimit = 120
	}
	if req.Model == "" {
		req.Model = "auto"
	}

	req.SeriesID = strings.TrimSpace(req.SeriesID)
	req.Model = strings.TrimSpace(strings.ToLower(req.Model))
	req.MetricName = strings.TrimSpace(req.MetricName)
	req.ResourceKind = strings.TrimSpace(req.ResourceKind)

	var (
		series *model.MetricSeries
		err    error
	)

	if req.SeriesID != "" {
		series, err = s.metricRepo.GetSeriesByIDForCluster(ctx, clusterID, req.SeriesID)
	} else {
		if req.MetricName == "" || req.ResourceKind == "" {
			return nil, ErrInvalidMetricsPayload
		}

		series, err = s.metricRepo.FindSeriesByIdentity(
			ctx,
			clusterID,
			req.MetricName,
			req.ResourceKind,
			trimPtr(req.NodeName),
			trimPtr(req.Namespace),
			trimPtr(req.PodName),
			trimPtr(req.PodUID),
			trimPtr(req.ContainerName),
			trimPtr(req.ControllerUID),
			trimPtr(req.ControllerKind),
			trimPtr(req.ControllerName),
		)
	}
	if err != nil {
		return nil, err
	}

	history, err := s.metricRepo.GetSeriesSamples(ctx, series.ID, req.HistoryLimit)
	if err != nil {
		return nil, err
	}

	history = normalizeHistory(history)
	if len(history) == 0 {
		return nil, ErrInsufficientForecastData
	}
	if len(history) > 20 {
		history = winsorizeHistory(history, 0.05, 0.95)
	}

	var (
		forecast    []model.ForecastPoint
		stepSeconds int64
		modelName   string
	)

	switch req.Model {
	case "auto":
		forecast, stepSeconds, modelName, err = forecastAuto(history, req.Steps, series.MetricName)
	case "last_value":
		forecast, stepSeconds, err = forecastLastValue(history, req.Steps)
		modelName = "last_value"
	case "damped_holt":
		runner := bestDampedHoltRunner()
		forecast, stepSeconds, err = runner(history, req.Steps)
		modelName = "damped_holt"
	case "holt_linear":
		runner := bestHoltLinearRunner()
		forecast, stepSeconds, err = runner(history, req.Steps)
		modelName = "holt_linear"
	case "weighted_moving_average":
		forecast, stepSeconds, err = forecastWeightedMovingAverage(history, req.Steps, 10)
		modelName = "weighted_moving_average"
	case "moving_average":
		forecast, stepSeconds, err = forecastMovingAverage(history, req.Steps, 10)
		modelName = "moving_average"
	default:
		forecast, stepSeconds, modelName, err = forecastAuto(history, req.Steps, series.MetricName)
	}
	if err != nil {
		return nil, err
	}

	return &model.ForecastResponse{
		Series:   *series,
		Actual:   history,
		Forecast: forecast,
		Model: model.ForecastModelInfo{
			Name:           modelName,
			TrainingPoints: len(history),
			StepSeconds:    stepSeconds,
		},
	}, nil
}

func defaultForecastWindow(historyLen int) int {
	if historyLen <= 0 {
		return 1
	}

	window := historyLen / 4

	if window < 5 {
		window = 5
	}

	if window > 20 {
		window = 20
	}

	if window > historyLen {
		window = historyLen
	}

	return window
}

func trimPtr(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}

func normalizeHistory(history []model.MetricSeriesPoint) []model.MetricSeriesPoint {
	if len(history) == 0 {
		return nil
	}

	points := make([]model.MetricSeriesPoint, 0, len(history))
	for _, p := range history {
		if p.CollectedAt.IsZero() {
			continue
		}
		if math.IsNaN(p.Value) || math.IsInf(p.Value, 0) {
			continue
		}
		points = append(points, p)
	}

	if len(points) == 0 {
		return nil
	}

	sort.Slice(points, func(i, j int) bool {
		return points[i].CollectedAt.Before(points[j].CollectedAt)
	})

	merged := make([]model.MetricSeriesPoint, 0, len(points))
	cur := points[0]
	count := 1

	for i := 1; i < len(points); i++ {
		if points[i].CollectedAt.Equal(cur.CollectedAt) {
			cur.Value += points[i].Value
			count++
			continue
		}

		cur.Value /= float64(count)
		merged = append(merged, cur)

		cur = points[i]
		count = 1
	}

	cur.Value /= float64(count)
	merged = append(merged, cur)

	return merged
}

func medianStepSeconds(history []model.MetricSeriesPoint) int64 {
	if len(history) < 2 {
		return 60
	}

	steps := make([]int64, 0, len(history)-1)
	for i := 1; i < len(history); i++ {
		sec := int64(history[i].CollectedAt.Sub(history[i-1].CollectedAt).Seconds())
		if sec > 0 {
			steps = append(steps, sec)
		}
	}

	if len(steps) == 0 {
		return 60
	}

	sort.Slice(steps, func(i, j int) bool { return steps[i] < steps[j] })

	mid := len(steps) / 2
	if len(steps)%2 == 1 {
		return steps[mid]
	}

	v := (steps[mid-1] + steps[mid]) / 2
	if v <= 0 {
		return 60
	}

	return v
}

func clampNonNegative(v float64) float64 {
	if v < 0 {
		return 0
	}
	return v
}

func winsorizeHistory(history []model.MetricSeriesPoint, lowerQ, upperQ float64) []model.MetricSeriesPoint {
	if len(history) < 5 {
		out := make([]model.MetricSeriesPoint, len(history))
		copy(out, history)
		return out
	}

	values := make([]float64, 0, len(history))
	for _, p := range history {
		values = append(values, p.Value)
	}
	sort.Float64s(values)

	lo := quantileSorted(values, lowerQ)
	hi := quantileSorted(values, upperQ)

	out := make([]model.MetricSeriesPoint, len(history))
	for i, p := range history {
		v := p.Value
		if v < lo {
			v = lo
		}
		if v > hi {
			v = hi
		}
		out[i] = model.MetricSeriesPoint{
			CollectedAt: p.CollectedAt,
			Value:       v,
		}
	}

	return out
}

func forecastLastValue(
	history []model.MetricSeriesPoint,
	steps int,
) ([]model.ForecastPoint, int64, error) {
	if len(history) < 1 {
		return nil, 0, ErrInsufficientForecastData
	}
	if steps <= 0 {
		steps = 8
	}

	stepSeconds := medianStepSeconds(history)
	last := history[len(history)-1]

	out := make([]model.ForecastPoint, 0, steps)
	for i := 1; i <= steps; i++ {
		out = append(out, model.ForecastPoint{
			CollectedAt: last.CollectedAt.Add(time.Duration(int64(i)*stepSeconds) * time.Second),
			Value:       clampNonNegative(last.Value),
		})
	}

	return out, stepSeconds, nil
}

func forecastMovingAverage(
	history []model.MetricSeriesPoint,
	steps int,
	window int,
) ([]model.ForecastPoint, int64, error) {
	if len(history) < 3 {
		return nil, 0, ErrInsufficientForecastData
	}
	if steps <= 0 {
		steps = 8
	}
	if window <= 0 {
		window = defaultForecastWindow(len(history))
	}
	if window > len(history) {
		window = len(history)
	}

	stepSeconds := medianStepSeconds(history)
	lastTime := history[len(history)-1].CollectedAt

	sum := 0.0
	for _, p := range history[len(history)-window:] {
		sum += p.Value
	}
	avg := sum / float64(window)

	out := make([]model.ForecastPoint, 0, steps)
	for i := 1; i <= steps; i++ {
		out = append(out, model.ForecastPoint{
			CollectedAt: lastTime.Add(time.Duration(int64(i)*stepSeconds) * time.Second),
			Value:       clampNonNegative(avg),
		})
	}

	return out, stepSeconds, nil
}

func forecastWeightedMovingAverage(
	history []model.MetricSeriesPoint,
	steps int,
	window int,
) ([]model.ForecastPoint, int64, error) {
	if len(history) < 3 {
		return nil, 0, ErrInsufficientForecastData
	}
	if steps <= 0 {
		steps = 8
	}
	if window <= 0 {
		window = defaultForecastWindow(len(history))
	}
	if window > len(history) {
		window = len(history)
	}

	stepSeconds := medianStepSeconds(history)
	lastTime := history[len(history)-1].CollectedAt
	slice := history[len(history)-window:]

	weightedSum := 0.0
	weightTotal := 0.0
	for i, p := range slice {
		w := float64(i + 1)
		weightedSum += w * p.Value
		weightTotal += w
	}
	avg := weightedSum / weightTotal

	out := make([]model.ForecastPoint, 0, steps)
	for i := 1; i <= steps; i++ {
		out = append(out, model.ForecastPoint{
			CollectedAt: lastTime.Add(time.Duration(int64(i)*stepSeconds) * time.Second),
			Value:       clampNonNegative(avg),
		})
	}

	return out, stepSeconds, nil
}

func forecastHoltLinear(
	history []model.MetricSeriesPoint,
	steps int,
	alpha float64,
	beta float64,
) ([]model.ForecastPoint, int64, error) {
	if len(history) < 3 {
		return nil, 0, ErrInsufficientForecastData
	}
	if steps <= 0 {
		steps = 8
	}
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.6
	}
	if beta <= 0 || beta >= 1 {
		beta = 0.3
	}

	stepSeconds := medianStepSeconds(history)

	level := history[0].Value
	trend := history[1].Value - history[0].Value

	for i := 1; i < len(history); i++ {
		value := history[i].Value
		prevLevel := level

		level = alpha*value + (1-alpha)*(level+trend)
		trend = beta*(level-prevLevel) + (1-beta)*trend
	}

	lastTime := history[len(history)-1].CollectedAt
	out := make([]model.ForecastPoint, 0, steps)

	for m := 1; m <= steps; m++ {
		predictedValue := level + float64(m)*trend
		out = append(out, model.ForecastPoint{
			CollectedAt: lastTime.Add(time.Duration(int64(m)*stepSeconds) * time.Second),
			Value:       clampNonNegative(predictedValue),
		})
	}

	return out, stepSeconds, nil
}

func forecastDampedHolt(
	history []model.MetricSeriesPoint,
	steps int,
	alpha float64,
	beta float64,
	phi float64,
) ([]model.ForecastPoint, int64, error) {
	if len(history) < 3 {
		return nil, 0, ErrInsufficientForecastData
	}
	if steps <= 0 {
		steps = 8
	}
	if alpha <= 0 || alpha >= 1 {
		alpha = 0.55
	}
	if beta <= 0 || beta >= 1 {
		beta = 0.2
	}
	if phi <= 0 || phi >= 1 {
		phi = 0.85
	}

	stepSeconds := medianStepSeconds(history)

	level := history[0].Value
	trend := history[1].Value - history[0].Value

	for i := 1; i < len(history); i++ {
		value := history[i].Value
		prevLevel := level

		level = alpha*value + (1-alpha)*(level+phi*trend)
		trend = beta*(level-prevLevel) + (1-beta)*phi*trend
	}

	lastTime := history[len(history)-1].CollectedAt
	out := make([]model.ForecastPoint, 0, steps)

	for m := 1; m <= steps; m++ {
		dampedTrend := 0.0
		for j := 1; j <= m; j++ {
			dampedTrend += math.Pow(phi, float64(j)) * trend
		}

		predictedValue := level + dampedTrend
		out = append(out, model.ForecastPoint{
			CollectedAt: lastTime.Add(time.Duration(int64(m)*stepSeconds) * time.Second),
			Value:       clampNonNegative(predictedValue),
		})
	}

	return out, stepSeconds, nil
}

type forecastRunner func([]model.MetricSeriesPoint, int) ([]model.ForecastPoint, int64, error)

type forecastCandidate struct {
	name string
	run  forecastRunner
}

func smape(actual, predicted float64) float64 {
	den := math.Abs(actual) + math.Abs(predicted)
	if den == 0 {
		return 0
	}
	return 2 * math.Abs(actual-predicted) / den
}

func scoreForecast(actual []model.MetricSeriesPoint, predicted []model.ForecastPoint) float64 {
	n := len(actual)
	if len(predicted) < n {
		n = len(predicted)
	}
	if n == 0 {
		return math.Inf(1)
	}

	total := 0.0
	weightSum := 0.0
	for i := 0; i < n; i++ {
		w := float64(i + 1)
		total += w * smape(actual[i].Value, predicted[i].Value)
		weightSum += w
	}

	if weightSum == 0 {
		return math.Inf(1)
	}
	return total / weightSum
}

func bestHoltLinearRunner() forecastRunner {
	alphas := []float64{0.2, 0.4, 0.6, 0.8}
	betas := []float64{0.1, 0.2, 0.3, 0.5}

	return func(history []model.MetricSeriesPoint, steps int) ([]model.ForecastPoint, int64, error) {
		bestScore := math.Inf(1)
		bestAlpha := 0.6
		bestBeta := 0.3

		holdout := minInt(maxInt(len(history)/5, 3), 12)
		if len(history) < holdout+3 {
			return forecastHoltLinear(history, steps, bestAlpha, bestBeta)
		}

		subTrain := history[:len(history)-holdout]
		subTest := history[len(history)-holdout:]

		for _, a := range alphas {
			for _, b := range betas {
				pred, _, err := forecastHoltLinear(subTrain, len(subTest), a, b)
				if err != nil {
					continue
				}
				score := scoreForecast(subTest, pred)
				if score < bestScore {
					bestScore = score
					bestAlpha = a
					bestBeta = b
				}
			}
		}

		return forecastHoltLinear(history, steps, bestAlpha, bestBeta)
	}
}

func bestDampedHoltRunner() forecastRunner {
	alphas := []float64{0.2, 0.4, 0.6, 0.8}
	betas := []float64{0.1, 0.2, 0.3}
	phis := []float64{0.7, 0.8, 0.9, 0.95}

	return func(history []model.MetricSeriesPoint, steps int) ([]model.ForecastPoint, int64, error) {
		bestScore := math.Inf(1)
		bestAlpha := 0.55
		bestBeta := 0.2
		bestPhi := 0.85

		holdout := minInt(maxInt(len(history)/5, 3), 12)
		if len(history) < holdout+3 {
			return forecastDampedHolt(history, steps, bestAlpha, bestBeta, bestPhi)
		}

		subTrain := history[:len(history)-holdout]
		subTest := history[len(history)-holdout:]

		for _, a := range alphas {
			for _, b := range betas {
				for _, p := range phis {
					pred, _, err := forecastDampedHolt(subTrain, len(subTest), a, b, p)
					if err != nil {
						continue
					}
					score := scoreForecast(subTest, pred)
					if score < bestScore {
						bestScore = score
						bestAlpha = a
						bestBeta = b
						bestPhi = p
					}
				}
			}
		}

		return forecastDampedHolt(history, steps, bestAlpha, bestBeta, bestPhi)
	}
}

func forecastAuto(
	history []model.MetricSeriesPoint,
	steps int,
	metricName string,
) ([]model.ForecastPoint, int64, string, error) {
	if len(history) < 3 {
		return nil, 0, "", ErrInsufficientForecastData
	}

	metricName = strings.ToLower(strings.TrimSpace(metricName))
	window := defaultForecastWindow(len(history))

	candidates := []forecastCandidate{
		{
			name: "weighted_moving_average",
			run: func(h []model.MetricSeriesPoint, s int) ([]model.ForecastPoint, int64, error) {
				return forecastWeightedMovingAverage(h, s, defaultForecastWindow(len(h)))
			},
		},
		{
			name: "moving_average",
			run: func(h []model.MetricSeriesPoint, s int) ([]model.ForecastPoint, int64, error) {
				return forecastMovingAverage(h, s, defaultForecastWindow(len(h)))
			},
		},
		{
			name: "damped_holt",
			run:  bestDampedHoltRunner(),
		},
	}

	if !strings.Contains(metricName, "cpu") {
		candidates = append(candidates, forecastCandidate{
			name: "holt_linear",
			run:  bestHoltLinearRunner(),
		})
	}

	holdout := minInt(maxInt(len(history)/5, 3), 12)
	if len(history) < holdout+3 {
		pred, step, err := forecastWeightedMovingAverage(history, steps, window)
		return pred, step, "weighted_moving_average", err
	}

	train := history[:len(history)-holdout]
	test := history[len(history)-holdout:]

	bestScore := math.Inf(1)
	bestName := ""
	var bestRunner forecastRunner

	for _, c := range candidates {
		pred, _, err := c.run(train, len(test))
		if err != nil {
			continue
		}

		score := scoreForecast(test, pred)
		if score < bestScore {
			bestScore = score
			bestName = c.name
			bestRunner = c.run
		}
	}

	if bestRunner == nil {
		pred, step, err := forecastWeightedMovingAverage(history, steps, window)
		return pred, step, "weighted_moving_average", err
	}

	pred, step, err := bestRunner(history, steps)
	return pred, step, bestName, err
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
