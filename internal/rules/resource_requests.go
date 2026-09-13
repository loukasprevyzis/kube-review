package rules

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckResourceRequests(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.AllContainers() {

		if c.Resources.Requests == nil {

			findings = append(findings, Finding{
				Category: "Cost",
				Severity: "MEDIUM",
				Message: fmt.Sprintf(
					"%s has no resource requests configured",
					c.Name,
				),
			})
		}
	}

	return findings
}
