package service

import (
	"context"
	"errors"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
)

var ErrInvalidAnalysisQuery = errors.New("invalid analysis query")

type AnalysisService interface {
	GetWorkloadUtilization(
		ctx context.Context,
		clusterID string,
		from time.Time,
		to time.Time,
	) (*model.ClusterUtilizationResponse, error)

	GetOverProvisionedWorkloads(
		ctx context.Context,
		clusterID string,
		from time.Time,
		to time.Time,
		thresholds model.OverProvisionThresholds,
	) (*model.OverProvisionResponse, error)

	GetUnderProvisionedWorkloads(
		ctx context.Context,
		clusterID string,
		from time.Time,
		to time.Time,
		thresholds model.UnderProvisionThresholds,
	) (*model.UnderProvisionResponse, error)

	GetRightSizingRecommendations(
		ctx context.Context,
		clusterID string,
		from time.Time,
		to time.Time,
		thresholds model.RecommendationThresholds,
		underThresholds model.UnderProvisionThresholds,
	) (*model.RecommendationResponse, error)

	GetClusterCapacity(
		ctx context.Context,
		clusterID string,
		from time.Time,
		to time.Time,
		recThresholds model.RecommendationThresholds,
		underThresholds model.UnderProvisionThresholds,
	) (*model.ClusterCapacityResponse, error)
}

type analysisService struct {
	metricRepo    repository.MetricRepository
	inventoryRepo repository.InventoryRepository
}

func NewAnalysisService(
	metricRepo repository.MetricRepository,
	inventoryRepo repository.InventoryRepository,
) AnalysisService {
	return &analysisService{
		metricRepo:    metricRepo,
		inventoryRepo: inventoryRepo,
	}
}

type workloadMetricBucket struct {
	namespace      string
	controllerUID  string
	controllerKind string
	controllerName string

	cpuValues    []float64
	cpuUnit      string
	memoryValues []float64
	memoryUnit   string
}

type workloadRequestBucket struct {
	namespace      string
	controllerUID  string
	controllerKind string
	controllerName string

	cpuRequestCores    float64
	memoryRequestBytes int64
}

func (s *analysisService) GetWorkloadUtilization(
	ctx context.Context,
	clusterID string,
	from time.Time,
	to time.Time,
) (*model.ClusterUtilizationResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, ErrInvalidAnalysisQuery
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, ErrInvalidAnalysisQuery
	}

	rows, err := s.metricRepo.GetClusterMetricSamples(ctx, clusterID, from, to)
	if err != nil {
		return nil, err
	}

	grouped := groupUsageRows(rows)

	items := make([]model.WorkloadUtilizationSummary, 0, len(grouped))
	for _, bucket := range grouped {
		item := model.WorkloadUtilizationSummary{
			Namespace:      bucket.namespace,
			ControllerUID:  bucket.controllerUID,
			ControllerKind: bucket.controllerKind,
			ControllerName: bucket.controllerName,
			WindowFrom:     from,
			WindowTo:       to,
		}

		if len(bucket.cpuValues) > 0 {
			item.CPU = buildResourceStats(bucket.cpuValues, bucket.cpuUnit)
		}
		if len(bucket.memoryValues) > 0 {
			item.Memory = buildResourceStats(bucket.memoryValues, bucket.memoryUnit)
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace == items[j].Namespace {
			if items[i].ControllerKind == items[j].ControllerKind {
				return items[i].ControllerName < items[j].ControllerName
			}
			return items[i].ControllerKind < items[j].ControllerKind
		}
		return items[i].Namespace < items[j].Namespace
	})

	return &model.ClusterUtilizationResponse{
		ClusterID: clusterID,
		From:      from,
		To:        to,
		Items:     items,
	}, nil
}

