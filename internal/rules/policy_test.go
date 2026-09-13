package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

func vulnerableWorkload() *workload.Workload {
	return &workload.Workload{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "api",
					Image: "nginx:latest",
				},
			},
		},
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
}

// vulnerableWorkload triggers every rule except privileged-container, since
// that one only fires when securityContext.privileged is explicitly true.
var vulnerableWorkloadFindingCount = len(registry) - 1

func TestRunAll_DefaultPolicyRunsEveryRule(t *testing.T) {
	findings := RunAll(vulnerableWorkload(), Policy{})

	if len(findings) != vulnerableWorkloadFindingCount {
		t.Fatalf("expected %d findings, got %d", vulnerableWorkloadFindingCount, len(findings))
	}
}

func TestRunAll_DisabledRuleIsSkipped(t *testing.T) {
	disabled := false

	policy := Policy{Rules: map[string]RuleConfig{
		"latest-tag": {Enabled: &disabled},
	}}

	findings := RunAll(vulnerableWorkload(), policy)

	for _, f := range findings {
		if strings.Contains(f.Message, "latest tag") {
			t.Fatalf("expected latest-tag finding to be suppressed, got: %v", f)
		}
	}

	want := vulnerableWorkloadFindingCount - 1
	if len(findings) != want {
		t.Fatalf("expected %d findings with latest-tag disabled, got %d", want, len(findings))
	}
}

func TestRunAll_SeverityOverride(t *testing.T) {
	policy := Policy{Rules: map[string]RuleConfig{
		"latest-tag": {Severity: Low},
	}}

	findings := RunAll(vulnerableWorkload(), policy)

	found := false
	for _, f := range findings {
		if strings.Contains(f.Message, "latest tag") {
			found = true
			if f.Severity != Low {
				t.Fatalf("expected overridden severity %q, got %q", Low, f.Severity)
			}
		}
	}
	if !found {
		t.Fatal("expected a latest-tag finding")
	}
}

func TestLoadPolicy_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yml")

	writeFile(t, path, `
rules:
  latest-tag:
    enabled: false
  privileged-container:
    severity: medium
`)

	policy, err := LoadPolicy(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if policy.enabled("latest-tag") {
		t.Fatal("expected latest-tag to be disabled")
	}

	if got := policy.severityOverride("privileged-container"); got != Medium {
		t.Fatalf("expected severity override %q, got %q", Medium, got)
	}
}

func TestLoadPolicy_UnknownRule(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yml")

	writeFile(t, path, `
rules:
  totally-not-a-rule:
    enabled: false
`)

	if _, err := LoadPolicy(path); err == nil {
		t.Fatal("expected an error for an unknown rule ID")
	}
}

func TestLoadPolicy_InvalidSeverity(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "policy.yml")

	writeFile(t, path, `
rules:
  latest-tag:
    severity: EXTREME
`)

	if _, err := LoadPolicy(path); err == nil {
		t.Fatal("expected an error for an invalid severity")
	}
}
