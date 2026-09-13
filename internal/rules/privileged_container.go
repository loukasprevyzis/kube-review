package rules

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckPrivilegedContainer(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.AllContainers() {

		if c.SecurityContext != nil &&
			c.SecurityContext.Privileged != nil &&
			*c.SecurityContext.Privileged {

			findings = append(findings, Finding{
				Category: "Security",
				Severity: High,
				Message: fmt.Sprintf(
					"%s runs as a privileged container",
					c.Name,
				),
			})
		}
	}

	return findings
}