func (s *analysisService) GetOverProvisionedWorkloads(
	ctx context.Context,
	clusterID string,
	from time.Time,
	to time.Time,
	thresholds model.OverProvisionThresholds,
) (*model.OverProvisionResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, ErrInvalidAnalysisQuery
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, ErrInvalidAnalysisQuery
	}

	thresholds = normalizeOverProvisionThresholds(thresholds)

	rows, err := s.metricRepo.GetClusterMetricSamples(ctx, clusterID, from, to)
	if err != nil {
		return nil, err
	}
	log.Printf("GetOverProvisionedWorkloads metrics rows=%d", len(rows))

	usageGrouped := groupUsageRows(rows)
	log.Printf("GetOverProvisionedWorkloads usageGrouped=%d", len(usageGrouped))

	snapshot, err := s.inventoryRepo.GetLatestSnapshotByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	log.Printf("GetOverProvisionedWorkloads snapshotID=%s", snapshot.ID)

	pods, err := s.inventoryRepo.GetPodsBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("GetOverProvisionedWorkloads pods=%d", len(pods))

	containers, err := s.inventoryRepo.GetContainersBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("GetOverProvisionedWorkloads containers=%d", len(containers))

	requestsGrouped := groupWorkloadRequests(pods, containers)
	log.Printf("GetOverProvisionedWorkloads requestsGrouped=%d", len(requestsGrouped))

	items := make([]model.OverProvisionFinding, 0)

	for key, usage := range usageGrouped {
		requests, ok := requestsGrouped[key]
		if !ok {
			log.Printf("GetOverProvisionedWorkloads skip key=%s reason=no matching requests", key)
			continue
		}

		cpuStats := buildResourceStats(usage.cpuValues, usage.cpuUnit)
		memStats := buildResourceStats(usage.memoryValues, usage.memoryUnit)

		if cpuStats == nil && memStats == nil {
			log.Printf("GetOverProvisionedWorkloads skip key=%s reason=no cpu or memory stats", key)
			continue
		}

		finding := model.OverProvisionFinding{
			Namespace:      usage.namespace,
			ControllerUID:  usage.controllerUID,
			ControllerKind: usage.controllerKind,
			ControllerName: usage.controllerName,
			WindowFrom:     from,
			WindowTo:       to,
		}

		if cpuStats != nil {
			finding.CurrentCPURequestCores = requests.cpuRequestCores
			finding.ObservedCPUP95Cores = cpuStats.P95
			finding.RecommendedCPURequestCores = cpuStats.P95 * thresholds.CPUSafetyMargin
			finding.ReclaimableCPUCores = clampNonNegative(
				finding.CurrentCPURequestCores - finding.RecommendedCPURequestCores,
			)

			if finding.RecommendedCPURequestCores > 0 &&
				finding.CurrentCPURequestCores >= finding.RecommendedCPURequestCores*thresholds.CPUOverprovisionRatio &&
				finding.ReclaimableCPUCores >= thresholds.MinReclaimCPUCores {
				finding.CPUOverProvisioned = true
			}

			log.Printf(
				"GetOverProvisionedWorkloads cpu key=%s current=%.4f p95=%.4f recommended=%.4f reclaimable=%.4f over=%t",
				key,
				finding.CurrentCPURequestCores,
				finding.ObservedCPUP95Cores,
				finding.RecommendedCPURequestCores,
				finding.ReclaimableCPUCores,
				finding.CPUOverProvisioned,
			)
		}

		if memStats != nil {
			finding.CurrentMemoryRequestBytes = requests.memoryRequestBytes
			finding.ObservedMemoryP95Bytes = int64(math.Round(memStats.P95))
			finding.RecommendedMemoryRequestBytes = int64(math.Round(memStats.P95 * thresholds.MemorySafetyMargin))
			finding.ReclaimableMemoryBytes = clampNonNegativeInt64(
				finding.CurrentMemoryRequestBytes - finding.RecommendedMemoryRequestBytes,
			)

			if finding.RecommendedMemoryRequestBytes > 0 &&
				float64(finding.CurrentMemoryRequestBytes) >= float64(finding.RecommendedMemoryRequestBytes)*thresholds.MemoryOverprovisionRatio &&
				finding.ReclaimableMemoryBytes >= thresholds.MinReclaimMemoryBytes {
				finding.MemoryOverProvisioned = true
			}

			log.Printf(
				"GetOverProvisionedWorkloads mem key=%s current=%d p95=%d recommended=%d reclaimable=%d over=%t",
				key,
				finding.CurrentMemoryRequestBytes,
				finding.ObservedMemoryP95Bytes,
				finding.RecommendedMemoryRequestBytes,
				finding.ReclaimableMemoryBytes,
				finding.MemoryOverProvisioned,
			)
		}

		if finding.CPUOverProvisioned || finding.MemoryOverProvisioned {
			log.Printf(
				"GetOverProvisionedWorkloads add key=%s namespace=%s kind=%s name=%s",
				key,
				finding.Namespace,
				finding.ControllerKind,
				finding.ControllerName,
			)
			items = append(items, finding)
		} else {
			log.Printf(
				"GetOverProvisionedWorkloads skip key=%s reason=below thresholds cpuOver=%t memOver=%t",
				key,
				finding.CPUOverProvisioned,
				finding.MemoryOverProvisioned,
			)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace == items[j].Namespace {
			if items[i].ControllerKind == items[j].ControllerKind {
				return items[i].ControllerName < items[j].ControllerName
			}
			return items[i].ControllerKind < items[j].ControllerKind
		}
		return items[i].Namespace < items[j].Namespace
	})

	log.Printf("GetOverProvisionedWorkloads final items=%d", len(items))

	return &model.OverProvisionResponse{
		ClusterID:  clusterID,
		From:       from,
		To:         to,
		Thresholds: thresholds,
		Items:      items,
	}, nil
}

