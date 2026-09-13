package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/loukasprevyzis/kube-review/internal/output"
	"github.com/loukasprevyzis/kube-review/internal/parser"
	"github.com/loukasprevyzis/kube-review/internal/rules"
	"github.com/loukasprevyzis/kube-review/internal/workload"
)

const defaultConfigFile = ".kube-review.yml"

func usage() {
	fmt.Println("Usage: kube-review review [--fail-on HIGH|MEDIUM|LOW|NONE] [--output text|json|sarif] [--config path] [--values file]... <file-or-directory-or-helm-chart>")
}

func main() {

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	command := os.Args[1]

	if command != "review" {
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}

	failOn := rules.High
	outputFormat := "text"
	configPath := ""
	var valuesFiles []string
	var positional []string

	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		arg := args[i]

		switch {
		case arg == "--fail-on":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --fail-on requires a value")
				os.Exit(1)
			}
			failOn = args[i]

		case strings.HasPrefix(arg, "--fail-on="):
			failOn = strings.TrimPrefix(arg, "--fail-on=")

		case arg == "--output":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --output requires a value")
				os.Exit(1)
			}
			outputFormat = args[i]

		case strings.HasPrefix(arg, "--output="):
			outputFormat = strings.TrimPrefix(arg, "--output=")

		case arg == "--config":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --config requires a value")
				os.Exit(1)
			}
			configPath = args[i]

		case strings.HasPrefix(arg, "--config="):
			configPath = strings.TrimPrefix(arg, "--config=")

		case arg == "--values":
			i++
			if i >= len(args) {
				fmt.Fprintln(os.Stderr, "Error: --values requires a value")
				os.Exit(1)
			}
			valuesFiles = append(valuesFiles, args[i])

		case strings.HasPrefix(arg, "--values="):
			valuesFiles = append(valuesFiles, strings.TrimPrefix(arg, "--values="))

		default:
			positional = append(positional, arg)
		}
	}

	threshold := strings.ToUpper(failOn)
	if !rules.ValidSeverityThreshold(threshold) {
		fmt.Fprintf(os.Stderr, "Error: invalid --fail-on value %q (expected HIGH, MEDIUM, LOW, or NONE)\n", failOn)
		os.Exit(1)
	}

	format := strings.ToLower(outputFormat)
	if format != "text" && format != "json" && format != "sarif" {
		fmt.Fprintf(os.Stderr, "Error: invalid --output value %q (expected text, json, or sarif)\n", outputFormat)
		os.Exit(1)
	}

	if len(positional) < 1 {
		usage()
		os.Exit(1)
	}

	path := positional[0]
	isChart := parser.IsHelmChart(path)

	if len(valuesFiles) > 0 && !isChart {
		fmt.Fprintf(os.Stderr, "Error: --values only applies when <path> is a Helm chart (%s has no Chart.yaml)\n", path)
		os.Exit(1)
	}

	if configPath == "" {
		if _, err := os.Stat(defaultConfigFile); err == nil {
			configPath = defaultConfigFile
		}
	}

	var policy rules.Policy
	if configPath != "" {
		p, err := rules.LoadPolicy(configPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
		policy = p
	}

	var results []output.Result
	hadFailure := false
	shouldFail := false

	if isChart {

		rendered, err := parser.LoadHelmChart(path, valuesFiles)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		for _, rw := range rendered {
			result, fail := reviewWorkload(rw.Source, rw.Workload, policy, threshold)
			results = append(results, result)
			if fail {
				shouldFail = true
			}
		}

	} else {

		var files []string
		var chartDirs []string

		if parser.IsDirectory(path) {

			found, err := parser.ListYAMLFiles(path)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			files = found

			charts, err := parser.ListHelmCharts(path)

			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}

			chartDirs = charts

		} else {
			files = []string{path}
		}

		for _, file := range files {

			workloads, err := parser.LoadWorkloads(file)

			if err != nil {
				results = append(results, output.Result{File: file, Error: err.Error()})
				hadFailure = true
				continue
			}

			for _, w := range workloads {
				result, fail := reviewWorkload(file, w, policy, threshold)
				results = append(results, result)
				if fail {
					shouldFail = true
				}
			}
		}

		for _, chartDir := range chartDirs {

			rendered, err := parser.LoadHelmChart(chartDir, nil)

			if err != nil {
				results = append(results, output.Result{File: chartDir, Error: err.Error()})
				hadFailure = true
				continue
			}

			for _, rw := range rendered {
				result, fail := reviewWorkload(rw.Source, rw.Workload, policy, threshold)
				results = append(results, result)
				if fail {
					shouldFail = true
				}
			}
		}
	}

	switch format {
	case "json":
		if err := output.PrintJSON(os.Stdout, results); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	case "sarif":
		if err := output.PrintSARIF(os.Stdout, results); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	default:
		printText(results)
	}

	if hadFailure || shouldFail {
		os.Exit(1)
	}
}

// reviewWorkload runs every rule against w and reports whether any finding
// met the fail-on threshold.
func reviewWorkload(file string, w *workload.Workload, policy rules.Policy, threshold string) (output.Result, bool) {

	findings := rules.RunAll(w, policy)

	shouldFail := false
	for _, f := range findings {
		if rules.MeetsThreshold(f.Severity, threshold) {
			shouldFail = true
		}
	}

	return output.Result{
		File:     file,
		Kind:     w.Kind,
		Name:     w.Name,
		Findings: findings,
	}, shouldFail
}

func printText(results []output.Result) {
	for _, r := range results {

		fmt.Println()
		fmt.Println("File:", r.File)

		if r.Error != "" {
			fmt.Println("Failed to load:", r.Error)
			continue
		}

		fmt.Printf("%s: %s\n", r.Kind, r.Name)

		if len(r.Findings) == 0 {
			fmt.Println("No findings")
			continue
		}

		output.PrintFindings(r.Findings)
	}
}
