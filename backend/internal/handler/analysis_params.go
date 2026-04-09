package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"backend/internal/model"

	"github.com/go-chi/chi/v5"
)

type analysisRequestContext struct {
	OwnerID   int64
	ClusterID string
	From      time.Time
	To        time.Time
}

func parseAnalysisRequestContext(r *http.Request) (analysisRequestContext, error) {
	ownerID, ok := getAuthenticatedUserID(r)
	if !ok || ownerID <= 0 {
		return analysisRequestContext{}, newBadRequestError("unauthorized")
	}

	clusterID := chi.URLParam(r, "id")
	if clusterID == "" {
		return analysisRequestContext{}, newBadRequestError("invalid cluster id")
	}

	from, to, err := parseRequiredTimeRange(r)
	if err != nil {
		return analysisRequestContext{}, err
	}

	return analysisRequestContext{
		OwnerID:   ownerID,
		ClusterID: clusterID,
		From:      from,
		To:        to,
	}, nil
}

func parseRequiredTimeRange(r *http.Request) (time.Time, time.Time, error) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	if fromStr == "" || toStr == "" {
		return time.Time{}, time.Time{}, newBadRequestError("from and to are required")
	}

	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		return time.Time{}, time.Time{}, newBadRequestError("invalid from")
	}

	to, err := time.Parse(time.RFC3339, toStr)
	if err != nil {
		return time.Time{}, time.Time{}, newBadRequestError("invalid to")
	}

	if !from.Before(to) {
		return time.Time{}, time.Time{}, newBadRequestError("from must be before to")
	}

	return from, to, nil
}

type badRequestError struct {
	message string
}

func (e *badRequestError) Error() string {
	return e.message
}

func newBadRequestError(message string) error {
	return &badRequestError{message: message}
}

func parseOptionalFloat64(r *http.Request, key string, fallback float64) (float64, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}

	v, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, newBadRequestError(fmt.Sprintf("invalid %s", key))
	}

	return v, nil
}

func parseOptionalInt64(r *http.Request, key string, fallback int64) (int64, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}

	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, newBadRequestError(fmt.Sprintf("invalid %s", key))
	}

	return v, nil
}

func parseOptionalInt(r *http.Request, key string, fallback int) (int, error) {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback, nil
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, newBadRequestError(fmt.Sprintf("invalid %s", key))
	}

	return v, nil
}

func parseRecommendationThresholds(r *http.Request) (model.RecommendationThresholds, error) {
	cpuSafetyMargin, err := parseOptionalFloat64(r, "cpuSafetyMargin", 1.20)
	if err != nil {
		return model.RecommendationThresholds{}, err
	}

	memorySafetyMargin, err := parseOptionalFloat64(r, "memorySafetyMargin", 1.15)
	if err != nil {
		return model.RecommendationThresholds{}, err
	}

	minCPURequestCores, err := parseOptionalFloat64(r, "minCpuRequestCores", 0.05)
	if err != nil {
		return model.RecommendationThresholds{}, err
	}

	minMemoryRequestBytes, err := parseOptionalInt64(r, "minMemoryRequestBytes", 64*1024*1024)
	if err != nil {
		return model.RecommendationThresholds{}, err
	}

	minSampleCount, err := parseOptionalInt(r, "minSampleCount", 5)
	if err != nil {
		return model.RecommendationThresholds{}, err
	}

	return model.RecommendationThresholds{
		CPUSafetyMargin:       cpuSafetyMargin,
		MemorySafetyMargin:    memorySafetyMargin,
		MinCPURequestCores:    minCPURequestCores,
		MinMemoryRequestBytes: minMemoryRequestBytes,
		MinSampleCount:        minSampleCount,
	}, nil
}

func parseCapacityRecommendationThresholds(r *http.Request) (model.RecommendationThresholds, error) {
	cpuSafetyMargin, err := parseOptionalFloat64(r, "cpuSafetyMargin", 1.20)
	if err != nil {
		return model.RecommendationThresholds{}, err
	}

	memorySafetyMargin, err := parseOptionalFloat64(r, "memorySafetyMargin", 1.15)
	if err != nil {
		return model.RecommendationThresholds{}, err
	}

	return model.RecommendationThresholds{
		CPUSafetyMargin:    cpuSafetyMargin,
		MemorySafetyMargin: memorySafetyMargin,
	}, nil
}

func parseUnderProvisionThresholds(r *http.Request) (model.UnderProvisionThresholds, error) {
	cpuPressureThreshold, err := parseOptionalFloat64(r, "cpuPressureThreshold", 0.90)
	if err != nil {
		return model.UnderProvisionThresholds{}, err
	}

	memoryPressureThreshold, err := parseOptionalFloat64(r, "memoryPressureThreshold", 0.90)
	if err != nil {
		return model.UnderProvisionThresholds{}, err
	}

	cpuPressureFrequencyThreshold, err := parseOptionalFloat64(r, "cpuPressureFrequencyThreshold", 0.10)
	if err != nil {
		return model.UnderProvisionThresholds{}, err
	}

	memoryPressureFrequencyThreshold, err := parseOptionalFloat64(r, "memoryPressureFrequencyThreshold", 0.10)
	if err != nil {
		return model.UnderProvisionThresholds{}, err
	}

	return model.UnderProvisionThresholds{
		CPUPressureThreshold:             cpuPressureThreshold,
		MemoryPressureThreshold:          memoryPressureThreshold,
		CPUPressureFrequencyThreshold:    cpuPressureFrequencyThreshold,
		MemoryPressureFrequencyThreshold: memoryPressureFrequencyThreshold,
	}, nil
}

func parseOverProvisionThresholds(r *http.Request) (model.OverProvisionThresholds, error) {
	cpuSafetyMargin, err := parseOptionalFloat64(r, "cpuSafetyMargin", 1.20)
	if err != nil {
		return model.OverProvisionThresholds{}, err
	}

	memorySafetyMargin, err := parseOptionalFloat64(r, "memorySafetyMargin", 1.15)
	if err != nil {
		return model.OverProvisionThresholds{}, err
	}

	cpuOverprovisionRatio, err := parseOptionalFloat64(r, "cpuOverprovisionRatio", 1.50)
	if err != nil {
		return model.OverProvisionThresholds{}, err
	}

	memoryOverprovisionRatio, err := parseOptionalFloat64(r, "memoryOverprovisionRatio", 1.50)
	if err != nil {
		return model.OverProvisionThresholds{}, err
	}

	minReclaimCPUCores, err := parseOptionalFloat64(r, "minReclaimCpuCores", 0.10)
	if err != nil {
		return model.OverProvisionThresholds{}, err
	}

	minReclaimMemoryBytes, err := parseOptionalInt64(r, "minReclaimMemoryBytes", 128*1024*1024)
	if err != nil {
		return model.OverProvisionThresholds{}, err
	}

	return model.OverProvisionThresholds{
		CPUSafetyMargin:          cpuSafetyMargin,
		MemorySafetyMargin:       memorySafetyMargin,
		CPUOverprovisionRatio:    cpuOverprovisionRatio,
		MemoryOverprovisionRatio: memoryOverprovisionRatio,
		MinReclaimCPUCores:       minReclaimCPUCores,
		MinReclaimMemoryBytes:    minReclaimMemoryBytes,
	}, nil
}
