package rules

import (
	"fmt"
	"os"
	"strings"

	"sigs.k8s.io/yaml"
)

// RuleConfig is one rule's entry in a policy file.
type RuleConfig struct {
	Enabled  *bool  `json:"enabled,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// Policy is a set of per-rule overrides, keyed by rule ID. The zero value
// runs every rule at its default severity.
type Policy struct {
	Rules map[string]RuleConfig `json:"rules,omitempty"`
}

func (p Policy) enabled(id string) bool {
	rc, ok := p.Rules[id]
	if !ok || rc.Enabled == nil {
		return true
	}
	return *rc.Enabled
}

func (p Policy) severityOverride(id string) string {
	return p.Rules[id].Severity
}

// LoadPolicy reads and validates a policy file. Unknown rule IDs and invalid
// severities are rejected outright rather than ignored, since a silently
// mistyped rule ID would otherwise disable a security check without warning.
func LoadPolicy(path string) (Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Policy{}, err
	}

	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Policy{}, fmt.Errorf("parsing %s: %w", path, err)
	}

	known := make(map[string]bool, len(registry))
	for _, id := range RuleIDs() {
		known[id] = true
	}

	for id, rc := range p.Rules {
		if !known[id] {
			return Policy{}, fmt.Errorf("%s: unknown rule %q", path, id)
		}

		if rc.Severity == "" {
			continue
		}

		severity := strings.ToUpper(rc.Severity)
		if severity != High && severity != Medium && severity != Low {
			return Policy{}, fmt.Errorf(
				"%s: rule %q has invalid severity %q (expected HIGH, MEDIUM, or LOW)",
				path, id, rc.Severity,
			)
		}

		rc.Severity = severity
		p.Rules[id] = rc
	}

	return p, nil
}
