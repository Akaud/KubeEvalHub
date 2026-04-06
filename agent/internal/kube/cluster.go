package kube

import (
	"context"
	"fmt"
	"strings"

	"agent/internal/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func DiscoverCluster(ctx context.Context, c *Clients) (model.ClusterPayload, error) {
	versionInfo, err := c.Core.Discovery().ServerVersion()
	if err != nil {
		return model.ClusterPayload{}, err
	}

	nodes, err := c.Core.CoreV1().Nodes().List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		return model.ClusterPayload{}, err
	}

	clusterUID := "unknown"
	clusterName := "unknown"

	if len(nodes.Items) > 0 {
		clusterUID = string(nodes.Items[0].UID)
		clusterName = firstNonEmpty(
			nodes.Items[0].Labels["cluster.x-k8s.io/cluster-name"],
			nodes.Items[0].Labels["kubernetes.io/hostname"],
			"self-hosted-cluster",
		)
	}

	return model.ClusterPayload{
		ClusterUID:    clusterUID,
		ClusterName:   clusterName,
		KubeVersion:   versionInfo.GitVersion,
		Distribution:  detectDistribution(versionInfo.GitVersion),
		APIServerHost: "",
	}, nil
}

func detectDistribution(v string) string {
	s := strings.ToLower(v)
	switch {
	case strings.Contains(s, "k3s"):
		return "k3s"
	case strings.Contains(s, "rke2"):
		return "rke2"
	case strings.Contains(s, "microk8s"):
		return "microk8s"
	default:
		return "unknown"
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return fmt.Sprintf("cluster-%s", "unknown")
}
