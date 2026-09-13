package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8syaml "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	"github.com/loukasprevyzis/kube-review/internal/workload"
)

// LoadWorkloads reads a YAML file that may contain multiple "---"-separated
// documents and returns one workload per document whose kind is supported.
// Documents of unsupported or non-workload kinds (Service, ConfigMap, ...)
// are silently skipped.
func LoadWorkloads(path string) ([]*workload.Workload, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	docs, err := splitYAMLDocuments(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}

	var workloads []*workload.Workload

	for _, doc := range docs {
		w, err := parseWorkload(doc)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if w == nil {
			continue
		}

		workloads = append(workloads, w)
	}

	return workloads, nil
}

// splitYAMLDocuments splits a "---"-separated YAML stream into its
// individual documents, dropping empty ones.
func splitYAMLDocuments(data []byte) ([][]byte, error) {
	var docs [][]byte

	reader := k8syaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))

	for {
		doc, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(bytes.TrimSpace(doc)) == 0 {
			continue
		}

		docs = append(docs, doc)
	}

	return docs, nil
}

func parseWorkload(doc []byte) (*workload.Workload, error) {

	// A directory scan runs into plenty of non-Kubernetes YAML: docker-compose
	// files, kustomizations, CI configs, or documents that aren't even an
	// object (a bare list or scalar). Object-shaped YAML with an unknown
	// `kind` already falls through the switch below and is skipped, but a
	// non-object document can't be unmarshaled into TypeMeta at all - check
	// for that up front and skip it too, rather than surfacing it as a load
	// failure. A genuine YAML syntax error is still reported, since that
	// could be a real mistake in an actual manifest.
	var generic interface{}
	if err := yaml.Unmarshal(doc, &generic); err != nil {
		return nil, err
	}
	if _, ok := generic.(map[string]interface{}); !ok {
		return nil, nil
	}

	var meta metav1.TypeMeta

	if err := yaml.Unmarshal(doc, &meta); err != nil {
		return nil, err
	}

	switch strings.TrimSpace(meta.Kind) {

	case "Deployment":
		var d appsv1.Deployment
		if err := yaml.Unmarshal(doc, &d); err != nil {
			return nil, err
		}
		return &workload.Workload{Kind: "Deployment", Name: d.Name, Spec: d.Spec.Template.Spec}, nil

	case "StatefulSet":
		var s appsv1.StatefulSet
		if err := yaml.Unmarshal(doc, &s); err != nil {
			return nil, err
		}
		return &workload.Workload{Kind: "StatefulSet", Name: s.Name, Spec: s.Spec.Template.Spec}, nil

	case "DaemonSet":
		var ds appsv1.DaemonSet
		if err := yaml.Unmarshal(doc, &ds); err != nil {
			return nil, err
		}
		return &workload.Workload{Kind: "DaemonSet", Name: ds.Name, Spec: ds.Spec.Template.Spec}, nil

	case "ReplicaSet":
		var rs appsv1.ReplicaSet
		if err := yaml.Unmarshal(doc, &rs); err != nil {
			return nil, err
		}
		return &workload.Workload{Kind: "ReplicaSet", Name: rs.Name, Spec: rs.Spec.Template.Spec}, nil

	case "Job":
		var j batchv1.Job
		if err := yaml.Unmarshal(doc, &j); err != nil {
			return nil, err
		}
		return &workload.Workload{Kind: "Job", Name: j.Name, Spec: j.Spec.Template.Spec}, nil

	case "CronJob":
		var cj batchv1.CronJob
		if err := yaml.Unmarshal(doc, &cj); err != nil {
			return nil, err
		}
		return &workload.Workload{Kind: "CronJob", Name: cj.Name, Spec: cj.Spec.JobTemplate.Spec.Template.Spec}, nil

	case "Pod":
		var p corev1.Pod
		if err := yaml.Unmarshal(doc, &p); err != nil {
			return nil, err
		}
		return &workload.Workload{Kind: "Pod", Name: p.Name, Spec: p.Spec}, nil

	default:
		return nil, nil
	}
}
