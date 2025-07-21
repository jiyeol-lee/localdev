package view

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// convertCommandKeyToCharacter converts a key string to its corresponding character.
func convertCommandKeyToCharacter(key string) (string, error) {
	if strings.HasPrefix(key, "lower") && len(key) == 6 {
		return strings.ToLower(string(key[5])), nil
	}

	if strings.HasPrefix(key, "upper") && len(key) == 6 {
		return strings.ToUpper(string(key[5])), nil
	}

	return "", fmt.Errorf("invalid key format: %s", key)
}

// flushInput flushes any buffered input from the terminal.
func flushInput() error {
	fd := int(os.Stdin.Fd())

	// Get current terminal attributes
	oldState, err := term.GetState(fd)
	if err != nil {
		return fmt.Errorf("failed to get terminal state: %w", err)
	}

	// Set terminal to raw mode temporarily
	_, err = term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to set terminal to raw mode: %w", err)
	}

	// Restore terminal state when done
	defer term.Restore(fd, oldState)

	// Set non-blocking mode and read/discard buffered input
	err = unix.SetNonblock(fd, true)
	if err != nil {
		return fmt.Errorf("failed to set non-blocking mode: %w", err)
	}

	// Read and discard all available input
	buffer := make([]byte, 1024)
	for {
		_, err := unix.Read(fd, buffer)
		if err != nil {
			break // No more data to read
		}
	}

	// Restore blocking mode
	err = unix.SetNonblock(fd, false)
	if err != nil {
		return fmt.Errorf("failed to restore blocking mode: %w", err)
	}

	return nil
}
