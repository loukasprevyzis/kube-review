package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/loukasprevyzis/kube-review/internal/output"
	"github.com/loukasprevyzis/kube-review/internal/parser"
	"github.com/loukasprevyzis/kube-review/internal/rules"
)

const defaultConfigFile = ".kube-review.yml"

func usage() {
	fmt.Println("Usage: kube-review review [--fail-on HIGH|MEDIUM|LOW|NONE] [--output text|json] [--config path] <file-or-directory>")
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
	if format != "text" && format != "json" {
		fmt.Fprintf(os.Stderr, "Error: invalid --output value %q (expected text or json)\n", outputFormat)
		os.Exit(1)
	}

	if len(positional) < 1 {
		usage()
		os.Exit(1)
	}

	path := positional[0]

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

	var files []string

	if parser.IsDirectory(path) {

		found, err := parser.ListYAMLFiles(path)

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}

		files = found

	} else {
		files = []string{path}
	}

	var results []output.Result
	hadFailure := false
	shouldFail := false

	for _, file := range files {

		workloads, err := parser.LoadWorkloads(file)

		if err != nil {
			results = append(results, output.Result{File: file, Error: err.Error()})
			hadFailure = true
			continue
		}

		for _, w := range workloads {

			findings := rules.RunAll(w, policy)

			results = append(results, output.Result{
				File:     file,
				Kind:     w.Kind,
				Name:     w.Name,
				Findings: findings,
			})

			for _, f := range findings {
				if rules.MeetsThreshold(f.Severity, threshold) {
					shouldFail = true
				}
			}
		}
	}

	if format == "json" {
		if err := output.PrintJSON(os.Stdout, results); err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	} else {
		printText(results)
	}

	if hadFailure || shouldFail {
		os.Exit(1)
	}
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
