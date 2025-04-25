package view

import (
	"fmt"
	"strings"
)

// convertCommandKeyToCharater converts a key string to its corresponding character.
func convertCommandKeyToCharater(key string) (string, error) {
	if strings.HasPrefix(key, "lower") && len(key) == 6 {
		return strings.ToLower(string(key[5])), nil
	}

	if strings.HasPrefix(key, "upper") && len(key) == 6 {
		return strings.ToUpper(string(key[5])), nil
	}

	return "", fmt.Errorf("invalid key format: %s", key)
}
