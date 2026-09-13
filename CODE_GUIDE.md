# Code Guide (for Go beginners)

This walks through every `.go` file in this repo, in the order you'd want to
read them to understand the codebase, not alphabetically. If you're new to
Go, read the **Go concepts you'll see everywhere** section first — it
explains the language features this codebase leans on, so the file-by-file
walkthrough doesn't have to keep stopping to explain syntax.

---

## Go concepts you'll see everywhere

**Packages.** Every `.go` file starts with `package foo`. A package is Go's
unit of code organization — think of it like a namespace. All files in the
same directory belong to the same package and can see each other's code
without importing anything. This repo has one package per directory:
`main`, `parser`, `rules`, `output`, `workload`.

**`internal/`.** Any package under a directory named `internal/` can only be
imported by code inside this same module (`github.com/loukasprevyzis/kube-review`).
It's Go's way of saying "this is an implementation detail, not a public API."
That's why `parser`, `rules`, `output`, and `workload` all live under
`internal/` — nobody outside this repo is meant to `import` them directly.

**Exported vs. unexported names.** Go has no `public`/`private` keywords.
Instead: a name that starts with a **capital letter** (`LoadWorkloads`,
`Finding`, `RuleID`) is exported — visible from other packages. A name
starting **lowercase** (`parseWorkload`, `check`, `severityRank`) is
unexported — only visible inside its own package. Scan for capitalization
and you instantly know what's "public API" vs. internal plumbing.

**Structs.** Go's version of a plain data object:
```go
type Finding struct {
    RuleID   string
    Category string
    Severity string
    Message  string
}
```
No classes, no constructors required — `Finding{RuleID: "latest-tag", ...}`
just builds one.

**Pointers (`*T`).** `*workload.Workload` means "a pointer to a Workload" —
an address pointing at a Workload value somewhere in memory, rather than a
copy of it. You'll see this constantly: functions take `*workload.Workload`
so they can read the caller's actual struct without copying it, and methods
like `func (w *Workload) AllContainers()` are defined on a pointer so they
work regardless of whether you have a `Workload` or a `*Workload` in hand.

**Slices (`[]T`).** Go's resizable arrays. `[]Finding` is "a list of
Finding." `append(findings, newFinding)` adds one and returns the
(possibly reallocated) slice — which is why you'll always see
`findings = append(findings, ...)`, not just `append(findings, ...)` on its
own.

**Multiple return values and the error idiom.** Go functions routinely
return `(value, error)`:
```go
data, err := os.ReadFile(path)
if err != nil {
    return nil, err
}
```
There are no exceptions in Go. Every call that can fail returns an `error`
as its last value, and the caller is expected to check it immediately. This
`if err != nil { ... }` pattern appears dozens of times in this repo — it's
not boilerplate to skim past, it's the entire error-handling strategy.

**Interfaces.** A type is satisfied implicitly — there's no `implements`
keyword. If a type has all the methods an interface requires, it satisfies
that interface automatically. (This repo doesn't lean on interfaces much;
`rules.Rule` and `output.Result` are plain structs, not interfaces.)

**Struct tags.** The backtick strings after a struct field, like:
```go
type Finding struct {
    RuleID string `json:"ruleId"`
}
```
tell the `encoding/json` package what key name to use when converting this
struct to/from JSON. `omitempty` means "leave this key out entirely if the
field is empty."

**Methods vs. functions.** `func (w *Workload) AllContainers() []Container`
is a *method* on `Workload` — call it as `w.AllContainers()`. A plain
function like `func RunAll(w *workload.Workload) []Finding` is called as
`rules.RunAll(w)`. The `(w *Workload)` part before the function name is
what makes it a method instead of a standalone function.

**`_test.go` files.** Any file ending in `_test.go` is only compiled when
running `go test`, never part of the built binary. A function
`func TestSomething(t *testing.T)` is one test case; `go test ./...` runs
every test in every package.

---

## The big picture

```
                  ┌──────────┐
   file/dir/chart │  parser  │  reads YAML / renders Helm → []*workload.Workload
   on disk    ───►│          │
                  └────┬─────┘
                       │ *workload.Workload
                       ▼
                  ┌──────────┐
                  │  rules   │  runs every check → []Finding
                  └────┬─────┘
                       │ []Finding
                       ▼
                  ┌──────────┐
                  │  output  │  renders as text / JSON / SARIF
                  └──────────┘
```

`main.go` is the only place that wires these three packages together. Each
package only knows about the ones "below" it in this diagram — `rules`
doesn't know how to parse YAML, and `parser` doesn't know what a
"finding" even is. That separation is deliberate: it's what makes each
package independently testable (see all the `_test.go` files).

