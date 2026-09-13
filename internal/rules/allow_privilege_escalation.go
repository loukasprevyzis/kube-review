package rules

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckAllowPrivilegeEscalation(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.AllContainers() {

		if c.SecurityContext == nil ||
			c.SecurityContext.AllowPrivilegeEscalation == nil ||
			*c.SecurityContext.AllowPrivilegeEscalation {

			findings = append(findings, Finding{
				Category: "Security",
				Severity: High,
				Message: fmt.Sprintf(
					"%s does not set allowPrivilegeEscalation=false",
					c.Name,
				),
			})
		}
	}

	return findings
}
