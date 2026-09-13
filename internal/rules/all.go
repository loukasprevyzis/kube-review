package rules

import "github.com/loukasprevyzis/kube-review/internal/workload"

func RunAll(w *workload.Workload) []Finding {

	var findings []Finding

	findings = append(findings, CheckLatestTag(w)...)
	findings = append(findings, CheckResourceLimits(w)...)
	findings = append(findings, CheckResourceRequests(w)...)
	findings = append(findings, CheckRunAsNonRoot(w)...)
	findings = append(findings, CheckReadinessProbe(w)...)
	findings = append(findings, CheckLivenessProbe(w)...)
	findings = append(findings, CheckPrivilegedContainer(w)...)
	findings = append(findings, CheckAllowPrivilegeEscalation(w)...)

	return findings
}
