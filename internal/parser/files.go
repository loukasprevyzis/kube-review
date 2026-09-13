package parser

import (
	"os"
	"path/filepath"
)

func IsDirectory(path string) bool {

	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	return info.IsDir()
}

// ListYAMLFiles walks dir and returns every .yaml/.yml file found, except
// those inside a nested Helm chart: chart templates use Go template syntax
// and aren't valid standalone YAML, so ListHelmCharts handles those instead.
func ListYAMLFiles(dir string) ([]string, error) {

	var files []string

	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if info.IsDir() {
				if path != dir && IsHelmChart(path) {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(path)

			if ext == ".yaml" || ext == ".yml" {
				files = append(files, path)
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return files, nil
}

// ListHelmCharts walks dir and returns the root directory of every Helm
// chart found. It does not descend into a chart once found, so subcharts
// vendored under charts/ aren't reported as separate top-level charts.
func ListHelmCharts(dir string) ([]string, error) {

	var charts []string

	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if !info.IsDir() {
				return nil
			}

			if IsHelmChart(path) {
				charts = append(charts, path)
				return filepath.SkipDir
			}

			return nil
		},
	)

	if err != nil {
		return nil, err
	}

	return charts, nil
}
