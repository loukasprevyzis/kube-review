package parser

import (
	"os"
	"path/filepath"
	"testing"
)

// A directory scan must not treat a nested Helm chart's raw templates as
// standalone YAML: templates use Go template syntax ({{ .Values.x }}) that
// isn't valid YAML on its own and would fail to parse.
func TestListYAMLFiles_SkipsNestedHelmChart(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, filepath.Join(dir, "plain.yaml"), "kind: Deployment\n")

	chartDir := filepath.Join(dir, "chart")
	mkdirTestDir(t, filepath.Join(chartDir, "templates"))
	writeTestFile(t, filepath.Join(chartDir, "Chart.yaml"), "apiVersion: v2\nname: test\n")
	writeTestFile(t, filepath.Join(chartDir, "templates", "deployment.yaml"), "name: {{ .Values.x }}\n")

	files, err := ListYAMLFiles(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("expected only the plain YAML file, got %v", files)
	}
}

func TestListHelmCharts(t *testing.T) {
	dir := t.TempDir()

	writeTestFile(t, filepath.Join(dir, "plain.yaml"), "kind: Deployment\n")

	chartDir := filepath.Join(dir, "chart")
	mkdirTestDir(t, chartDir)
	writeTestFile(t, filepath.Join(chartDir, "Chart.yaml"), "apiVersion: v2\nname: test\n")

	charts, err := ListHelmCharts(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(charts) != 1 || charts[0] != chartDir {
		t.Fatalf("expected to find %q, got %v", chartDir, charts)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mkdirTestDir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}
