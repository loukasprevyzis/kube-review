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

The tool will recursively discover Kubernetes manifests and review each file.

### Flags

| Flag | Values | Default | Description |
|------|--------|---------|-------------|
| `--fail-on` | `HIGH`, `MEDIUM`, `LOW`, `NONE` | `HIGH` | Minimum severity that causes a non-zero exit code. `NONE` never fails on findings (a file that fails to parse still exits non-zero). |
| `--output` | `text`, `json` | `text` | Output format. `json` is intended for CI/tooling integration. |

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

| Category | Rule | Severity |
|-----------|--------|----------|
| Security | Latest image tag | HIGH |
| Security | Missing runAsNonRoot | HIGH |
| Security | Missing allowPrivilegeEscalation=false | HIGH |
| Reliability | Missing readiness probe | MEDIUM |
| Reliability | Missing liveness probe | MEDIUM |
| Cost | Missing resource requests | MEDIUM |
| Cost | Missing resource limits | MEDIUM |

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
- [ ] Helm chart support
- [ ] SARIF output
- [ ] GitHub Action integration (composite action wrapping the binary)
- [ ] Configurable rule policies (enable/disable individual rules, per-rule severity overrides)

---

## Author

Built by Loukas Prevyzis as a Go learning project focused on Platform Engineering, Kubernetes, and DevSecOps.