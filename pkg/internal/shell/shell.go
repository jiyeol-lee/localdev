package shell

import "os"

// Current returns the path to the current user's shell.
func Current() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	return shell
}
