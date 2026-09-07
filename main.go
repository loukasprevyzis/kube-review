package main

import (
	"fmt"
	"os"

	"github.com/loukasprevyzis/kube-review/internal/output"
	"github.com/loukasprevyzis/kube-review/internal/parser"
	"github.com/loukasprevyzis/kube-review/internal/rules"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Println("Usage: kube-review review <file-or-directory>")
		os.Exit(1)
	}

	command := os.Args[1]

	if command != "review" {
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}

	path := os.Args[2]

	if parser.IsDirectory(path) {

		files, err := parser.ListYAMLFiles(path)

		if err != nil {
			panic(err)
		}

		for _, file := range files {

			deployment, err := parser.LoadDeployment(file)

			if err != nil {
				fmt.Printf("Failed to load %s: %v\n", file, err)
				continue
			}

			fmt.Println()
			fmt.Println("File:", file)
			fmt.Println("Deployment:", deployment.Name)

			findings := rules.RunAll(deployment)

			if len(findings) == 0 {
				fmt.Println("No findings")
				continue
			}

			output.PrintFindings(findings)
		}

		return
	}

	deployment, err := parser.LoadDeployment(path)

	if err != nil {
		panic(err)
	}

	fmt.Println("Deployment:", deployment.Name)
	fmt.Println()

	findings := rules.RunAll(deployment)

	if len(findings) == 0 {
		fmt.Println("No findings")
		return
	}

	output.PrintFindings(findings)
}
