package view

import (
	"fmt"
	"os"
	"strings"
	"syscall"

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
func flushInput() {
	fd := int(os.Stdin.Fd())

	// Get current terminal attributes
	oldState, err := term.GetState(fd)
	if err != nil {
		return
	}

	// Set terminal to raw mode temporarily
	rawState, err := term.MakeRaw(fd)
	if err != nil {
		return
	}

	// Restore terminal state when done
	defer term.Restore(fd, oldState)

	// Set non-blocking mode and read/discard buffered input
	err = syscall.SetNonblock(fd, true)
	if err != nil {
		term.Restore(fd, rawState) // Restore raw state first
		return
	}

	// Read and discard all available input
	buffer := make([]byte, 1024)
	for {
		_, err := syscall.Read(fd, buffer)
		if err != nil {
			break // No more data to read
		}
	}

	// Restore blocking mode
	syscall.SetNonblock(fd, false)
}
