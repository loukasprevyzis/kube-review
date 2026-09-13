package parser

import (
	"path/filepath"
	"testing"
)

func TestLoadWorkloads_SkipsNonKubernetesObjects(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "docker-compose.yml")

	writeTestFile(t, path, `
version: "3"
services:
  web:
    image: nginx
`)

	workloads, err := LoadWorkloads(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workloads) != 0 {
		t.Fatalf("expected no workloads for a non-Kubernetes object, got %d", len(workloads))
	}
}

func TestLoadWorkloads_SkipsYAMLList(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "list.yaml")

	writeTestFile(t, path, "- foo\n- bar\n")

	workloads, err := LoadWorkloads(path)
	if err != nil {
		t.Fatalf("expected a bare YAML list to be skipped, not errored: %v", err)
	}
	if len(workloads) != 0 {
		t.Fatalf("expected no workloads for a YAML list, got %d", len(workloads))
	}
}

func TestLoadWorkloads_SkipsYAMLScalar(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scalar.yaml")

	writeTestFile(t, path, "just a string\n")

	workloads, err := LoadWorkloads(path)
	if err != nil {
		t.Fatalf("expected a bare YAML scalar to be skipped, not errored: %v", err)
	}
	if len(workloads) != 0 {
		t.Fatalf("expected no workloads for a YAML scalar, got %d", len(workloads))
	}
}

func TestLoadWorkloads_ReportsGenuineSyntaxErrors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "malformed.yaml")

	writeTestFile(t, path, "this: is: not: valid: yaml: [\n")

	if _, err := LoadWorkloads(path); err == nil {
		t.Fatal("expected an error for genuinely malformed YAML")
	}
}

func TestLoadWorkloads_ParsesAKnownKind(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deployment.yaml")

	writeTestFile(t, path, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api
spec:
  template:
    spec:
      containers:
      - name: api
        image: nginx:1.4.2
`)

	workloads, err := LoadWorkloads(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(workloads) != 1 || workloads[0].Name != "api" {
		t.Fatalf("expected one workload named api, got %+v", workloads)
	}
}
