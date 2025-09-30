package env_vars_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	envvars "github.com/jiyeol-lee/localdev/pkg/internal/env_vars"
)

func TestRunCommandAndCaptureEnvVars(t *testing.T) {
	type args struct {
		command string
	}
	tests := []struct {
		name   string
		args   args
		assert func(t *testing.T, beforeFile, afterFile string, err error)
	}{
		{
			name: "command adds single variable",
			args: args{command: "TEST_ADDED_VAR=hello; export TEST_ADDED_VAR"},
			assert: func(t *testing.T, beforeFile, afterFile string, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				for _, f := range []string{beforeFile, afterFile} {
					info, statErr := os.Stat(f)
					if statErr != nil {
						t.Fatalf("expected file %s to exist: %v", f, statErr)
					}
					if info.Size() == 0 {
						t.Fatalf("expected file %s not to be empty", f)
					}
				}
				diff, diffErr := envvars.GetDiffEnvVars(beforeFile, afterFile)
				if diffErr != nil {
					t.Fatalf("diff error: %v", diffErr)
				}
				found := false
				for _, line := range diff {
					if strings.HasPrefix(line, "TEST_ADDED_VAR=") {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected TEST_ADDED_VAR in diff, got %v", diff)
				}
			},
		},
		{
			name: "command modifies existing variable",
			args: args{command: "PATH=/tmp/custom:$PATH; export PATH"},
			assert: func(t *testing.T, beforeFile, afterFile string, err error) {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				diff, diffErr := envvars.GetDiffEnvVars(beforeFile, afterFile)
				if diffErr != nil {
					t.Fatalf("diff error: %v", diffErr)
				}
				foundPath := false
				for _, d := range diff {
					if strings.HasPrefix(d, "PATH=") && strings.Contains(d, "/tmp/custom") {
						foundPath = true
					}
				}
				if !foundPath {
					t.Fatalf("expected modified PATH in diff, got %v", diff)
				}
			},
		},
		{
			name: "failing command returns error and no after capture",
			args: args{command: "exit 5"},
			assert: func(t *testing.T, beforeFile, afterFile string, err error) {
				if err == nil {
					t.Fatalf("expected error for failing command")
				}
				if beforeFile == "" {
					t.Fatalf("expected before file path on failure (got empty)")
				}
				if afterFile != "" {
					t.Fatalf("expected empty after file path on failure (got %s)", afterFile)
				}
				if _, statErr := os.Stat(beforeFile); statErr != nil {
					t.Fatalf("expected before file to exist: %v", statErr)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforeFile, afterFile, err := envvars.RunCommandAndCaptureEnvVars(tt.args.command)
			t.Cleanup(func() {
				for _, f := range []string{beforeFile, afterFile} {
					if f != "" {
						_ = os.Remove(f)
					}
				}
			})
			for _, f := range []string{beforeFile, afterFile} {
				if f == "" {
					continue
				}
				if !strings.HasPrefix(f, os.TempDir()) && !filepath.IsAbs(f) {
					t.Fatalf("expected temp file path, got %s", f)
				}
			}
			tt.assert(t, beforeFile, afterFile, err)
		})
	}
}

func TestGetDiffEnvVars(t *testing.T) {
	type args struct{ beforeFile, afterFile string }
	tests := []struct {
		name    string
		args    args
		setup   func(t *testing.T) (beforePath, afterPath string)
		want    []string
		wantErr bool
	}{
		{
			name: "added and modified vars are returned",
			setup: func(t *testing.T) (string, string) {
				before, err := os.CreateTemp("", "env-before-")
				if err != nil {
					t.Fatalf("temp: %v", err)
				}
				after, err := os.CreateTemp("", "env-after-")
				if err != nil {
					t.Fatalf("temp: %v", err)
				}
				_, _ = before.WriteString("A=1\x00B=2\x00C=unchanged\x00")
				_, _ = after.WriteString("A=1\x00B=3\x00C=unchanged\x00D=new\x00")
				before.Close()
				after.Close()
				return before.Name(), after.Name()
			},
			want:    []string{"B=3", "D=new"},
			wantErr: false,
		},
		{
			name: "missing before file returns error",
			setup: func(t *testing.T) (string, string) {
				bf := filepath.Join(os.TempDir(), "nonexistent-before")
				af, _ := os.CreateTemp("", "env-after-")
				af.WriteString("X=1\x00")
				af.Close()
				return bf, af.Name()
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "no differences returns empty slice",
			setup: func(t *testing.T) (string, string) {
				before, _ := os.CreateTemp("", "env-before-")
				after, _ := os.CreateTemp("", "env-after-")
				before.WriteString("Z=9\x00Y=8\x00")
				after.WriteString("Z=9\x00Y=8\x00")
				before.Close()
				after.Close()
				return before.Name(), after.Name()
			},
			want:    []string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beforePath, afterPath := tt.args.beforeFile, tt.args.afterFile
			if tt.setup != nil {
				beforePath, afterPath = tt.setup(t)
				t.Cleanup(func() { _ = os.Remove(beforePath); _ = os.Remove(afterPath) })
			}
			got, err := envvars.GetDiffEnvVars(beforePath, afterPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDiffEnvVars() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if tt.want != nil {
				gotSorted := append([]string{}, got...)
				wantSorted := append([]string{}, tt.want...)
				sortFn := func(s []string) {
					for i := range s {
						for j := i + 1; j < len(s); j++ {
							if s[j] < s[i] {
								s[i], s[j] = s[j], s[i]
							}
						}
					}
				}
				sortFn(gotSorted)
				sortFn(wantSorted)
				if !reflect.DeepEqual(gotSorted, wantSorted) {
					t.Errorf("GetDiffEnvVars() = %v, want %v", gotSorted, wantSorted)
				}
			}
		})
	}
}