---

## `main.go` — the entry point

Every Go program's execution starts in `func main()` in `package main`.
This file's job is entirely **argument parsing and wiring** — it contains no
actual review logic itself, just calls into `parser`, `rules`, and `output`.

Walking through it top to bottom:

1. **Command check**: `os.Args` is the list of command-line arguments
   (`os.Args[0]` is the program name itself, `os.Args[1]` is `"review"`,
   etc.). The code checks the first argument is literally `"review"` —
   this repo only has one command.

2. **Manual flag parsing** (the `for i := 0; i < len(args); i++` loop):
   Go has a standard `flag` package for parsing `--foo bar` style options,
   but it stops parsing at the first non-flag argument — which broke this
   tool when someone wrote `kube-review review file.yaml --fail-on NONE`
   (flag *after* the path). So this code parses arguments by hand instead,
   recognizing `--fail-on`, `--output`, `--config`, and `--values` in either
   `--flag value` or `--flag=value` form, regardless of where they appear.
   Anything that isn't a recognized flag goes into `positional` — that's
   where the actual file/directory/chart path ends up.

3. **Validation**: `rules.ValidSeverityThreshold` and the `format != "text" && ...`
   check reject bad `--fail-on`/`--output` values immediately with a
   clear error, rather than silently doing the wrong thing.

4. **Policy loading**: if `--config` was given, or a `.kube-review.yml`
   file exists in the current directory, `rules.LoadPolicy` reads it (see
   the `rules` section below).

5. **The big `if isChart { ... } else { ... }` block**: this is the part
   that decides *what* to review.
   - If the path is a Helm chart (has a `Chart.yaml`), render it with
     `parser.LoadHelmChart` and review what comes out.
   - Otherwise, if it's a directory, find every `.yaml`/`.yml` file with
     `parser.ListYAMLFiles` *and* every nested Helm chart with
     `parser.ListHelmCharts` (so a chart living inside a plain directory
     of manifests is rendered, not read as raw YAML).
   - For every file/chart found, it calls `reviewWorkload` (a small helper
     defined later in the same file) which runs `rules.RunAll` and checks
     whether any finding is severe enough to fail the build.

6. **Output**: a `switch format { case "json": ... }` picks which
   `output.Print*` function to call.

7. **Exit code**: `os.Exit(1)` if anything failed to load or any finding
   met the `--fail-on` threshold; otherwise the function just returns,
   which exits with status `0` (success) — this is what lets CI use this
   tool as a pass/fail gate.

`reviewWorkload` and `printText` are small helper functions defined below
`main()` — Go doesn't require functions to be declared before they're used,
as long as they're in the same package.

---

## `internal/workload/workload.go`

The smallest package, and the one everything else depends on. Kubernetes has
many resource kinds (Deployment, StatefulSet, DaemonSet, Job, CronJob, bare
Pod...) but they all boil down to "some metadata plus a `PodSpec`" (the part
that actually describes containers, images, probes, etc.). Rather than
writing every security/reliability/cost check eight times — once per kind —
this package defines one small struct:

```go
type Workload struct {
    Kind string          // "Deployment", "StatefulSet", ...
    Name string
    Spec corev1.PodSpec  // from k8s.io/api/core/v1 — the real Kubernetes type
}
```

The `parser` package's job is converting each Kubernetes kind into this one
shape; the `rules` package's job is checking `Workload` values without ever
needing to know which original kind they came from.

`AllContainers()` is the one method: it returns `InitContainers` and
`Containers` concatenated, because most checks (image tag, root user,
resource limits) should apply to both — a compromised init container is
just as dangerous as a compromised main container. Probe checks
deliberately *don't* use this method (see `readiness_probe.go` below),
because Kubernetes itself ignores probes on init containers.

---

## `internal/parser/` — turning files into `Workload`s

### `parser.go`

`LoadWorkloads(path string)` is the entry point for a single file. Two
things make Kubernetes YAML trickier than "just unmarshal it":

1. **One file can contain multiple documents**, separated by `---` lines.
   `splitYAMLDocuments` handles this using a reader from Kubernetes's own
   `apimachinery` library (`k8s.io/apimachinery/pkg/util/yaml`) rather than
   naively splitting on `"---"` text, since that string can legally appear
   inside a YAML value too.

