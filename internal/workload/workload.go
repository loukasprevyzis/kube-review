// Package workload provides a resource-kind-agnostic view over anything
// that carries a pod spec (Deployment, StatefulSet, DaemonSet, Job, CronJob,
// bare Pod, ...) so rules only need to be written once.
package workload

import corev1 "k8s.io/api/core/v1"

type Workload struct {
	Kind string
	Name string
	Spec corev1.PodSpec
}

// AllContainers returns init containers and regular containers together,
// for checks (security, image, resources) that apply to both.
func (w *Workload) AllContainers() []corev1.Container {
	all := make([]corev1.Container, 0, len(w.Spec.InitContainers)+len(w.Spec.Containers))
	all = append(all, w.Spec.InitContainers...)
	all = append(all, w.Spec.Containers...)
	return all
}
