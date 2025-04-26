package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// defaultConfigFile returns the default configuration file path.
func defaultConfigFile() (string, error) {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		dir, err := os.UserConfigDir()
		if err != nil {
			return "", err
		}
		configDir = dir
	}
	return filepath.Join(configDir, "localdev", "config.json"), nil
}

// loadConfig loads the configuration from the default config file.
func (c *Config) LoadConfig() error {
	configFile, err := defaultConfigFile()
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
