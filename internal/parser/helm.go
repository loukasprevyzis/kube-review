package parser

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

// RenderedWorkload pairs a workload rendered from a Helm chart with the
// template file it came from, so review output can point at the source
// template instead of just the chart directory.
type RenderedWorkload struct {
	Source   string
	Workload *workload.Workload
}

// IsHelmChart reports whether dir is the root of a Helm chart, i.e. it
// contains a Chart.yaml.
func IsHelmChart(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}

	_, err = os.Stat(filepath.Join(dir, "Chart.yaml"))
	return err == nil
}

// LoadHelmChart renders a chart with `helm template` and parses the
// resulting manifests the same way a plain YAML file would be. valuesFiles
// are passed through as --values flags, in the order given.
func LoadHelmChart(dir string, valuesFiles []string) ([]RenderedWorkload, error) {
	if _, err := exec.LookPath("helm"); err != nil {
		return nil, fmt.Errorf(
			"helm not found on PATH (required to render chart %s): install it from https://helm.sh/docs/intro/install/",
			dir,
		)
	}

	args := []string{"template", dir}
	for _, f := range valuesFiles {
		args = append(args, "--values", f)
	}

	cmd := exec.Command("helm", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("helm template %s: %w: %s", dir, err, strings.TrimSpace(stderr.String()))
	}

	docs, err := splitYAMLDocuments(stdout.Bytes())
	if err != nil {
		return nil, fmt.Errorf("parsing helm template output for %s: %w", dir, err)
	}

	var rendered []RenderedWorkload

	for _, doc := range docs {
		w, err := parseWorkload(doc)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", dir, err)
		}
		if w == nil {
			continue
		}

		source := sourceComment(doc)
		if source == "" {
			source = dir
		}

		rendered = append(rendered, RenderedWorkload{Source: source, Workload: w})
	}

	return rendered, nil
}

// sourceComment extracts the template path from Helm's "# Source: <path>"
// comment, which it prepends to every document it renders.
func sourceComment(doc []byte) string {
	for _, line := range strings.Split(string(doc), "\n") {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "# Source:"); ok {
			return strings.TrimSpace(after)
		}
	}
	return ""
}
