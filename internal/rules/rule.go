package rules

import "github.com/loukasprevyzis/kube-review/internal/workload"

// Rule pairs a stable ID with the check function that implements it. The ID
// is the key Policy uses to enable/disable a rule or override its severity,
// and the key SARIF output uses to identify a rule, so it must stay stable
// across releases (it's user-facing config).
type Rule struct {
	ID          string
	Description string
	check       func(*workload.Workload) []Finding
}

var registry = []Rule{
	{
		ID:          "latest-tag",
		Description: "Container image uses the mutable :latest tag instead of a pinned version.",
		check:       CheckLatestTag,
	},
	{
		ID:          "resource-limits",
		Description: "Container does not define resource limits (cpu/memory).",
		check:       CheckResourceLimits,
	},
	{
		ID:          "resource-requests",
		Description: "Container does not define resource requests (cpu/memory).",
		check:       CheckResourceRequests,
	},
	{
		ID:          "run-as-non-root",
		Description: "Container does not set securityContext.runAsNonRoot=true.",
		check:       CheckRunAsNonRoot,
	},
	{
		ID:          "readiness-probe",
		Description: "Container does not define a readiness probe.",
		check:       CheckReadinessProbe,
	},
	{
		ID:          "liveness-probe",
		Description: "Container does not define a liveness probe.",
		check:       CheckLivenessProbe,
	},
	{
		ID:          "privileged-container",
		Description: "Container runs with securityContext.privileged=true.",
		check:       CheckPrivilegedContainer,
	},
	{
		ID:          "allow-privilege-escalation",
		Description: "Container does not set securityContext.allowPrivilegeEscalation=false.",
		check:       CheckAllowPrivilegeEscalation,
	},
}

// RuleIDs returns every known rule ID, used to validate policy config.
func RuleIDs() []string {
	ids := make([]string, len(registry))
	for i, r := range registry {
		ids[i] = r.ID
	}
	return ids
}

// Registry exposes the rule catalog (ID + description) for output formats,
// such as SARIF, that need to declare rules independently of findings.
func Registry() []Rule {
	return registry
}
