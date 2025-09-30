package env_vars

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/jiyeol-lee/localdev/pkg/internal/shell"
)

// RunCommandAndCaptureEnvVars runs the given command in a shell and captures the environment variables before and after its execution.
// It returns the paths to temporary files containing the environment variable dumps before and after the command execution.
func RunCommandAndCaptureEnvVars(
	command string,
) (beforeCommandEnvVars, AfterCommandEnvVars string, err error) {
	fb, err := os.CreateTemp("", "envdump-before-")
	if err != nil {
		return "", "", err
	}
	fa, err := os.CreateTemp("", "envdump-after-")
	if err != nil {
		return "", "", err
	}
	defer fb.Close()
	defer fa.Close()

	sh := shell.Current()
	// Use portable printenv output (newline separated). BSD printenv (macOS) lacks -0.
	cmdBefore := exec.Command(sh, "-c", "printenv")
	cmdBefore.Stdout = fb
	if err := cmdBefore.Run(); err != nil {
		return "", "", err
	}

	cmdAfter := exec.Command(
		sh,
		"-c",
		fmt.Sprintf(`{ %s; }; printenv > %q`, command, fa.Name()),
	)
	cmdAfter.Stdout = os.Stdout
	cmdAfter.Stderr = os.Stderr
	if err := cmdAfter.Run(); err != nil {
		// Return the before file so caller can still inspect the original env on failure.
		return fb.Name(), "", err
	}

	return fb.Name(), fa.Name(), nil
}

// GetDiffEnvVars compares the environment variable dumps in the given files and returns a slice of strings
func GetDiffEnvVars(
	beforeFile, afterFile string,
) ([]string, error) {
	beforeData, err := os.ReadFile(beforeFile)
	if err != nil {
		return nil, err
	}
	afterData, err := os.ReadFile(afterFile)
	if err != nil {
		return nil, err
	}

	parseEnv := func(b []byte) map[string]string {
		m := make(map[string]string)
		// Normalize any NUL separators to newlines for portability.
		b = bytes.ReplaceAll(b, []byte{0}, []byte{'\n'})
		for line := range bytes.SplitSeq(b, []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			if i := bytes.IndexByte(line, '='); i > 0 {
				m[string(line[:i])] = string(line[i+1:])
			}
		}
		return m
	}

	before := parseEnv(beforeData)
	after := parseEnv(afterData)

	newVars := []string{}
	for k, v := range after {
		if old, ok := before[k]; !ok || old != v {
			newVars = append(newVars, fmt.Sprintf("%s=%s", k, v))
		}
	}

	return newVars, nil
}
