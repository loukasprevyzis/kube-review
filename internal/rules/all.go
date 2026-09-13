package rules

import "github.com/loukasprevyzis/kube-review/internal/workload"

func RunAll(w *workload.Workload, policy Policy) []Finding {

	var findings []Finding

	for _, r := range registry {

		if !policy.enabled(r.ID) {
			continue
		}

		results := r.check(w)

		for i := range results {
			results[i].RuleID = r.ID
			if severity := policy.severityOverride(r.ID); severity != "" {
				results[i].Severity = severity
			}
		}

		findings = append(findings, results...)
	}

	return findings
}
