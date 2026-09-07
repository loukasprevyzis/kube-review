package parser

import (
	"os"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/yaml"
)

func LoadDeployment(path string) (*appsv1.Deployment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var deployment appsv1.Deployment

	err = yaml.Unmarshal(data, &deployment)
	if err != nil {
		return nil, err
	}

	return &deployment, nil
}
