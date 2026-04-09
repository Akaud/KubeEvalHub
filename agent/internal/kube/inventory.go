package kube

import (
	"context"

	"agent/internal/model"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
		out.Nodes = append(out.Nodes, toNodePayload(node))
	}

	deployments, err := c.Core.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, d := range deployments.Items {
		out.Deployments = append(out.Deployments, toDeploymentPayload(d))
	}

	statefulSets, err := c.Core.AppsV1().StatefulSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, s := range statefulSets.Items {
		out.StatefulSets = append(out.StatefulSets, toStatefulSetPayload(s))
	}

	daemonSets, err := c.Core.AppsV1().DaemonSets("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, d := range daemonSets.Items {
		out.DaemonSets = append(out.DaemonSets, toDaemonSetPayload(d))
	}

	pods, err := c.Core.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return out, err
	}
	for _, p := range pods.Items {
		out.Pods = append(out.Pods, toPodPayload(p))
	}

	return out, nil
}

func toNodePayload(node corev1.Node) model.NodePayload {
	cpuCapacity := quantityMilliValuePtr(node.Status.Capacity[corev1.ResourceCPU])
	memCapacity := quantityValuePtr(node.Status.Capacity[corev1.ResourceMemory])
	podCapacity := quantityValuePtr(node.Status.Capacity[corev1.ResourcePods])

	cpuAllocatable := quantityMilliValuePtr(node.Status.Allocatable[corev1.ResourceCPU])
	memAllocatable := quantityValuePtr(node.Status.Allocatable[corev1.ResourceMemory])
	podAllocatable := quantityValuePtr(node.Status.Allocatable[corev1.ResourcePods])

	return model.NodePayload{
		UID:    string(node.UID),
		Name:   node.Name,
		Labels: copyStringMap(node.Labels),

		KubeletVersion:   node.Status.NodeInfo.KubeletVersion,
		ContainerRuntime: node.Status.NodeInfo.ContainerRuntimeVersion,
		OperatingSystem:  node.Status.NodeInfo.OperatingSystem,
		Architecture:     node.Status.NodeInfo.Architecture,
		KernelVersion:    node.Status.NodeInfo.KernelVersion,
		OSImage:          node.Status.NodeInfo.OSImage,

		CPUCapacityMillicores:    cpuCapacity,
		MemoryCapacityBytes:      memCapacity,
		CPUAllocatableMillicores: cpuAllocatable,
		MemoryAllocatableBytes:   memAllocatable,
		PodCapacity:              podCapacity,
		PodAllocatable:           podAllocatable,
	}
}

func toDeploymentPayload(d appsv1.Deployment) model.DeploymentPayload {
	containers := make([]model.ContainerSpecPayload, 0, len(d.Spec.Template.Spec.Containers)+len(d.Spec.Template.Spec.InitContainers))
	containers = append(containers, toContainerSpecs(d.Spec.Template.Spec.Containers, false)...)
	containers = append(containers, toContainerSpecs(d.Spec.Template.Spec.InitContainers, true)...)

	return model.DeploymentPayload{
		UID:        string(d.UID),
		Name:       d.Name,
		Namespace:  d.Namespace,
		Replicas:   derefInt32(d.Spec.Replicas),
		Labels:     copyStringMap(d.Labels),
		Containers: containers,
	}
}

func toStatefulSetPayload(s appsv1.StatefulSet) model.StatefulSetPayload {
	containers := make([]model.ContainerSpecPayload, 0, len(s.Spec.Template.Spec.Containers)+len(s.Spec.Template.Spec.InitContainers))
	containers = append(containers, toContainerSpecs(s.Spec.Template.Spec.Containers, false)...)
	containers = append(containers, toContainerSpecs(s.Spec.Template.Spec.InitContainers, true)...)

	return model.StatefulSetPayload{
		UID:        string(s.UID),
		Name:       s.Name,
		Namespace:  s.Namespace,
		Replicas:   derefInt32(s.Spec.Replicas),
		Labels:     copyStringMap(s.Labels),
		Containers: containers,
	}
}

