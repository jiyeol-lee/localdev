package app

import (
	"os"
	"testing"
)

func Test_defaultConfigFile(t *testing.T) {
	userConfigDir, _ := os.UserConfigDir()

	tests := []struct {
		name    string
		xdgEnv  string
		want    string
		wantErr bool
	}{
		{
			name:    "If XDG_CONFIG_HOME is set, it should return the config file path",
			xdgEnv:  "/home/user/.config",
			want:    "/home/user/.config/localdev/config.json",
			wantErr: false,
		},
		{
			name:    "If XDG_CONFIG_HOME is not set, it should return the default config file path",
			xdgEnv:  "",
			want:    userConfigDir + "/localdev/config.json",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", tt.xdgEnv)
			got, err := defaultConfigFile()
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
