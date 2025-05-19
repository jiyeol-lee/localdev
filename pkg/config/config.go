package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

// defaultConfigFile constructs the default configuration file path using the provided configFileName.
func defaultConfigFile(configFileName string) (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configDir = dir
	}
	return filepath.Join(configDir, "localdev", configFileName), nil
}

// LoadConfig loads the configuration from the specified configuration file name.
// A valid configFileName must be provided as an argument.
func (c *Config) LoadConfig(configFileName string) error {
	configFile, err := defaultConfigFile(configFileName)
	if err != nil {
		return err
	}

	file, err := os.ReadFile(configFile)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	err = yaml.Unmarshal(file, c)
	if err != nil {
		return err
	}

	// Configuration validation: ensure at least one pane and required fields are present
	if len(c.Panes) == 0 {
		return fmt.Errorf("configuration must contain at least one pane")
	}
	var validationErrors []string
	for i, pane := range c.Panes {
		if pane.Name == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("pane[%d] is missing required field: name", i))
		}
		if pane.Dir == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("pane[%d] is missing required field: dir", i))
		}
		if pane.Start == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("pane[%d] is missing required field: start", i))
		}
		if pane.Stop == "" {
			validationErrors = append(validationErrors, fmt.Sprintf("pane[%d] is missing required field: stop", i))
		}
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf("configuration validation errors:\n%s", strings.Join(validationErrors, "\n"))
	}

	return nil
}
