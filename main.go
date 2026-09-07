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
		fmt.Println("Usage: kube-review review <file>")
		os.Exit(1)
	}

	command := os.Args[1]

	if command != "review" {
		fmt.Println("Unknown command:", command)
		os.Exit(1)
	}

	deployment, err := parser.LoadDeployment(os.Args[2])

	if err != nil {
		panic(err)
	}

	fmt.Println("Deployment:", deployment.Name)

	findings := rules.RunAll(deployment)

	fmt.Println()

	if len(findings) == 0 {
		fmt.Println("No findings")
		return
	}

	output.PrintFindings(findings)
}
