package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ConfigSpace struct {
	Name  string `json:"name"`
	Dir   string `json:"dir"`
	Start string `json:"start"`
	Stop  string `json:"stop"`
}

type Config struct {
	Spaces []ConfigSpace `json:"spaces"`
}

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
func (c *Config) loadConfig() error {
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
