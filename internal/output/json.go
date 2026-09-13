package output

import (
	"encoding/json"
	"io"

	"github.com/loukasprevyzis/kube-review/internal/rules"
)

// Result is one reviewed unit of work: either a workload with its findings,
// or a file that failed to load (Error set, everything else zero-valued).
type Result struct {
	File     string          `json:"file"`
	Kind     string          `json:"kind,omitempty"`
	Name     string          `json:"name,omitempty"`
	Findings []rules.Finding `json:"findings,omitempty"`
	Error    string          `json:"error,omitempty"`
}

func PrintJSON(w io.Writer, results []Result) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(results)
}
