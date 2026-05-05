package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

var (
	mu           sync.Mutex
	file         *os.File
	logger       *log.Logger
	currentPath  string
	userCacheDir = os.UserCacheDir
)

// Init initializes file logging at the production log path.
func Init() error {
	path, err := defaultPath()
	if err != nil {
		return err
	}
	return InitWithPath(path)
}

// InitWithPath initializes file logging at the provided path.
// It is intended for tests and other callers that need deterministic paths.
func InitWithPath(path string) error {
	mu.Lock()
	defer mu.Unlock()

	currentPath = path
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create log directory %q: %w", filepath.Dir(path), err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("open log file %q: %w", path, err)
	}
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return fmt.Errorf("set log file permissions %q: %w", path, err)
	}

	oldFile := file
	file = f
	logger = log.New(f, "", log.LstdFlags)
	if oldFile != nil {
		_ = oldFile.Close()
	}
	return nil
}

// Close closes the active log file, if initialized.
func Close() error {
	mu.Lock()
	defer mu.Unlock()

	if file == nil {
		logger = nil
		return nil
	}
	err := file.Close()
	file = nil
	logger = nil
	return err
}

// Path returns the configured log file path, if known.
func Path() string {
	mu.Lock()
	defer mu.Unlock()
	return currentPath
}

// Initialized reports whether file logging is currently active.
func Initialized() bool {
	mu.Lock()
	defer mu.Unlock()
	return logger != nil
}

// Infof writes an informational log line. It is a safe no-op before initialization.
func Infof(format string, args ...any) {
	printf("INFO", format, args...)
}

// Warnf writes a warning log line. It is a safe no-op before initialization.
func Warnf(format string, args ...any) {
	printf("WARN", format, args...)
}

// Errorf writes an error log line. It is a safe no-op before initialization.
func Errorf(format string, args ...any) {
	printf("ERROR", format, args...)
}

func printf(level, format string, args ...any) {
	mu.Lock()
	l := logger
	mu.Unlock()
	if l == nil {
		return
	}
	l.Printf("%s: %s", level, fmt.Sprintf(format, args...))
}

func defaultPath() (string, error) {
	cacheDir, err := userCacheDir()
	if err != nil {
		return "", fmt.Errorf("resolve user cache directory: %w", err)
	}
	return filepath.Join(cacheDir, "localdev", "localdev.log"), nil
}