func normalizeOverProvisionThresholds(in model.OverProvisionThresholds) model.OverProvisionThresholds {
	if in.CPUSafetyMargin <= 0 {
		in.CPUSafetyMargin = 1.20
	}
	if in.MemorySafetyMargin <= 0 {
		in.MemorySafetyMargin = 1.15
	}
	if in.CPUOverprovisionRatio <= 0 {
		in.CPUOverprovisionRatio = 1.50
	}
	if in.MemoryOverprovisionRatio <= 0 {
		in.MemoryOverprovisionRatio = 1.50
	}
	if in.MinReclaimCPUCores <= 0 {
		in.MinReclaimCPUCores = 0.10
	}
	if in.MinReclaimMemoryBytes <= 0 {
		in.MinReclaimMemoryBytes = 128 * 1024 * 1024
	}
	return in
}

func groupUsageRows(rows []model.MetricSampleRow) map[string]*workloadMetricBucket {
	grouped := make(map[string]*workloadMetricBucket)

	for _, row := range rows {
		if row.ControllerUID == nil || strings.TrimSpace(*row.ControllerUID) == "" {
			continue
		}
		if row.Namespace == nil || strings.TrimSpace(*row.Namespace) == "" {
			continue
		}

		key := workloadKey(
			deref(row.Namespace),
			deref(row.ControllerUID),
			deref(row.ControllerKind),
			deref(row.ControllerName),
		)

		bucket, ok := grouped[key]
		if !ok {
			bucket = &workloadMetricBucket{
				namespace:      deref(row.Namespace),
				controllerUID:  deref(row.ControllerUID),
				controllerKind: deref(row.ControllerKind),
				controllerName: deref(row.ControllerName),
			}
			grouped[key] = bucket
		}

		switch strings.TrimSpace(row.MetricName) {
		case "kube_pod_cpu_usage_cores":
			bucket.cpuValues = append(bucket.cpuValues, row.Value)
			bucket.cpuUnit = row.Unit
		case "kube_pod_memory_usage_bytes":
			bucket.memoryValues = append(bucket.memoryValues, row.Value)
			bucket.memoryUnit = row.Unit
		}
	}

	return grouped
}

func groupWorkloadRequests(
	pods []model.PodInventory,
	containers []model.ContainerInventory,
) map[string]*workloadRequestBucket {
	podMap := make(map[string]model.PodInventory, len(pods))
	for _, pod := range pods {
		podMap[pod.ID] = pod
	}

	grouped := make(map[string]*workloadRequestBucket)

	for _, ctr := range containers {
		if strings.TrimSpace(ctr.ParentKind) != "pod" {
			continue
		}
		pod, ok := podMap[ctr.ParentRefID]
		if !ok {
			continue
		}
		if pod.ControllerUID == nil || strings.TrimSpace(*pod.ControllerUID) == "" {
			continue
		}

		key := workloadKey(
			pod.Namespace,
			deref(pod.ControllerUID),
			deref(pod.ControllerKind),
			deref(pod.ControllerName),
		)

		bucket, ok := grouped[key]
		if !ok {
			bucket = &workloadRequestBucket{
				namespace:      pod.Namespace,
				controllerUID:  deref(pod.ControllerUID),
				controllerKind: deref(pod.ControllerKind),
				controllerName: deref(pod.ControllerName),
			}
			grouped[key] = bucket
		}

		if ctr.CPURequestMillicores != nil {
			bucket.cpuRequestCores += float64(*ctr.CPURequestMillicores) / 1000.0
		}
		if ctr.MemoryRequestBytes != nil {
			bucket.memoryRequestBytes += *ctr.MemoryRequestBytes
		}
	}

	return grouped
}

func workloadKey(namespace, controllerUID, controllerKind, controllerName string) string {
	return namespace + "|" + controllerUID + "|" + controllerKind + "|" + controllerName
}