2. **You don't know the resource kind until you've partially parsed it.**
   `parseWorkload` first unmarshals just the `apiVersion`/`kind` fields
   (into `metav1.TypeMeta`), then does a `switch` on `Kind` to decide which
   real Kubernetes struct (`appsv1.Deployment`, `batchv1.Job`, `corev1.Pod`,
   etc.) to unmarshal the full document into. A `Service`, `ConfigMap`, or
   any other kind not in the `switch` falls to `default: return nil, nil` —
   not an error, just "nothing to review here," which is why `LoadWorkloads`
   checks `if w == nil { continue }`.

   Note this uses `sigs.k8s.io/yaml`, not Go's standard YAML library — it
   works by converting YAML to JSON internally and then using the same
   `encoding/json` tags the Kubernetes API types already have, which is why
   `appsv1.Deployment` "just works" with YAML input.

### `files.go`

Two directory-walking helpers, both built on Go's standard
`filepath.Walk`, which visits every file and subdirectory recursively and
calls your callback function for each one.

- `ListYAMLFiles` collects `.yaml`/`.yml` files, but explicitly skips
  descending into any directory that `IsHelmChart` recognizes — because a
  chart's `templates/*.yaml` files contain Go template syntax like
  `{{ .Values.image }}`, which isn't valid YAML on its own and would fail
  to parse.
- `ListHelmCharts` finds chart root directories (anywhere a `Chart.yaml`
  exists) and, once it finds one, returns `filepath.SkipDir` so it doesn't
  also report subcharts vendored inside `charts/` as separate top-level
  charts.

### `helm.go`

Rather than reimplementing Helm's templating engine, this shells out to the
real `helm` binary: `exec.Command("helm", "template", dir, ...)`. This is
the same "wrap a well-tested external tool via its CLI" pattern lots of
Go tools use instead of vendoring a whole dependency.

- `IsHelmChart` is just `os.Stat` on `<dir>/Chart.yaml`.
- `LoadHelmChart` runs `helm template`, captures stdout, and feeds it
  through the same `splitYAMLDocuments` + `parseWorkload` pipeline as a
  plain file.
- `sourceComment` is a small nicety: Helm annotates every document it
  renders with a `# Source: chart/templates/deployment.yaml` comment, and
  this function scans for that line so findings can be attributed to the
  actual template file instead of just "the chart directory."

---

## `internal/rules/` — the actual checks

### The eight `Check*.go` files

`latest_tag.go`, `resource_limits.go`, `resource_requests.go`,
`run_as_non_root.go`, `readiness_probe.go`, `liveness_probe.go`,
`privileged_container.go`, `allow_privilege_escalation.go` — these all
follow the exact same shape, so once you've read one you've read them all:

```go
func CheckLatestTag(w *workload.Workload) []Finding {
    var findings []Finding

    for _, c := range w.AllContainers() {
        if strings.HasSuffix(c.Image, ":latest") {
            findings = append(findings, Finding{
                Category: "Security",
                Severity: "HIGH",
                Message:  fmt.Sprintf("%s uses latest tag (%s)", c.Name, c.Image),
            })
        }
    }

    return findings
}
```

Loop over containers, check one condition, append a `Finding` if it's
violated. `readiness_probe.go` and `liveness_probe.go` are the two
exceptions: they loop over `w.Spec.Containers` directly instead of
`w.AllContainers()`, since probes on init containers are meaningless (the
kubelet ignores them).

### `finding.go`

The `Finding` struct — the one piece of data every check produces. Note the
`RuleID` field is *not* set by the `Check*` functions themselves; it gets
stamped on afterward in `all.go`. That's a deliberate separation: the check
functions only know "what's wrong," not "which rule ID I am" — that
bookkeeping lives in one place.

### `rule.go`

Defines the `Rule` struct (an `ID`, a human `Description`, and an
unexported `check` field holding the actual check function) and a package
level `registry` — a slice listing all eight rules. This is what makes the
system driven by data instead of a hardcoded list of function calls: adding
a ninth rule means adding one entry to `registry`, not touching `all.go` or
`main.go`.

`RuleIDs()` and `Registry()` are how other code (the policy loader, SARIF
output) can ask "what rules exist?" without duplicating this list.

### `all.go`

`RunAll(w, policy)` loops over `registry` and, for each rule:
1. Skips it if the `policy` says it's disabled.
2. Calls its `check` function.
3. Stamps `RuleID` onto every finding it produced.
4. Overwrites `Severity` if the policy specifies an override.

This is the one function `main.go` actually calls — everything above it
exists to make this loop possible.

### `severity.go`

Defines the three severity levels as string constants (`High`, `Medium`,
`Low`) plus a special `None` used only for `--fail-on`. `severityRank` is a
`map[string]int` giving each level a numeric weight so `MeetsThreshold` can
do `severityRank[severity] >= severityRank[threshold]` instead of a chain of
string comparisons.

