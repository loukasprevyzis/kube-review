package rules

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckRunAsNonRoot(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.AllContainers() {

		if c.SecurityContext == nil ||
			c.SecurityContext.RunAsNonRoot == nil ||
			!*c.SecurityContext.RunAsNonRoot {

			findings = append(findings, Finding{
				Category: "Security",
				Severity: "HIGH",
				Message: fmt.Sprintf(
					"%s does not set runAsNonRoot=true",
					c.Name,
				),
			})
		}
	}

	return findings
}