func buildResourceStats(values []float64, unit string) *model.ResourceStats {
	clean := sanitizeValues(values)
	if len(clean) == 0 {
		return nil
	}

	sum := 0.0
	maxVal := clean[0]
	for _, v := range clean {
		sum += v
		if v > maxVal {
			maxVal = v
		}
	}

	sorted := make([]float64, len(clean))
	copy(sorted, clean)
	sort.Float64s(sorted)

	return &model.ResourceStats{
		Average:     sum / float64(len(clean)),
		Maximum:     maxVal,
		P50:         quantileSorted(sorted, 0.50),
		P95:         quantileSorted(sorted, 0.95),
		P99:         quantileSorted(sorted, 0.99),
		SampleCount: len(clean),
		Unit:        unit,
	}
}

func sanitizeValues(values []float64) []float64 {
	out := make([]float64, 0, len(values))
	for _, v := range values {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			continue
		}
		if v < 0 {
			continue
		}
		out = append(out, v)
	}
	return out
}

func quantileSorted(sortedVals []float64, q float64) float64 {
	if len(sortedVals) == 0 {
		return 0
	}
	if q <= 0 {
		return sortedVals[0]
	}
	if q >= 1 {
		return sortedVals[len(sortedVals)-1]
	}

	pos := q * float64(len(sortedVals)-1)
	i := int(math.Floor(pos))
	j := int(math.Ceil(pos))
	if i == j {
		return sortedVals[i]
	}

	frac := pos - float64(i)
	return sortedVals[i] + frac*(sortedVals[j]-sortedVals[i])
}

