package rules

import appsv1 "k8s.io/api/apps/v1"

type Rule interface {
	Name() string
	Check(*appsv1.Deployment) []Finding
}
