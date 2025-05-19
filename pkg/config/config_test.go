package config

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func Test_defaultConfigFile(t *testing.T) {
	userConfigDir, _ := os.UserConfigDir()

	type args struct {
		configFileName string
	}
	tests := []struct {
		name    string
		xdgEnv  string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   "If XDG_CONFIG_HOME is set, it should return the config file path",
			xdgEnv: "/home/user/.config",
			args: args{
				configFileName: "config.json",
			},
			want:    "/home/user/.config/localdev/config.json",
			wantErr: false,
		},
		{
			name:   "If XDG_CONFIG_HOME is not set, it should return the default config file path",
			xdgEnv: "",
			args: args{
				configFileName: "config.json",
			},
			want:    userConfigDir + "/localdev/config.json",
			wantErr: false,
		},
		{
			name:   "If XDG_CONFIG_HOME is set and config file is not config.json, it should use the provided name",
			xdgEnv: "/home/user/.config",
			args: args{
				configFileName: "custom_config.json",
			},
			want:    "/home/user/.config/localdev/custom_config.json",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tt.xdgEnv)
			got, err := defaultConfigFile(tt.args.configFileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("defaultConfigFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("defaultConfigFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ConfigValidation(t *testing.T) {
	t.Run("all required fields present", func(t *testing.T) {
		cfg := &Config{
			Panes: []ConfigPane{
				{
					Name:  "pane1",
					Dir:   "/tmp",
					Start: "echo start",
					Stop:  "echo stop",
				},
				{
					Name:  "pane2",
					Dir:   "/tmp2",
					Start: "echo start2",
					Stop:  "echo stop2",
				},
			},
		}
		err := cfg.LoadConfigFromStruct()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("missing some required fields in multiple panes", func(t *testing.T) {
		cfg := &Config{
			Panes: []ConfigPane{
				{
					Name:  "pane1",
					Dir:   "",
					Start: "echo start",
					Stop:  "echo stop",
				},
				{
					Name:  "",
					Dir:   "/tmp2",
					Start: "",
					Stop:  "echo stop2",
				},
			},
		}
		err := cfg.LoadConfigFromStruct()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errMsg := err.Error()
		// Pane 0 is missing Dir
		if !strings.Contains(errMsg, "pane[0] is missing required field: dir") {
			t.Errorf("expected missing dir error for pane[0], got: %v", errMsg)
		}
		// Pane 1 is missing Name and Start
		if !strings.Contains(errMsg, "pane[1] is missing required field: name") {
			t.Errorf("expected missing name error for pane[1], got: %v", errMsg)
		}
		if !strings.Contains(errMsg, "pane[1] is missing required field: start") {
			t.Errorf("expected missing start error for pane[1], got: %v", errMsg)
		}
		// Pane 1 should not have a missing stop error
		if strings.Contains(errMsg, "pane[1] is missing required field: stop") {
			t.Errorf("did not expect missing stop error for pane[1], got: %v", errMsg)
		}
	})
}

// Helper for testing validation logic directly
func (c *Config) LoadConfigFromStruct() error {
	if len(c.Panes) == 0 {
		return fmt.Errorf("configuration must contain at least one pane")
	}
	var validationErrors []string
	for i, pane := range c.Panes {
		if pane.Name == "" {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("pane[%d] is missing required field: name", i),
			)
		}
		if pane.Dir == "" {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("pane[%d] is missing required field: dir", i),
			)
		}
		if pane.Start == "" {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("pane[%d] is missing required field: start", i),
			)
		}
		if pane.Stop == "" {
			validationErrors = append(
				validationErrors,
				fmt.Sprintf("pane[%d] is missing required field: stop", i),
			)
		}
	}
	if len(validationErrors) > 0 {
		return fmt.Errorf(
			"configuration validation errors:\n%s",
			strings.Join(validationErrors, "\n"),
		)
	}
	return nil
}
