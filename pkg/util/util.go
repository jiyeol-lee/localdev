package util

import (
	"github.com/goccy/go-yaml"
)

func YamlToMap[T any, V any](data T) (map[string]V, error) {
	jsonData, err := yaml.Marshal(data)
	if err != nil {
		return nil, err
	}

	var result map[string]V
	err = yaml.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
