package parser

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestIsHelmChart(t *testing.T) {
	dir := t.TempDir()

	if IsHelmChart(dir) {
		t.Fatal("expected an empty directory not to be a Helm chart")
	}

	if err := os.WriteFile(filepath.Join(dir, "Chart.yaml"), []byte("apiVersion: v2\nname: test\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if !IsHelmChart(dir) {
		t.Fatal("expected a directory with Chart.yaml to be a Helm chart")
	}
}

func TestIsHelmChart_NotADirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "not-a-dir.yaml")

	if err := os.WriteFile(file, []byte("kind: Deployment\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if IsHelmChart(file) {
		t.Fatal("expected a plain file not to be a Helm chart")
	}
}

func TestSourceComment(t *testing.T) {
	doc := []byte("---\n# Source: mychart/templates/deployment.yaml\napiVersion: apps/v1\nkind: Deployment\n")

	if got := sourceComment(doc); got != "mychart/templates/deployment.yaml" {
		t.Fatalf("expected extracted source path, got %q", got)
	}

	if got := sourceComment([]byte("apiVersion: apps/v1\nkind: Deployment\n")); got != "" {
		t.Fatalf("expected empty string when no Source comment present, got %q", got)
	}
}

func TestLoadHelmChart(t *testing.T) {
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm not installed, skipping Helm integration test")
	}

	rendered, err := LoadHelmChart("../../examples/helm-chart", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rendered) != 1 {
		t.Fatalf("expected 1 rendered workload, got %d", len(rendered))
	}

	rw := rendered[0]

	if rw.Workload.Kind != "Deployment" || rw.Workload.Name != "payments-api" {
		t.Fatalf("unexpected workload: %+v", rw.Workload)
	}

	if len(rw.Workload.Spec.Containers) != 1 || rw.Workload.Spec.Containers[0].Image != "ghcr.io/acme/payments-api:latest" {
		t.Fatalf("unexpected container image, got: %+v", rw.Workload.Spec.Containers)
	}

	if rw.Source == "" {
		t.Fatal("expected a non-empty source template path")
	}
}

func TestLoadHelmChart_WithValuesOverride(t *testing.T) {
	if _, err := exec.LookPath("helm"); err != nil {
		t.Skip("helm not installed, skipping Helm integration test")
	}

	rendered, err := LoadHelmChart("../../examples/helm-chart", []string{"../../examples/helm-chart/values-prod.yaml"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(rendered) != 1 {
		t.Fatalf("expected 1 rendered workload, got %d", len(rendered))
	}

	image := rendered[0].Workload.Spec.Containers[0].Image
	if image != "ghcr.io/acme/payments-api:1.4.2" {
		t.Fatalf("expected values override to change the image tag, got %q", image)
	}
}

func TestLoadHelmChart_HelmNotOnPath(t *testing.T) {
	t.Setenv("PATH", "")

	if _, err := LoadHelmChart("../../examples/helm-chart", nil); err == nil {
		t.Fatal("expected an error when helm is not on PATH")
	}
}
