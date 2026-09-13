package output

import (
	"encoding/json"
	"io"

	"github.com/loukasprevyzis/kube-review/internal/rules"
)

// SARIF (Static Analysis Results Interchange Format) 2.1.0 is what GitHub
// code scanning, and most other CI security dashboards, expect. Structs
// here cover only the subset of the spec kube-review actually populates.
type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string                 `json:"id"`
	ShortDescription sarifText              `json:"shortDescription"`
	DefaultConfig    sarifRuleDefaultConfig `json:"defaultConfiguration"`
}

type sarifRuleDefaultConfig struct {
	Level string `json:"level"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifText       `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

func sarifLevel(severity string) string {
	switch severity {
	case rules.High:
		return "error"
	case rules.Medium:
		return "warning"
	case rules.Low:
		return "note"
	default:
		return "warning"
	}
}

// PrintSARIF writes results as a SARIF 2.1.0 log. Results whose Error field
// is set (files that failed to parse) are skipped, since SARIF results
// describe findings in an artifact, not load failures.
func PrintSARIF(w io.Writer, results []Result) error {

	log := sarifLog{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:           "kube-review",
						InformationURI: "https://github.com/loukasprevyzis/kube-review",
						Rules:          sarifRules(),
					},
				},
				Results: []sarifResult{},
			},
		},
	}

	for _, r := range results {
		if r.Error != "" {
			continue
		}

		for _, f := range r.Findings {
			log.Runs[0].Results = append(log.Runs[0].Results, sarifResult{
				RuleID:  f.RuleID,
				Level:   sarifLevel(f.Severity),
				Message: sarifText{Text: f.Message},
				Locations: []sarifLocation{
					{
						PhysicalLocation: sarifPhysicalLocation{
							ArtifactLocation: sarifArtifactLocation{URI: r.File},
						},
					},
				},
			})
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(log)
}

func sarifRules() []sarifRule {
	registry := rules.Registry()
	out := make([]sarifRule, len(registry))

	for i, r := range registry {
		out[i] = sarifRule{
			ID:               r.ID,
			ShortDescription: sarifText{Text: r.Description},
			DefaultConfig:    sarifRuleDefaultConfig{Level: "warning"},
		}
	}

	return out
}
