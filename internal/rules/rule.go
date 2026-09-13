package rules

import "github.com/loukasprevyzis/kube-review/internal/workload"

// Rule pairs a stable ID with the check function that implements it. The ID
// is the key Policy uses to enable/disable a rule or override its severity,
// so it must stay stable across releases (it's user-facing config).
type Rule struct {
	ID    string
	check func(*workload.Workload) []Finding
}

var registry = []Rule{
	{ID: "latest-tag", check: CheckLatestTag},
	{ID: "resource-limits", check: CheckResourceLimits},
	{ID: "resource-requests", check: CheckResourceRequests},
	{ID: "run-as-non-root", check: CheckRunAsNonRoot},
	{ID: "readiness-probe", check: CheckReadinessProbe},
	{ID: "liveness-probe", check: CheckLivenessProbe},
	{ID: "privileged-container", check: CheckPrivilegedContainer},
	{ID: "allow-privilege-escalation", check: CheckAllowPrivilegeEscalation},
}

// RuleIDs returns every known rule ID, used to validate policy config.
func RuleIDs() []string {
	ids := make([]string, len(registry))
	for i, r := range registry {
		ids[i] = r.ID
	}
	return ids
}
