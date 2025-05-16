package util

import (
	"github.com/goccy/go-yaml"
)

func YamlToMap[T any, V any](data T) (map[string]V, error) {
	yamlData, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result map[string]V
	err = yaml.Unmarshal(yamlData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