func toDaemonSetPayload(d appsv1.DaemonSet) model.DaemonSetPayload {
	containers := make([]model.ContainerSpecPayload, 0, len(d.Spec.Template.Spec.Containers)+len(d.Spec.Template.Spec.InitContainers))
	containers = append(containers, toContainerSpecs(d.Spec.Template.Spec.Containers, false)...)
	containers = append(containers, toContainerSpecs(d.Spec.Template.Spec.InitContainers, true)...)

	return model.DaemonSetPayload{
		UID:        string(d.UID),
		Name:       d.Name,
		Namespace:  d.Namespace,
		Labels:     copyStringMap(d.Labels),
		Containers: containers,
	}
}

func toPodPayload(p corev1.Pod) model.PodPayload {
	controllerUID, controllerKind, controllerName := primaryOwnerRef(p.OwnerReferences)

	containers := make([]model.ContainerSpecPayload, 0, len(p.Spec.Containers)+len(p.Spec.InitContainers))
	containers = append(containers, toContainerSpecs(p.Spec.Containers, false)...)
	containers = append(containers, toContainerSpecs(p.Spec.InitContainers, true)...)

	statuses := make([]model.ContainerStatusPayload, 0, len(p.Status.ContainerStatuses)+len(p.Status.InitContainerStatuses))
	statuses = append(statuses, toContainerStatuses(p.Status.ContainerStatuses, false)...)
	statuses = append(statuses, toContainerStatuses(p.Status.InitContainerStatuses, true)...)

	return model.PodPayload{
		UID:               string(p.UID),
		Name:              p.Name,
		Namespace:         p.Namespace,
		NodeName:          p.Spec.NodeName,
		Phase:             string(p.Status.Phase),
		Labels:            copyStringMap(p.Labels),
		ControllerUID:     controllerUID,
		ControllerKind:    controllerKind,
		ControllerName:    controllerName,
		Containers:        containers,
		ContainerStatuses: statuses,
	}
}

func toContainerSpecs(containers []corev1.Container, isInit bool) []model.ContainerSpecPayload {
	out := make([]model.ContainerSpecPayload, 0, len(containers))

	for _, ctr := range containers {
		item := model.ContainerSpecPayload{
			Name:            ctr.Name,
			Image:           ctr.Image,
			IsInitContainer: isInit,
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

func toContainerStatuses(statuses []corev1.ContainerStatus, isInit bool) []model.ContainerStatusPayload {
	out := make([]model.ContainerStatusPayload, 0, len(statuses))

	for _, st := range statuses {
		item := model.ContainerStatusPayload{
			Name:            st.Name,
			ContainerID:     st.ContainerID,
			RestartCount:    st.RestartCount,
			Ready:           st.Ready,
			Started:         st.Started,
			IsInitContainer: isInit,
		}

		switch {
		case st.State.Running != nil:
			item.State = "running"
		case st.State.Waiting != nil:
			item.State = "waiting"
		case st.State.Terminated != nil:
			item.State = "terminated"
		}

		if st.LastTerminationState.Terminated != nil {
			term := st.LastTerminationState.Terminated
			item.LastTerminationReason = term.Reason

			exitCode := int32(term.ExitCode)
			item.LastTerminationExitCode = &exitCode

			if term.Reason == "OOMKilled" {
				item.OOMKilled = true
			}
		}

		if st.State.Terminated != nil && st.State.Terminated.Reason == "OOMKilled" {
			item.OOMKilled = true
			if item.LastTerminationReason == "" {
				item.LastTerminationReason = st.State.Terminated.Reason
			}
			if item.LastTerminationExitCode == nil {
				exitCode := int32(st.State.Terminated.ExitCode)
				item.LastTerminationExitCode = &exitCode
			}
		}

		out = append(out, item)
	}

	return out
}

func primaryOwnerRef(refs []metav1.OwnerReference) (uid, kind, name string) {
	for _, ref := range refs {
		if ref.Controller != nil && *ref.Controller {
			return string(ref.UID), ref.Kind, ref.Name
		}
	}
	if len(refs) > 0 {
		return string(refs[0].UID), refs[0].Kind, refs[0].Name
	}
	return "", "", ""
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

func quantityValuePtr(q resource.Quantity) *int64 {
	if q.IsZero() {
		return nil
	}
	v := q.Value()
	return &v
}

func quantityMilliValuePtr(q resource.Quantity) *int64 {
	if q.IsZero() {
		return nil
	}
	v := q.MilliValue()
	return &v
}
