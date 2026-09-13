package rules

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckResourceLimits(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.AllContainers() {

		if c.Resources.Limits == nil {

			findings = append(findings, Finding{
				Category: "Cost",
				Severity: "MEDIUM",
				Message: fmt.Sprintf(
					"%s has no resource limits configured",
					c.Name,
				),
			})
		}
	}

	return findings
}
