package utils

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadYAML[T any](path string) (*T, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("leyendo %s: %w", path, err)
	}

	var out T
	if err := yaml.Unmarshal(data, &out); err != nil {
		return nil, fmt.Errorf("parseando %s: %w", path, err)
	}
	return &out, nil
}
