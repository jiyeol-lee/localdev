package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitWithPathCreatesPrivateAppendLogFile(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "nested", "localdev.log")

	if err := InitWithPath(logPath); err != nil {
		t.Fatalf("InitWithPath() error = %v", err)
	}
	Infof("first %s", "line")
	if err := Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("log file mode = %v, want 0600", got)
	}

	if err := InitWithPath(logPath); err != nil {
		t.Fatalf("second InitWithPath() error = %v", err)
	}
	Warnf("second line")
	if err := Close(); err != nil {
		t.Fatalf("second Close() error = %v", err)
	}

	contents, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	logText := string(contents)
	if !strings.Contains(logText, "INFO: first line") ||
		!strings.Contains(logText, "WARN: second line") {
		t.Fatalf("log contents did not append expected entries: %q", logText)
	}
	if Path() != logPath {
		t.Fatalf("Path() = %q, want %q", Path(), logPath)
	}
}

func TestInitWithPathRestrictsExistingBroadPermissions(t *testing.T) {
	logPath := filepath.Join(t.TempDir(), "localdev.log")
	if err := os.WriteFile(logPath, []byte("existing\n"), 0o644); err != nil {
		t.Fatalf("pre-create log file: %v", err)
	}

	if err := InitWithPath(logPath); err != nil {
		t.Fatalf("InitWithPath() error = %v", err)
	}
	if err := Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("existing log file mode = %v, want 0600", got)
	}
}

func TestInitUsesProductionCachePath(t *testing.T) {
	cacheDir := t.TempDir()
	originalUserCacheDir := userCacheDir
	userCacheDir = func() (string, error) { return cacheDir, nil }
	t.Cleanup(func() {
		_ = Close()
		userCacheDir = originalUserCacheDir
	})

	if err := Init(); err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	wantPath := filepath.Join(cacheDir, "localdev", "localdev.log")
	if Path() != wantPath {
		t.Fatalf("Path() = %q, want %q", Path(), wantPath)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("stat production log path: %v", err)
	}
}

func TestHelpersAreSafeNoOpsBeforeInitialization(t *testing.T) {
	if err := Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	Infof("ignored")
	Warnf("ignored")
	Errorf("ignored")
	if Initialized() {
		t.Fatal("Initialized() = true, want false")
	}
}
