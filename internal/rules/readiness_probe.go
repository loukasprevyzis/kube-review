package rules

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckReadinessProbe(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.Spec.Containers {

		if c.ReadinessProbe == nil {

			findings = append(findings, Finding{
				Category: "Reliability",
				Severity: Medium,
				Message: fmt.Sprintf(
					"%s does not define a readiness probe",
					c.Name,
				),
			})
		}
	}

	return findings
}
