package rules

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

func CheckPrivilegedContainer(d *appsv1.Deployment) []Finding {

	var findings []Finding

	for _, c := range d.Spec.Template.Spec.Containers {

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
