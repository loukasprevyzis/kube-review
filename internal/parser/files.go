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

func ListYAMLFiles(dir string) ([]string, error) {

	var files []string

	err := filepath.Walk(
		dir,
		func(path string, info os.FileInfo, err error) error {

			if err != nil {
				return err
			}

			if info.IsDir() {
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
