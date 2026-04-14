package kube

import (
	"context"
	"errors"
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

	var clusterUID string
	var nodeLabels map[string]string

	if ns, err := c.Core.CoreV1().Namespaces().Get(ctx, "kube-system", metav1.GetOptions{}); err == nil {
		clusterUID = string(ns.UID)
	}

	if len(nodes.Items) > 0 {
		node := nodes.Items[0]
		nodeLabels = node.Labels
	}

	if strings.TrimSpace(clusterUID) == "" {
		return model.ClusterPayload{}, errors.New("failed to discover stable cluster uid")
	}

	return model.ClusterPayload{
		ClusterUID:    clusterUID,
		KubeVersion:   versionInfo.GitVersion,
		Distribution:  detectDistribution(versionInfo.GitVersion, versionInfo.Platform, nodeLabels),
		APIServerHost: "",
	}, nil
}

func detectDistribution(gitVersion, platform string, labels map[string]string) string {
	v := strings.ToLower(gitVersion)
	p := strings.ToLower(platform)

	containsAny := func(s string, parts ...string) bool {
		for _, part := range parts {
			if strings.Contains(s, part) {
				return true
			}
		}
		return false
	}

	hasAnyLabel := func(keys ...string) bool {
		if labels == nil {
			return false
		}
		for _, k := range keys {
			if _, ok := labels[k]; ok {
				return true
			}
		}
		return false
	}

	switch {
	case containsAny(v, "k3s") || containsAny(p, "k3s"):
		return "k3s"
	case containsAny(v, "rke2") || containsAny(p, "rke2"):
		return "rke2"
	case containsAny(v, "microk8s") || containsAny(p, "microk8s"):
		return "microk8s"
	case containsAny(v, "openshift") || containsAny(p, "openshift"):
		return "openshift"

	case hasAnyLabel(
		"eks.amazonaws.com/nodegroup",
		"alpha.eksctl.io/nodegroup-name",
		"eks.amazonaws.com/capacityType",
	):
		return "eks"

	case hasAnyLabel(
		"cloud.google.com/gke-nodepool",
		"cloud.google.com/gke-boot-disk",
		"iam.gke.io/gke-metadata-server-enabled",
	):
		return "gke"

	case hasAnyLabel(
		"kubernetes.azure.com/cluster",
		"kubernetes.azure.com/agentpool",
		"topology.disk.csi.azure.com/zone",
	):
		return "aks"

	case hasAnyLabel(
		"node.openshift.io/os_id",
		"machine.openshift.io/cluster-api-machine-role",
	):
		return "openshift"

	case hasAnyLabel("k3s.io/hostname"):
		return "k3s"

	case hasAnyLabel("rke.cattle.io/machine", "rke2.io/machine"):
		return "rke2"

	default:
		return "unknown"
	}
}
