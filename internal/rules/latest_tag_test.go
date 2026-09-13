package rules

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func TestLatestTagRule(t *testing.T) {

	w := &workload.Workload{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "api",
					Image: "nginx:latest",
				},
			},
		},
	}

	findings := CheckLatestTag(w)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
