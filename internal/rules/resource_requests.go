package rules

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

func CheckResourceRequests(d *appsv1.Deployment) []Finding {

	var findings []Finding

	for _, c := range d.Spec.Template.Spec.Containers {

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
