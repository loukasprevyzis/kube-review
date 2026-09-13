package rules

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func TestResourceRequestsRule(t *testing.T) {

	w := &workload.Workload{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name: "api",
				},
			},
		},
	}

	findings := CheckResourceRequests(w)

	if len(findings) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findings))
	}
}