func deref(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

func clampNonNegativeInt64(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}

type workloadLimitBucket struct {
	namespace      string
	controllerUID  string
	controllerKind string
	controllerName string

	cpuLimitCores    float64
	memoryLimitBytes int64
}

type workloadRuntimeBucket struct {
	namespace      string
	controllerUID  string
	controllerKind string
	controllerName string

	memoryOOMDetected bool
	oomContainerCount int
	restartCount      int32
}

func (s *analysisService) GetUnderProvisionedWorkloads(
	ctx context.Context,
	clusterID string,
	from time.Time,
	to time.Time,
	thresholds model.UnderProvisionThresholds,
) (*model.UnderProvisionResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, ErrInvalidAnalysisQuery
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, ErrInvalidAnalysisQuery
	}

	thresholds = normalizeUnderProvisionThresholds(thresholds)

	rows, err := s.metricRepo.GetClusterMetricSamples(ctx, clusterID, from, to)
	if err != nil {
		return nil, err
	}
	log.Printf("GetUnderProvisionedWorkloads metrics rows=%d", len(rows))

	usageGrouped := groupUsageRows(rows)
	log.Printf("GetUnderProvisionedWorkloads usageGrouped=%d", len(usageGrouped))

	snapshot, err := s.inventoryRepo.GetLatestSnapshotByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	log.Printf("GetUnderProvisionedWorkloads snapshotID=%s", snapshot.ID)

	pods, err := s.inventoryRepo.GetPodsBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("GetUnderProvisionedWorkloads pods=%d", len(pods))

	containers, err := s.inventoryRepo.GetContainersBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("GetUnderProvisionedWorkloads containers=%d", len(containers))

	containerStatuses, err := s.inventoryRepo.GetContainerStatusesBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}
	log.Printf("GetUnderProvisionedWorkloads containerStatuses=%d", len(containerStatuses))

	limitsGrouped := groupWorkloadLimits(pods, containers)
	runtimeGrouped := groupWorkloadRuntimeSignals(pods, containerStatuses)

	log.Printf("GetUnderProvisionedWorkloads limitsGrouped=%d", len(limitsGrouped))
	log.Printf("GetUnderProvisionedWorkloads runtimeGrouped=%d", len(runtimeGrouped))

	items := make([]model.UnderProvisionFinding, 0)

	for key, usage := range usageGrouped {
		limits, hasLimits := limitsGrouped[key]
		runtimeSignals, hasRuntimeSignals := runtimeGrouped[key]

		if !hasLimits && !hasRuntimeSignals {
			log.Printf("GetUnderProvisionedWorkloads skip key=%s reason=no limits and no runtime signals", key)
			continue
		}

		finding := model.UnderProvisionFinding{
			Namespace:      usage.namespace,
			ControllerUID:  usage.controllerUID,
			ControllerKind: usage.controllerKind,
			ControllerName: usage.controllerName,
			WindowFrom:     from,
			WindowTo:       to,
		}

		if hasLimits {
			finding.CurrentCPULimitCores = limits.cpuLimitCores
			finding.CurrentMemoryLimitBytes = limits.memoryLimitBytes

			if limits.cpuLimitCores > 0 && len(usage.cpuValues) > 0 {
				pressured := 0
				total := 0
				for _, v := range sanitizeValues(usage.cpuValues) {
					total++
					if v/limits.cpuLimitCores >= thresholds.CPUPressureThreshold {
						pressured++
					}
				}
				if total > 0 {
					finding.CPUPressureFrequency = float64(pressured) / float64(total)
					if finding.CPUPressureFrequency >= thresholds.CPUPressureFrequencyThreshold {
						finding.CPUUnderProvisioned = true
					}
				}
			}

			if limits.memoryLimitBytes > 0 && len(usage.memoryValues) > 0 {
				pressured := 0
				total := 0
				limit := float64(limits.memoryLimitBytes)
				for _, v := range sanitizeValues(usage.memoryValues) {
					total++
					if v/limit >= thresholds.MemoryPressureThreshold {
						pressured++
					}
				}
				if total > 0 {
					finding.MemoryPressureFrequency = float64(pressured) / float64(total)
					if finding.MemoryPressureFrequency >= thresholds.MemoryPressureFrequencyThreshold {
						finding.MemoryUnderProvisioned = true
					}
				}
			}

			log.Printf(
				"GetUnderProvisionedWorkloads pressure key=%s cpuLimit=%.4f cpuFreq=%.4f memLimit=%d memFreq=%.4f cpuUnder=%t memUnder=%t",
				key,
				finding.CurrentCPULimitCores,
				finding.CPUPressureFrequency,
				finding.CurrentMemoryLimitBytes,
				finding.MemoryPressureFrequency,
				finding.CPUUnderProvisioned,
				finding.MemoryUnderProvisioned,
			)
		}

		if hasRuntimeSignals {
			finding.MemoryOOMDetected = runtimeSignals.memoryOOMDetected
			finding.OOMContainerCount = runtimeSignals.oomContainerCount
			finding.RestartCount = runtimeSignals.restartCount

			if runtimeSignals.memoryOOMDetected {
				finding.MemoryUnderProvisioned = true
			}

			log.Printf(
				"GetUnderProvisionedWorkloads runtime key=%s oom=%t oomContainers=%d restarts=%d",
				key,
				finding.MemoryOOMDetected,
				finding.OOMContainerCount,
				finding.RestartCount,
			)
		}

		if finding.MemoryOOMDetected {
			finding.Reason = "memory-related restarts detected"
		} else if finding.MemoryUnderProvisioned && finding.CPUUnderProvisioned {
			finding.Reason = "cpu and memory usage frequently approach limits"
		} else if finding.MemoryUnderProvisioned {
			finding.Reason = "memory usage frequently approaches limits"
		} else if finding.CPUUnderProvisioned {
			finding.Reason = "cpu usage frequently approaches limits"
		}

		if finding.CPUUnderProvisioned || finding.MemoryUnderProvisioned {
			log.Printf(
				"GetUnderProvisionedWorkloads add key=%s namespace=%s kind=%s name=%s reason=%s",
				key,
				finding.Namespace,
				finding.ControllerKind,
				finding.ControllerName,
				finding.Reason,
			)
			items = append(items, finding)
		} else {
			log.Printf("GetUnderProvisionedWorkloads skip key=%s reason=below thresholds", key)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace == items[j].Namespace {
			if items[i].ControllerKind == items[j].ControllerKind {
				return items[i].ControllerName < items[j].ControllerName
			}
			return items[i].ControllerKind < items[j].ControllerKind
		}
		return items[i].Namespace < items[j].Namespace
	})

	log.Printf("GetUnderProvisionedWorkloads final items=%d", len(items))

	return &model.UnderProvisionResponse{
		ClusterID:  clusterID,
		From:       from,
		To:         to,
		Thresholds: thresholds,
		Items:      items,
	}, nil
}

func normalizeUnderProvisionThresholds(in model.UnderProvisionThresholds) model.UnderProvisionThresholds {
	if in.CPUPressureThreshold <= 0 {
		in.CPUPressureThreshold = 0.90
	}
	if in.MemoryPressureThreshold <= 0 {
		in.MemoryPressureThreshold = 0.90
	}
	if in.CPUPressureFrequencyThreshold <= 0 {
		in.CPUPressureFrequencyThreshold = 0.10
	}
	if in.MemoryPressureFrequencyThreshold <= 0 {
		in.MemoryPressureFrequencyThreshold = 0.10
	}
	return in
}

