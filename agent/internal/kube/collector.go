package kube

import (
	"context"
	"time"

	"agent/internal/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CollectSamples(ctx context.Context, c *Clients) ([]model.MetricPointPayload, error) {
	now := time.Now().UTC()
	out := make([]model.MetricPointPayload, 0, 256)

	nodeMetrics, err := c.Metrics.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, nm := range nodeMetrics.Items {
		if cpu := nm.Usage.Cpu(); cpu != nil {
			out = append(out, model.MetricPointPayload{
				MetricName:   "kube_node_cpu_usage_cores",
				MetricType:   "gauge",
				Unit:         "cores",
				ResourceKind: "node",
				NodeName:     nm.Name,
				CollectedAt:  now,
				Value:        float64(cpu.MilliValue()) / 1000.0,
			})
		}

		if mem := nm.Usage.Memory(); mem != nil {
			out = append(out, model.MetricPointPayload{
				MetricName:   "kube_node_memory_usage_bytes",
				MetricType:   "gauge",
				Unit:         "bytes",
				ResourceKind: "node",
				NodeName:     nm.Name,
				CollectedAt:  now,
				Value:        float64(mem.Value()),
			})
		}
	}

	podMetrics, err := c.Metrics.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pm := range podMetrics.Items {
		var totalMilliCPU int64
		var totalMemBytes int64

		for _, ctr := range pm.Containers {
			if cpu := ctr.Usage.Cpu(); cpu != nil {
				totalMilliCPU += cpu.MilliValue()
			}
			if mem := ctr.Usage.Memory(); mem != nil {
				totalMemBytes += mem.Value()
			}
		}

		out = append(out, model.MetricPointPayload{
			MetricName:   "kube_pod_cpu_usage_cores",
			MetricType:   "gauge",
			Unit:         "cores",
			ResourceKind: "pod",
			Namespace:    pm.Namespace,
			PodName:      pm.Name,
			CollectedAt:  now,
			Value:        float64(totalMilliCPU) / 1000.0,
		})

		out = append(out, model.MetricPointPayload{
			MetricName:   "kube_pod_memory_usage_bytes",
			MetricType:   "gauge",
			Unit:         "bytes",
			ResourceKind: "pod",
			Namespace:    pm.Namespace,
			PodName:      pm.Name,
			CollectedAt:  now,
			Value:        float64(totalMemBytes),
		})
	}

	return out, nil
}
