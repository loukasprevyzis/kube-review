package rules

import (
	"fmt"
	"strings"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func CheckLatestTag(w *workload.Workload) []Finding {

	var findings []Finding

	for _, c := range w.AllContainers() {

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
