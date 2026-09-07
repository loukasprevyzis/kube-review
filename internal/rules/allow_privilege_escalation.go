package rules

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

func CheckAllowPrivilegeEscalation(d *appsv1.Deployment) []Finding {

	var findings []Finding

	for _, c := range d.Spec.Template.Spec.Containers {

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
