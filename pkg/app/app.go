package app

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/rivo/tview"
)

// App is the main application structure that holds the configuration and text views.
type App struct {
	tviewApp  *tview.Application
	config    *Config
	textViews []*tview.TextView
}

// Run initializes the application, loads the configuration, and sets up the root view.
func Run() (*App, error) {
	a := &App{
		tviewApp: tview.NewApplication(),
		config:   &Config{},
	}
	a.tviewApp.EnableMouse(true).EnablePaste(true).SetInputCapture(a.keyMapping)

	if err := a.config.loadConfig(); err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
	}

	root := a.getRootView()
	a.tviewApp.SetRoot(root, true)
	if err := a.tviewApp.Run(); err != nil {
		return nil, fmt.Errorf("error running app: %w", err)
	}

	return a, nil
}

// StopSpaces stops all spaces defined in the configuration.
func (a *App) StopSpaces() {
	var wg sync.WaitGroup
	colors := []string{
		"\033[38;2;255;165;0m",   // Orange
		"\033[38;2;255;255;0m",   // Yellow
		"\033[38;2;0;255;0m",     // Green
		"\033[38;2;0;0;255m",     // Blue
		"\033[38;2;128;0;128m",   // Purple
		"\033[38;2;135;206;235m", // Sky Blue
		"\033[38;2;139;69;19m",   // Brown
		"\033[38;2;127;255;212m", // Aqua
		"\033[38;2;75;0;130m",    // Indigo
		"\033[38;2;255;105;180m", // Pink
	}
	reset := "\033[0m"

	for i, space := range a.config.Spaces {
		space := space // capture
		color := colors[i%len(colors)]

		wg.Add(1)
		go func() {
			defer wg.Done()

			cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s", space.Dir, space.Stop))
			stdout, err := cmd.StdoutPipe()
			stderr, err2 := cmd.StderrPipe()
			if err != nil || err2 != nil {
				fmt.Printf(
					"❌ %sError creating pipes for space %s: %v%s\n",
					color,
					space.Name,
					err,
					reset,
				)
				return
			}

			if err := cmd.Start(); err != nil {
				fmt.Printf(
					"❌ %sFailed to start stop command for %s: %v%s\n",
					color,
					space.Name,
					err,
					reset,
				)
				return
			}

			scanAndPrint := func(r io.ReadCloser) {
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					fmt.Printf("%s[%s] %s%s\n", color, space.Name, scanner.Text(), reset)
				}
			}

			go scanAndPrint(stdout)
			go scanAndPrint(stderr)

			cmd.Wait()
		}()
	}

	wg.Wait()
}