func groupWorkloadLimits(
	pods []model.PodInventory,
	containers []model.ContainerInventory,
) map[string]*workloadLimitBucket {
	podMap := make(map[string]model.PodInventory, len(pods))
	for _, pod := range pods {
		podMap[pod.ID] = pod
	}

	grouped := make(map[string]*workloadLimitBucket)

	for _, ctr := range containers {
		if strings.TrimSpace(ctr.ParentKind) != "pod" {
			continue
		}

		pod, ok := podMap[ctr.ParentRefID]
		if !ok {
			continue
		}
		if pod.ControllerUID == nil || strings.TrimSpace(*pod.ControllerUID) == "" {
			continue
		}

		key := workloadKey(
			pod.Namespace,
			deref(pod.ControllerUID),
			deref(pod.ControllerKind),
			deref(pod.ControllerName),
		)

		bucket, ok := grouped[key]
		if !ok {
			bucket = &workloadLimitBucket{
				namespace:      pod.Namespace,
				controllerUID:  deref(pod.ControllerUID),
				controllerKind: deref(pod.ControllerKind),
				controllerName: deref(pod.ControllerName),
			}
			grouped[key] = bucket
		}

		if ctr.CPULimitMillicores != nil {
			bucket.cpuLimitCores += float64(*ctr.CPULimitMillicores) / 1000.0
		}
		if ctr.MemoryLimitBytes != nil {
			bucket.memoryLimitBytes += *ctr.MemoryLimitBytes
		}
	}

	return grouped
}

func groupWorkloadRuntimeSignals(
	pods []model.PodInventory,
	statuses []model.ContainerStatusInventory,
) map[string]*workloadRuntimeBucket {
	podMap := make(map[string]model.PodInventory, len(pods))
	for _, pod := range pods {
		podMap[pod.ID] = pod
	}

	grouped := make(map[string]*workloadRuntimeBucket)

	for _, st := range statuses {
		pod, ok := podMap[st.PodRefID]
		if !ok {
			continue
		}
		if pod.ControllerUID == nil || strings.TrimSpace(*pod.ControllerUID) == "" {
			continue
		}

		key := workloadKey(
			pod.Namespace,
			deref(pod.ControllerUID),
			deref(pod.ControllerKind),
			deref(pod.ControllerName),
		)

		bucket, ok := grouped[key]
		if !ok {
			bucket = &workloadRuntimeBucket{
				namespace:      pod.Namespace,
				controllerUID:  deref(pod.ControllerUID),
				controllerKind: deref(pod.ControllerKind),
				controllerName: deref(pod.ControllerName),
			}
			grouped[key] = bucket
		}

		bucket.restartCount += st.RestartCount

		reason := strings.TrimSpace(deref(st.LastTerminationReason))
		if st.OOMKilled || strings.EqualFold(reason, "OOMKilled") {
			bucket.memoryOOMDetected = true
			bucket.oomContainerCount++
		}
	}

	return grouped
}

