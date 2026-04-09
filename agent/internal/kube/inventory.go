package kube

import (
	"context"

	"agent/internal/model"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func CollectInventory(ctx context.Context, c *Clients) (model.InventoryPayload, error) {
	var out model.InventoryPayload

	namespaces, err := c.Core.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, ns := range namespaces.Items {
		out.Namespaces = append(out.Namespaces, model.NamespacePayload{
			UID:    string(ns.UID),
			Name:   ns.Name,
			Labels: copyStringMap(ns.Labels),
		})
	}

	nodes, err := c.Core.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, node := range nodes.Items {
		out.Nodes = append(out.Nodes, model.NodePayload{
			UID:              string(node.UID),
			Name:             node.Name,
			Labels:           copyStringMap(node.Labels),
			KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
			ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
			OperatingSystem:  node.Status.NodeInfo.OperatingSystem,
			Architecture:     node.Status.NodeInfo.Architecture,
			KernelVersion:    node.Status.NodeInfo.KernelVersion,
			OSImage:          node.Status.NodeInfo.OSImage,
		})
	}

	deployments, err := c.Core.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, d := range deployments.Items {
		out.Deployments = append(out.Deployments, model.DeploymentPayload{
			UID:        string(d.UID),
			Name:       d.Name,
			Namespace:  d.Namespace,
			Replicas:   derefInt32(d.Spec.Replicas),
			Labels:     copyStringMap(d.Labels),
			Containers: toContainerSpecs(d.Spec.Template.Spec.Containers),
		})
	}

	statefulSets, err := c.Core.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, s := range statefulSets.Items {
		out.StatefulSets = append(out.StatefulSets, model.StatefulSetPayload{
			UID:        string(s.UID),
			Name:       s.Name,
			Namespace:  s.Namespace,
			Replicas:   derefInt32(s.Spec.Replicas),
			Labels:     copyStringMap(s.Labels),
			Containers: toContainerSpecs(s.Spec.Template.Spec.Containers),
		})
	}

	daemonSets, err := c.Core.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, d := range daemonSets.Items {
		out.DaemonSets = append(out.DaemonSets, model.DaemonSetPayload{
			UID:        string(d.UID),
			Name:       d.Name,
			Namespace:  d.Namespace,
			Labels:     copyStringMap(d.Labels),
			Containers: toContainerSpecs(d.Spec.Template.Spec.Containers),
		})
	}

	pods, err := c.Core.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, p := range pods.Items {
		ownerKind, ownerName := primaryOwnerRef(p.OwnerReferences)

		out.Pods = append(out.Pods, model.PodPayload{
			UID:        string(p.UID),
			Name:       p.Name,
			Namespace:  p.Namespace,
			NodeName:   p.Spec.NodeName,
			Phase:      string(p.Status.Phase),
			Labels:     copyStringMap(p.Labels),
			OwnerKind:  ownerKind,
			OwnerName:  ownerName,
			Containers: toContainerSpecs(p.Spec.Containers),
		})
	}

	return out, nil
}

func toContainerSpecs(containers []corev1.Container) []model.ContainerSpecPayload {
	out := make([]model.ContainerSpecPayload, 0, len(containers))

	for _, ctr := range containers {
		item := model.ContainerSpecPayload{
			Name:  ctr.Name,
			Image: ctr.Image,
		}

		if cpuReq := ctr.Resources.Requests.Cpu(); cpuReq != nil && !cpuReq.IsZero() {
			v := cpuReq.MilliValue()
			item.CPURequestMillicores = &v
		}
		if cpuLim := ctr.Resources.Limits.Cpu(); cpuLim != nil && !cpuLim.IsZero() {
			v := cpuLim.MilliValue()
			item.CPULimitMillicores = &v
		}
		if memReq := ctr.Resources.Requests.Memory(); memReq != nil && !memReq.IsZero() {
			v := memReq.Value()
			item.MemoryRequestBytes = &v
		}
		if memLim := ctr.Resources.Limits.Memory(); memLim != nil && !memLim.IsZero() {
			v := memLim.Value()
			item.MemoryLimitBytes = &v
		}

		out = append(out, item)
	}

	return out
}

func primaryOwnerRef(refs []metav1.OwnerReference) (string, string) {
	for _, ref := range refs {
		if ref.Controller != nil && *ref.Controller {
			return ref.Kind, ref.Name
		}
	}
	if len(refs) > 0 {
		return refs[0].Kind, refs[0].Name
	}
	return "", ""
}

func derefInt32(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
