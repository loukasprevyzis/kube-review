package rules

import "github.com/loukasprevyzis/kube-review/internal/workload"

type Rule interface {
	Name() string
	Check(*workload.Workload) []Finding
}
