# kube-review

Static Kubernetes manifest review tool written in Go.

`kube-review` scans Kubernetes manifests and highlights common security, reliability, and cost issues before they reach production.

---

## Features

### Security

- Detects use of `:latest` image tags
- Detects missing `runAsNonRoot: true`
- Detects missing `allowPrivilegeEscalation: false`

### Reliability

- Detects missing readiness probes
- Detects missing liveness probes

### Cost

- Detects missing resource requests
- Detects missing resource limits

### Repository Scanning

- Review a single manifest
- Review an entire directory
- Recursively scans subdirectories
- Supports `.yaml` and `.yml` files
- Supports multi-document YAML files (`---`-separated)
- Renders and reviews Helm charts (requires `helm` on `PATH`)

### Workload Support

- Deployment, StatefulSet, DaemonSet, ReplicaSet, Job, CronJob, and bare Pod
- Non-workload kinds (Service, ConfigMap, ...) are skipped automatically
- Security and cost rules also check init containers

---

## Installation

### Run from source

```bash
go run . review examples
```

### Build binary

```bash
go build -o kube-review
```

Run:

```bash
./kube-review review examples
```

### Install globally

```bash
go install .
```

Verify:

```bash
kube-review review examples
```

If the command is not found:

```bash
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc
source ~/.zshrc
```

---

## Usage

### Review a single manifest

```bash
kube-review review examples/deployment.yaml
```

### Review a directory

```bash
kube-review review examples
```

The tool will recursively discover Kubernetes manifests and review each file. A Helm chart found anywhere in the tree (any directory with a `Chart.yaml`) is rendered with `helm template` instead of being scanned as raw YAML — findings are attributed to the template file they came from.

### Review a Helm chart

```bash
kube-review review examples/helm-chart
```

Pass one or more values files, applied in order, the same way `helm template --values` would:

```bash
kube-review review examples/helm-chart --values examples/helm-chart/values-prod.yaml
```

Requires `helm` on `PATH`. `--values` is rejected if `<path>` isn't a chart (no `Chart.yaml`).

### Flags

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--fail-on` | `HIGH`, `MEDIUM`, `LOW`, `NONE` | `HIGH` | Minimum severity that causes a non-zero exit code. `NONE` never fails on findings (a file that fails to parse still exits non-zero). |
| `--output` | `text`, `json`, `sarif` | `text` | Output format. `json` is for CI/tooling integration; `sarif` produces a SARIF 2.1.0 report for GitHub code scanning and similar dashboards. |
| `--config` | path to a policy file | `.kube-review.yml` if present in the working directory | See [Policy Configuration](#policy-configuration). |
| `--values` | path to a Helm values file (repeatable) | none | Only valid when `<path>` is a Helm chart. |

Flags may appear before or after the path:

```bash
kube-review review examples --fail-on MEDIUM --output json
```

### Exit codes

- `0` — no findings at or above `--fail-on`, and every file parsed successfully
- `1` — a finding met the `--fail-on` threshold, or a file failed to parse

This makes `kube-review` usable as a CI gate, e.g. in a GitHub Actions step:

```yaml
- run: kube-review review manifests/ --fail-on HIGH
```

### GitHub code scanning (SARIF)

On repos with GitHub Advanced Security / code scanning enabled, upload findings as PR annotations:

```yaml
- run: kube-review review manifests/ --fail-on NONE --output sarif > kube-review.sarif

- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: kube-review.sarif
```

Use `--fail-on NONE` when generating the SARIF report so the step doesn't exit non-zero before the upload step runs — let code scanning surface the findings instead of failing the build here.

---

## GitHub Action

Use `kube-review` as a step without building it yourself:

```yaml
- uses: loukasprevyzis/kube-review@main
  with:
    path: manifests/
    fail-on: HIGH        # optional, default HIGH
    output: text         # optional, default text (text, json, or sarif)
    config: ""           # optional, path to a .kube-review.yml
    values-files: ""     # optional, newline-separated Helm values files
```

The action builds `kube-review` from source at the referenced ref and installs Helm, so chart review works out of the box. Combined with SARIF output and code scanning:

```yaml
- uses: loukasprevyzis/kube-review@main
  id: review
  with:
    path: manifests/
    fail-on: NONE
    output: sarif

- uses: github/codeql-action/upload-sarif@v3
  with:
    sarif_file: ${{ steps.review.outputs.sarif-file }}
```

---

## Example Output

### Vulnerable Deployment

```text
File: examples/deployment.yaml
Deployment: payments-api

## Security

[HIGH] payments-api uses latest tag (ghcr.io/acme/payments-api:latest)
[HIGH] payments-api does not set runAsNonRoot=true
[HIGH] payments-api does not set allowPrivilegeEscalation=false

## Reliability

[MEDIUM] payments-api does not define a readiness probe
[MEDIUM] payments-api does not define a liveness probe

## Cost

[MEDIUM] payments-api has no resource limits configured
[MEDIUM] payments-api has no resource requests configured
```

### Compliant Deployment

```text
File: examples/good-deployment.yaml
Deployment: payments-api

No findings
```

---

## Current Rules

| Category | Rule | ID | Default Severity |
|-----------|--------|----|----|
| Security | Latest image tag | `latest-tag` | HIGH |
| Security | Missing runAsNonRoot | `run-as-non-root` | HIGH |
| Security | Missing allowPrivilegeEscalation=false | `allow-privilege-escalation` | HIGH |
| Security | Privileged container | `privileged-container` | HIGH |
| Reliability | Missing readiness probe | `readiness-probe` | MEDIUM |
| Reliability | Missing liveness probe | `liveness-probe` | MEDIUM |
| Cost | Missing resource requests | `resource-requests` | MEDIUM |
| Cost | Missing resource limits | `resource-limits` | MEDIUM |

Security and cost rules also check init containers. Reliability probe rules only apply to regular containers, since the kubelet ignores probes on init containers.

---

## Policy Configuration

Drop a `.kube-review.yml` in the directory you run `kube-review` from (or pass `--config path/to/file.yml`) to disable individual rules or override their severity:

```yaml
rules:
  readiness-probe:
    enabled: false        # don't require readiness probes on this repo
  latest-tag:
    severity: MEDIUM      # downgrade from the default HIGH
```

An unknown rule ID or an invalid severity value is a hard error, not a silent no-op — a typo here would otherwise disable a security check without warning.

---

## Project Structure

```text
kube-review/
├── examples/
├── internal/
│   ├── output/
│   ├── parser/
│   └── rules/
├── main.go
├── go.mod
└── README.md
```

---

## Roadmap

- [x] Kubernetes Deployment parsing
- [x] Security rule engine
- [x] Reliability rule engine
- [x] Cost rule engine
- [x] Recursive directory scanning
- [x] Unit tests
- [x] Multi-resource manifest support (multi-document YAML, multiple workload kinds)
- [x] Configurable severity gate (`--fail-on`)
- [x] JSON output
- [x] CI pipeline (GitHub Actions)
- [x] Configurable rule policies (enable/disable individual rules, per-rule severity overrides)
- [x] SARIF output
- [x] Helm chart support
- [x] GitHub Action integration (composite action wrapping the binary)

---

## Author

Built by Loukas Prevyzis as a Platform Engineering project focused on Platform Engineering, Kubernetes, and DevSecOps.