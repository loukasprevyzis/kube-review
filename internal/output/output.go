package output

import (
	"fmt"

	"github.com/loukasprevyzis/kube-review/internal/rules"
)

func PrintFindings(findings []rules.Finding) {

	categories := map[string][]rules.Finding{}

	for _, finding := range findings {
		categories[finding.Category] = append(
			categories[finding.Category],
			finding,
		)
	}

	order := []string{
		"Security",
		"Reliability",
		"Cost",
	}

	for _, category := range order {

		items, exists := categories[category]

		if !exists || len(items) == 0 {
			continue
		}

		fmt.Printf("## %s\n\n", category)

		for _, item := range items {
			fmt.Printf(
				"[%s] %s\n",
				item.Severity,
				item.Message,
			)
		}

		fmt.Println()
	}
}
