package config

import (
	"os"
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