func (s *analysisService) GetRightSizingRecommendations(
	ctx context.Context,
	clusterID string,
	from time.Time,
	to time.Time,
	thresholds model.RecommendationThresholds,
	underThresholds model.UnderProvisionThresholds,
) (*model.RecommendationResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, ErrInvalidAnalysisQuery
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, ErrInvalidAnalysisQuery
	}

	thresholds = normalizeRecommendationThresholds(thresholds)
	underThresholds = normalizeUnderProvisionThresholds(underThresholds)

	rows, err := s.metricRepo.GetClusterMetricSamples(ctx, clusterID, from, to)
	if err != nil {
		return nil, err
	}
	usageGrouped := groupUsageRows(rows)

	snapshot, err := s.inventoryRepo.GetLatestSnapshotByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	pods, err := s.inventoryRepo.GetPodsBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	containers, err := s.inventoryRepo.GetContainersBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	containerStatuses, err := s.inventoryRepo.GetContainerStatusesBySnapshotID(ctx, snapshot.ID)
	if err != nil {
		return nil, err
	}

	requestsGrouped := groupWorkloadRequests(pods, containers)
	limitsGrouped := groupWorkloadLimits(pods, containers)
	runtimeGrouped := groupWorkloadRuntimeSignals(pods, containerStatuses)

	keys := make(map[string]struct{})
	for k := range usageGrouped {
		keys[k] = struct{}{}
	}
	for k := range requestsGrouped {
		keys[k] = struct{}{}
	}

	items := make([]model.RightSizingRecommendation, 0, len(keys))

	for key := range keys {
		usage := usageGrouped[key]
		requests := requestsGrouped[key]

		rec := model.RightSizingRecommendation{
			WindowFrom:             from,
			WindowTo:               to,
			CPUSafetyMarginUsed:    thresholds.CPUSafetyMargin,
			MemorySafetyMarginUsed: thresholds.MemorySafetyMargin,
		}

		if usage != nil {
			rec.Namespace = usage.namespace
			rec.ControllerUID = usage.controllerUID
			rec.ControllerKind = usage.controllerKind
			rec.ControllerName = usage.controllerName
		} else if requests != nil {
			rec.Namespace = requests.namespace
			rec.ControllerUID = requests.controllerUID
			rec.ControllerKind = requests.controllerKind
			rec.ControllerName = requests.controllerName
		}

		if requests != nil {
			rec.CurrentCPURequestCores = requests.cpuRequestCores
			rec.CurrentMemoryRequestBytes = requests.memoryRequestBytes
		}

		var cpuStats *model.ResourceStats
		var memStats *model.ResourceStats
		if usage != nil {
			cpuStats = buildResourceStats(usage.cpuValues, usage.cpuUnit)
			memStats = buildResourceStats(usage.memoryValues, usage.memoryUnit)
		}

		if cpuStats != nil {
			rec.ObservedCPUP95Cores = cpuStats.P95
			rec.RecommendedCPURequestCores = maxFloat64(
				cpuStats.P95*thresholds.CPUSafetyMargin,
				thresholds.MinCPURequestCores,
			)
		}

		if memStats != nil {
			rec.ObservedMemoryP95Bytes = int64(math.Round(memStats.P95))
			rec.RecommendedMemoryRequestBytes = maxInt64(
				int64(math.Round(memStats.P95*thresholds.MemorySafetyMargin)),
				thresholds.MinMemoryRequestBytes,
			)
		}

		eligible, reason, cpuUnder, memUnder, oomDetected := evaluateRecommendationEligibility(
			key,
			cpuStats,
			memStats,
			requests,
			limitsGrouped[key],
			runtimeGrouped[key],
			thresholds,
			underThresholds,
		)

		rec.Eligible = eligible
		rec.Reason = reason
		rec.CPUUnderProvisioned = cpuUnder
		rec.MemoryUnderProvisioned = memUnder
		rec.MemoryOOMDetected = oomDetected

		if !eligible {
			rec.RecommendedCPURequestCores = 0
			rec.RecommendedMemoryRequestBytes = 0
		}

		items = append(items, rec)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].Namespace == items[j].Namespace {
			if items[i].ControllerKind == items[j].ControllerKind {
				return items[i].ControllerName < items[j].ControllerName
			}
			return items[i].ControllerKind < items[j].ControllerKind
		}
		return items[i].Namespace < items[j].Namespace
	})

	return &model.RecommendationResponse{
		ClusterID:  clusterID,
		From:       from,
		To:         to,
		Thresholds: thresholds,
		Items:      items,
	}, nil
}

func normalizeRecommendationThresholds(in model.RecommendationThresholds) model.RecommendationThresholds {
	if in.CPUSafetyMargin <= 0 {
		in.CPUSafetyMargin = 1.20
	}
	if in.MemorySafetyMargin <= 0 {
		in.MemorySafetyMargin = 1.15
	}
	if in.MinCPURequestCores <= 0 {
		in.MinCPURequestCores = 0.05
	}
	if in.MinMemoryRequestBytes <= 0 {
		in.MinMemoryRequestBytes = 64 * 1024 * 1024
	}
	if in.MinSampleCount <= 0 {
		in.MinSampleCount = 5
	}
	return in
}

