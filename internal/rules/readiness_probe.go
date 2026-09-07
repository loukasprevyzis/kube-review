package rules

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
)

func CheckReadinessProbe(d *appsv1.Deployment) []Finding {

	var findings []Finding

	for _, c := range d.Spec.Template.Spec.Containers {

		if c.ReadinessProbe == nil {

			findings = append(findings, Finding{
				Category: "Reliability",
				Severity: Medium,
				Message: fmt.Sprintf(
					"%s does not define a readiness probe",
					c.Name,
				),
			})
		}
	}

	return findings
}
