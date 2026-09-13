package rules

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckLivenessProbe(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.Spec.Containers {

		if c.LivenessProbe == nil {

			findings = append(findings, Finding{
				Category: "Reliability",
				Severity: Medium,
				Message: fmt.Sprintf(
					"%s does not define a liveness probe",
					c.Name,
				),
			})
		}
	}

	return findings
}
