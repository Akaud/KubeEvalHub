package kube

import (
	"context"
	"time"

	"agent/internal/model"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type podMetricMeta struct {
	UID            string
	Namespace      string
	PodName        string
	NodeName       string
	ControllerUID  string
	ControllerKind string
	ControllerName string
}

func CollectSamples(ctx context.Context, c *Clients) ([]model.MetricPointPayload, error) {
	now := time.Now().UTC()
	out := make([]model.MetricPointPayload, 0, 512)

	podMetaByKey, err := collectPodMetricMeta(ctx, c)
	if err != nil {
		return nil, err
	}

	nodeMetrics, err := c.Metrics.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, nm := range nodeMetrics.Items {
		collectedAt := metricsTimestampOrNow(nm.Timestamp.Time, now)

		if cpu := nm.Usage.Cpu(); cpu != nil {
			out = append(out, model.MetricPointPayload{
				MetricName:   "kube_node_cpu_usage_cores",
				MetricType:   "gauge",
				Unit:         "cores",
				ResourceKind: "node",
				NodeName:     nm.Name,
				CollectedAt:  collectedAt,
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
				CollectedAt:  collectedAt,
				Value:        float64(mem.Value()),
			})
		}
	}

	podMetrics, err := c.Metrics.MetricsV1beta1().PodMetricses("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pm := range podMetrics.Items {
		collectedAt := metricsTimestampOrNow(pm.Timestamp.Time, now)

		meta, hasMeta := podMetaByKey[podMetricKey(pm.Namespace, pm.Name)]

		var totalMilliCPU int64
		var totalMemBytes int64

		for _, ctr := range pm.Containers {
			if cpu := ctr.Usage.Cpu(); cpu != nil {
				cpuMilli := cpu.MilliValue()
				totalMilliCPU += cpuMilli

				out = append(out, model.MetricPointPayload{
					MetricName:     "kube_container_cpu_usage_cores",
					MetricType:     "gauge",
					Unit:           "cores",
					ResourceKind:   "container",
					NodeName:       metricMetaNodeName(meta, hasMeta),
					Namespace:      pm.Namespace,
					PodName:        pm.Name,
					PodUID:         metricMetaPodUID(meta, hasMeta),
					ContainerName:  ctr.Name,
					ControllerUID:  metricMetaControllerUID(meta, hasMeta),
					ControllerKind: metricMetaControllerKind(meta, hasMeta),
					ControllerName: metricMetaControllerName(meta, hasMeta),
					CollectedAt:    collectedAt,
					Value:          float64(cpuMilli) / 1000.0,
				})
			}

			if mem := ctr.Usage.Memory(); mem != nil {
				memBytes := mem.Value()
				totalMemBytes += memBytes

				out = append(out, model.MetricPointPayload{
					MetricName:     "kube_container_memory_usage_bytes",
					MetricType:     "gauge",
					Unit:           "bytes",
					ResourceKind:   "container",
					NodeName:       metricMetaNodeName(meta, hasMeta),
					Namespace:      pm.Namespace,
					PodName:        pm.Name,
					PodUID:         metricMetaPodUID(meta, hasMeta),
					ContainerName:  ctr.Name,
					ControllerUID:  metricMetaControllerUID(meta, hasMeta),
					ControllerKind: metricMetaControllerKind(meta, hasMeta),
					ControllerName: metricMetaControllerName(meta, hasMeta),
					CollectedAt:    collectedAt,
					Value:          float64(memBytes),
				})
			}
		}

		out = append(out, model.MetricPointPayload{
			MetricName:     "kube_pod_cpu_usage_cores",
			MetricType:     "gauge",
			Unit:           "cores",
			ResourceKind:   "pod",
			NodeName:       metricMetaNodeName(meta, hasMeta),
			Namespace:      pm.Namespace,
			PodName:        pm.Name,
			PodUID:         metricMetaPodUID(meta, hasMeta),
			ControllerUID:  metricMetaControllerUID(meta, hasMeta),
			ControllerKind: metricMetaControllerKind(meta, hasMeta),
			ControllerName: metricMetaControllerName(meta, hasMeta),
			CollectedAt:    collectedAt,
			Value:          float64(totalMilliCPU) / 1000.0,
		})

		out = append(out, model.MetricPointPayload{
			MetricName:     "kube_pod_memory_usage_bytes",
			MetricType:     "gauge",
			Unit:           "bytes",
			ResourceKind:   "pod",
			NodeName:       metricMetaNodeName(meta, hasMeta),
			Namespace:      pm.Namespace,
			PodName:        pm.Name,
			PodUID:         metricMetaPodUID(meta, hasMeta),
			ControllerUID:  metricMetaControllerUID(meta, hasMeta),
			ControllerKind: metricMetaControllerKind(meta, hasMeta),
			ControllerName: metricMetaControllerName(meta, hasMeta),
			CollectedAt:    collectedAt,
			Value:          float64(totalMemBytes),
		})
	}

	return out, nil
}

func collectPodMetricMeta(ctx context.Context, c *Clients) (map[string]podMetricMeta, error) {
	pods, err := c.Core.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	out := make(map[string]podMetricMeta, len(pods.Items))
	for _, p := range pods.Items {
		controllerUID, controllerKind, controllerName := primaryOwnerRef(p.OwnerReferences)

		meta := podMetricMeta{
			UID:            string(p.UID),
			Namespace:      p.Namespace,
			PodName:        p.Name,
			NodeName:       p.Spec.NodeName,
			ControllerUID:  controllerUID,
			ControllerKind: controllerKind,
			ControllerName: controllerName,
		}
		out[podMetricKey(p.Namespace, p.Name)] = meta
	}

	return out, nil
}

func podMetricKey(namespace, podName string) string {
	return namespace + "/" + podName
}

func metricsTimestampOrNow(ts time.Time, fallback time.Time) time.Time {
	if ts.IsZero() {
		return fallback
	}
	return ts.UTC()
}

func metricMetaPodUID(meta podMetricMeta, ok bool) string {
	if !ok {
		return ""
	}
	return meta.UID
}

func metricMetaNodeName(meta podMetricMeta, ok bool) string {
	if !ok {
		return ""
	}
	return meta.NodeName
}

func metricMetaControllerUID(meta podMetricMeta, ok bool) string {
	if !ok {
		return ""
	}
	return meta.ControllerUID
}

func metricMetaControllerKind(meta podMetricMeta, ok bool) string {
	if !ok {
		return ""
	}
	return meta.ControllerKind
}

func metricMetaControllerName(meta podMetricMeta, ok bool) string {
	if !ok {
		return ""
	}
	return meta.ControllerName
}

func resolveTopLevelControllerFromPod(_ corev1.Pod) (uid, kind, name string) {
	return "", "", ""
}