### `policy.go`

`Policy` is what a `.kube-review.yml` file deserializes into — a map from
rule ID to `{enabled, severity}` overrides. `LoadPolicy` reads the file,
unmarshals it, and then **validates** it: an unknown rule ID or an invalid
severity string returns an error immediately rather than silently doing
nothing, because a typo'd rule ID that's silently ignored would mean a
security check got disabled without anyone noticing.

---

## `internal/output/` — rendering results

### `output.go`

`PrintFindings` is the plain-text renderer used by default. It buckets
findings by `Category` into a `map[string][]Finding`, then prints them in
a fixed order (`Security`, `Reliability`, `Cost`) rather than map iteration
order — Go deliberately randomizes map iteration order, so if this printed
`for category := range categories`, the section order would be different
every run.

### `json.go`

`Result` is the shape every output format actually renders — one per
file/workload (or per load failure, with `Error` set instead of
`Findings`). `PrintJSON` just wraps `encoding/json`'s `Encoder` with
indentation turned on.

### `sarif.go`

The most structurally complex file in the repo, but only because SARIF
(the format GitHub code scanning expects) is a deeply nested JSON schema.
Every `sarif*` type here is just a Go struct mirroring one small piece of
that schema — `sarifLog` → `sarifRun` → (`sarifTool` and `sarifResult`)
and so on. `PrintSARIF` builds one `sarifLog` value by:
1. Declaring every known rule up front (`sarifRules()`, from
   `rules.Registry()`), so tools reading the SARIF file know about a rule
   even in a run with zero findings for it.
2. Looping over `results` and turning each `Finding` into a `sarifResult`,
   mapping severity to SARIF's `level` vocabulary (`error`/`warning`/`note`)
   via `sarifLevel`.

Load failures (`r.Error != ""`) are skipped — SARIF describes findings
*in* a file, not "the file didn't parse."

---

## The `_test.go` files

Every test file follows Go's standard shape:
```go
func TestSomething(t *testing.T) {
    // 1. build some input
    // 2. call the function under test
    // 3. check the result, calling t.Fatalf/t.Errorf if it's wrong
}
```
`t.Fatalf` stops the test immediately (use it when a later check would
panic anyway, e.g. indexing into an empty slice); `t.Errorf` records a
failure but keeps running the rest of the test.

A few worth calling out specifically:

- **`internal/rules/*_test.go`** (one per `Check*.go` file): each builds a
  minimal `*workload.Workload` by hand, calls the `Check*` function
  directly, and asserts on the finding count. `run_as_non_root_test.go`
  has a second test case (`_InitContainer`) proving the rule also fires
  for init containers, not just regular ones.
- **`internal/rules/policy_test.go`**: tests `RunAll` respecting a
  disabled rule and a severity override, plus `LoadPolicy` rejecting an
  unknown rule ID and an invalid severity — using `t.TempDir()` to write a
  throwaway config file per test, which Go automatically cleans up.
- **`internal/parser/helm_test.go`**: `TestLoadHelmChart` and
  `TestLoadHelmChart_WithValuesOverride` actually shell out to a real
  `helm` binary — but `t.Skip(...)` bails out cleanly if `helm` isn't
  installed, so this test suite doesn't fail on a machine without Helm.
- **`internal/parser/files_test.go`**: `TestListYAMLFiles_SkipsNestedHelmChart`
  builds a fake directory tree with `t.TempDir()` containing both a plain
  YAML file and a nested chart, and asserts the chart's raw template
  doesn't get treated as standalone YAML — this is a regression test for a
  real bug that came up while building Helm support.
- **`internal/output/sarif_test.go`**: builds a `[]Result` by hand (one
  normal finding, one load error), runs it through `PrintSARIF`, then
  unmarshals the output back into the same `sarifLog` struct to assert on
  its shape — testing that the JSON round-trips correctly rather than
  string-matching the output.

---

## Suggested reading order

If you want to read the source files themselves in the order that builds
understanding most naturally:

1. `internal/workload/workload.go` — the core data shape
2. `internal/rules/finding.go`, `severity.go` — the other core data shapes
3. `internal/rules/latest_tag.go` — one concrete check, to see the pattern
4. `internal/rules/rule.go`, `all.go` — how checks are registered and run
5. `internal/rules/policy.go` — how checks get configured
6. `internal/parser/parser.go` — how a Kubernetes file becomes a `Workload`
7. `internal/parser/files.go`, `helm.go` — directory scanning and Helm
8. `internal/output/output.go`, `json.go`, `sarif.go` — rendering
9. `main.go` — how it all gets wired together and invoked
