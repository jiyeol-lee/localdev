package config

import (
	"encoding/json"
	"os"
	"path/filepath"
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

	err = json.Unmarshal(file, c)
	if err != nil {
		return err
	}

	return nil
}
