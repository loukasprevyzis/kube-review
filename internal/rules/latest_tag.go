package rules

import (
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
)

func CheckLatestTag(d *appsv1.Deployment) []Finding {

	var findings []Finding

	for _, c := range d.Spec.Template.Spec.Containers {

		if strings.HasSuffix(c.Image, ":latest") {
			findings = append(findings, Finding{
				Category: "Security",
				Severity: "HIGH",
				Message: fmt.Sprintf(
					"%s uses latest tag (%s)",
					c.Name,
					c.Image,
				),
			})
		}
	}

	return findings
}
