package runner

import (
	"sort"

	"agent/internal/model"
)

func sortInventoryPayload(in *model.InventoryPayload) {
	sort.Slice(in.Namespaces, func(i, j int) bool {
		return in.Namespaces[i].UID < in.Namespaces[j].UID
	})

	sort.Slice(in.Nodes, func(i, j int) bool {
		return in.Nodes[i].UID < in.Nodes[j].UID
	})

	sort.Slice(in.Deployments, func(i, j int) bool {
		if in.Deployments[i].Namespace == in.Deployments[j].Namespace {
			return in.Deployments[i].UID < in.Deployments[j].UID
		}
		return in.Deployments[i].Namespace < in.Deployments[j].Namespace
	})

	sort.Slice(in.StatefulSets, func(i, j int) bool {
		if in.StatefulSets[i].Namespace == in.StatefulSets[j].Namespace {
			return in.StatefulSets[i].UID < in.StatefulSets[j].UID
		}
		return in.StatefulSets[i].Namespace < in.StatefulSets[j].Namespace
	})

	sort.Slice(in.DaemonSets, func(i, j int) bool {
		if in.DaemonSets[i].Namespace == in.DaemonSets[j].Namespace {
			return in.DaemonSets[i].UID < in.DaemonSets[j].UID
		}
		return in.DaemonSets[i].Namespace < in.DaemonSets[j].Namespace
	})

	sort.Slice(in.Pods, func(i, j int) bool {
		if in.Pods[i].Namespace == in.Pods[j].Namespace {
			return in.Pods[i].UID < in.Pods[j].UID
		}
		return in.Pods[i].Namespace < in.Pods[j].Namespace
	})

	for i := range in.Deployments {
		sortContainerSpecs(in.Deployments[i].Containers)
	}
	for i := range in.StatefulSets {
		sortContainerSpecs(in.StatefulSets[i].Containers)
	}
	for i := range in.DaemonSets {
		sortContainerSpecs(in.DaemonSets[i].Containers)
	}
	for i := range in.Pods {
		sortContainerSpecs(in.Pods[i].Containers)
		sortContainerStatuses(in.Pods[i].ContainerStatuses)
	}
}

func sortContainerSpecs(items []model.ContainerSpecPayload) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsInitContainer == items[j].IsInitContainer {
			return items[i].Name < items[j].Name
		}
		return !items[i].IsInitContainer && items[j].IsInitContainer
	})
}

func sortContainerStatuses(items []model.ContainerStatusPayload) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsInitContainer == items[j].IsInitContainer {
			return items[i].Name < items[j].Name
		}
		return !items[i].IsInitContainer && items[j].IsInitContainer
	})
}
