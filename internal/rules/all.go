package rules

import appsv1 "k8s.io/api/apps/v1"

func RunAll(d *appsv1.Deployment) []Finding {

	var findings []Finding

	findings = append(findings, CheckLatestTag(d)...)
	findings = append(findings, CheckResourceLimits(d)...)
	findings = append(findings, CheckResourceRequests(d)...)
	findings = append(findings, CheckRunAsNonRoot(d)...)
	findings = append(findings, CheckReadinessProbe(d)...)
	findings = append(findings, CheckLivenessProbe(d)...)
	findings = append(findings, CheckPrivilegedContainer(d)...)
	findings = append(findings, CheckAllowPrivilegeEscalation(d)...)

	return findings
}