func evaluateRecommendationEligibility(
	key string,
	cpuStats *model.ResourceStats,
	memStats *model.ResourceStats,
	requests *workloadRequestBucket,
	limits *workloadLimitBucket,
	runtimeSignals *workloadRuntimeBucket,
	recThresholds model.RecommendationThresholds,
	underThresholds model.UnderProvisionThresholds,
) (eligible bool, reason string, cpuUnder bool, memUnder bool, oomDetected bool) {
	if requests == nil {
		return false, "no workload requests found", false, false, false
	}

	if cpuStats == nil && memStats == nil {
		return false, "no usage statistics available", false, false, false
	}

	if cpuStats != nil && cpuStats.SampleCount < recThresholds.MinSampleCount {
		return false, "insufficient cpu samples", false, false, false
	}
	if memStats != nil && memStats.SampleCount < recThresholds.MinSampleCount {
		return false, "insufficient memory samples", false, false, false
	}

	if requests.cpuRequestCores <= 0 && requests.memoryRequestBytes <= 0 {
		return false, "workload has no current requests", false, false, false
	}

	if limits != nil && cpuStats != nil && limits.cpuLimitCores > 0 {
		pressured := cpuStats.P95 / limits.cpuLimitCores
		if pressured >= underThresholds.CPUPressureThreshold {
			cpuUnder = true
		}
	}

	if limits != nil && memStats != nil && limits.memoryLimitBytes > 0 {
		pressured := memStats.P95 / float64(limits.memoryLimitBytes)
		if pressured >= underThresholds.MemoryPressureThreshold {
			memUnder = true
		}
	}

	if runtimeSignals != nil && runtimeSignals.memoryOOMDetected {
		oomDetected = true
		memUnder = true
	}

	if oomDetected {
		return false, "memory-related restarts detected", cpuUnder, memUnder, oomDetected
	}
	if cpuUnder && memUnder {
		return false, "cpu and memory are under-provisioned", cpuUnder, memUnder, oomDetected
	}
	if cpuUnder {
		return false, "cpu is under-provisioned", cpuUnder, memUnder, oomDetected
	}
	if memUnder {
		return false, "memory is under-provisioned", cpuUnder, memUnder, oomDetected
	}

	return true, "", cpuUnder, memUnder, oomDetected
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func maxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func (s *analysisService) GetClusterCapacity(
	ctx context.Context,
	clusterID string,
	from time.Time,
	to time.Time,
	recThresholds model.RecommendationThresholds,
	underThresholds model.UnderProvisionThresholds,
) (*model.ClusterCapacityResponse, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, ErrInvalidAnalysisQuery
	}
	if from.IsZero() || to.IsZero() || to.Before(from) {
		return nil, ErrInvalidAnalysisQuery
	}

	recs, err := s.GetRightSizingRecommendations(
		ctx,
		clusterID,
		from,
		to,
		recThresholds,
		underThresholds,
	)
	if err != nil {
		return nil, err
	}

	nsMap := make(map[string]*model.NamespaceCapacity)

	var totalCPU float64
	var totalMem int64
	var reclaimCPU float64
	var reclaimMem int64

	var totalWorkloads int
	var eligibleWorkloads int

	for _, item := range recs.Items {
		totalWorkloads++

		ns := item.Namespace
		bucket, ok := nsMap[ns]
		if !ok {
			bucket = &model.NamespaceCapacity{
				Namespace: ns,
			}
			nsMap[ns] = bucket
		}

		bucket.WorkloadCount++

		totalCPU += item.CurrentCPURequestCores
		totalMem += item.CurrentMemoryRequestBytes

		bucket.TotalCPURequestCores += item.CurrentCPURequestCores
		bucket.TotalMemoryRequestBytes += item.CurrentMemoryRequestBytes

		if !item.Eligible {
			continue
		}

		eligibleWorkloads++
		bucket.EligibleWorkloadCount++

		cpuDelta := item.CurrentCPURequestCores - item.RecommendedCPURequestCores
		memDelta := item.CurrentMemoryRequestBytes - item.RecommendedMemoryRequestBytes

		if cpuDelta > 0 {
			reclaimCPU += cpuDelta
			bucket.ReclaimableCPUCores += cpuDelta
		}

		if memDelta > 0 {
			reclaimMem += memDelta
			bucket.ReclaimableMemoryBytes += memDelta
		}
	}

	namespaces := make([]model.NamespaceCapacity, 0, len(nsMap))
	for _, v := range nsMap {
		namespaces = append(namespaces, *v)
	}

	sort.Slice(namespaces, func(i, j int) bool {
		return namespaces[i].Namespace < namespaces[j].Namespace
	})

	return &model.ClusterCapacityResponse{
		ClusterID: clusterID,
		From:      from,
		To:        to,

		Namespaces: namespaces,

		TotalCPURequestCores:    totalCPU,
		TotalMemoryRequestBytes: totalMem,

		ReclaimableCPUCores:    reclaimCPU,
		ReclaimableMemoryBytes: reclaimMem,

		EligibleWorkloads: eligibleWorkloads,
		TotalWorkloads:    totalWorkloads,
	}, nil
}
