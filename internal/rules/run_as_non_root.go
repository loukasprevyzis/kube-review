package rules

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

func CheckRunAsNonRoot(d *appsv1.Deployment) []Finding {

	var findings []Finding

	for _, c := range d.Spec.Template.Spec.Containers {

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
