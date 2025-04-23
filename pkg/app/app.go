package app

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/rivo/tview"
)

type AppView struct {
	textView *tview.TextView
}

// App is the main application structure that holds the configuration and text views.
type App struct {
	tviewApp *tview.Application
	config   *Config
	views    []*AppView
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

// StopPanes stops all panes defined in the configuration.
func (a *App) StopPanes() {
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

	for i, pane := range a.config.Panes {
		pane := pane // capture
		color := colors[i%len(colors)]

		wg.Add(1)
		go func() {
			defer wg.Done()

			cmd := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s", pane.Dir, pane.Stop))
			stdout, err := cmd.StdoutPipe()
			stderr, err2 := cmd.StderrPipe()
			if err != nil {
				fmt.Printf(
					"❌ %sError creating stdout pipe for pane %s: %v%s\n",
					color,
					pane.Name,
					err,
					reset,
				)
			}
			if err2 != nil {
				fmt.Printf(
					"❌ %sError creating stderr pipe for pane %s: %v%s\n",
					color,
					pane.Name,
					err2,
					reset,
				)
			}
			if err != nil || err2 != nil {
				return
			}

			if err := cmd.Start(); err != nil {
				fmt.Printf(
					"❌ %sFailed to start stop command for %s: %v%s\n",
					color,
					pane.Name,
					err,
					reset,
				)
				return
			}

			scanAndPrint := func(r io.ReadCloser) {
				scanner := bufio.NewScanner(r)
				for scanner.Scan() {
					fmt.Printf("%s[%s] %s%s\n", color, pane.Name, scanner.Text(), reset)
				}
			}

			go scanAndPrint(stdout)
			go scanAndPrint(stderr)

			cmd.Wait()
		}()
	}

	wg.Wait()
}
