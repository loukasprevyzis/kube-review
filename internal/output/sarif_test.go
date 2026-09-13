package output

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/loukasprevyzis/kube-review/internal/rules"
)

func TestPrintSARIF(t *testing.T) {
	results := []Result{
		{
			File: "deployment.yaml",
			Kind: "Deployment",
			Name: "payments-api",
			Findings: []rules.Finding{
				{RuleID: "latest-tag", Category: "Security", Severity: rules.High, Message: "uses latest tag"},
			},
		},
		{
			File:  "broken.yaml",
			Error: "yaml: bad indentation",
		},
	}

	var buf bytes.Buffer
	if err := PrintSARIF(&buf, results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var log sarifLog
	if err := json.Unmarshal(buf.Bytes(), &log); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	if log.Version != "2.1.0" {
		t.Fatalf("expected SARIF version 2.1.0, got %q", log.Version)
	}

	if len(log.Runs) != 1 {
		t.Fatalf("expected exactly one run, got %d", len(log.Runs))
	}

	run := log.Runs[0]

	if len(run.Tool.Driver.Rules) != len(rules.Registry()) {
		t.Fatalf("expected %d declared rules, got %d", len(rules.Registry()), len(run.Tool.Driver.Rules))
	}

	if len(run.Results) != 1 {
		t.Fatalf("expected 1 result (the load error must be skipped), got %d", len(run.Results))
	}

	result := run.Results[0]

	if result.RuleID != "latest-tag" {
		t.Fatalf("expected ruleId %q, got %q", "latest-tag", result.RuleID)
	}
	if result.Level != "error" {
		t.Fatalf("expected level %q for HIGH severity, got %q", "error", result.Level)
	}
	if len(result.Locations) != 1 || result.Locations[0].PhysicalLocation.ArtifactLocation.URI != "deployment.yaml" {
		t.Fatalf("expected a location pointing at deployment.yaml, got %+v", result.Locations)
	}
}

func TestSarifLevel(t *testing.T) {
	cases := map[string]string{
		rules.High:   "error",
		rules.Medium: "warning",
		rules.Low:    "note",
	}

	for severity, want := range cases {
		if got := sarifLevel(severity); got != want {
			t.Errorf("sarifLevel(%q) = %q, want %q", severity, got, want)
		}
	}
}
