package view

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"sync"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/internal/logger"
	"github.com/jiyeol-lee/localdev/pkg/command"
	"github.com/jiyeol-lee/localdev/pkg/config"
	"github.com/jiyeol-lee/localdev/pkg/constant"
	"github.com/jiyeol-lee/localdev/pkg/internal/env_vars"
	"github.com/jiyeol-lee/localdev/pkg/internal/shell"
	"github.com/jiyeol-lee/localdev/pkg/util"
	"github.com/rivo/tview"
	"golang.org/x/sys/unix"
)

// Pane represents a single terminal pane with its running process and UI component.
type Pane struct {
	textView     *tview.TextView
	config       config.ConfigPane
	mu           sync.Mutex
	cmd          *exec.Cmd
	generation   int
	stopExecuted bool

	expectedStopGenerations map[int]bool
}

func (p *Pane) markExpectedStop(gen int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.expectedStopGenerations == nil {
		p.expectedStopGenerations = make(map[int]bool)
	}
	p.expectedStopGenerations[gen] = true
}

func (p *Pane) isExpectedStop(gen int) bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.expectedStopGenerations[gen]
}

func (p *Pane) clearExpectedStop(gen int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.expectedStopGenerations, gen)
}

// IsRunning reports whether the pane's start process is currently running.
func (p *Pane) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.cmd == nil || p.cmd.Process == nil {
		return false
	}
	if p.cmd.ProcessState != nil {
		return false
	}
	return unix.Kill(-p.cmd.Process.Pid, 0) == nil
}

// View manages the terminal UI, panes, and user interactions.
type View struct {
	tviewApp           *tview.Application
	tviewPages         *tview.Pages
	panes              []*Pane
	envVars            []string
	commandOutputModal *commandOutputModal
	commandHelpModal   *commandHelpModal
}

// getGridDimensions calculates the number of rows and columns for the grid layout
func getGridDimensions(length int) (rows, cols int) {
	switch {
	case length <= 0:
		return 0, 0
	case length == 1:
		return 1, 1
	case length == 2:
		return 1, 2
	case length <= 4:
		return 2, 2
	case length <= 6:
		return 2, 3
	case length <= 8:
		return 2, 4
	default:
		return 2, 5
	}
}

// makeFlexibleSlice creates a slice of integers with the specified size and initializes all elements to 0
func makeFlexibleSlice(size int) []int {
	return make([]int, size)
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// sanitizeForDisplay removes or escapes ANSI escape sequences from a string for safe display
func sanitizeForDisplay(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func (v *View) runCustomUserCommand(dir string, userCmd string) {
	v.tviewApp.Suspend(func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, unix.SIGINT)
		defer func() {
			signal.Stop(sigCh)
			close(sigCh)
		}()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		sh := shell.Current()
		cmd := exec.CommandContext(ctx, sh, "-c", userCmd)
		cmd.Env = append(os.Environ(), v.envVars...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}

		// Sanitize the command for safe display
		sanitizedCmd := sanitizeForDisplay(userCmd)

		fmt.Printf(
			"\n%s+%s Executing command from %s%s%s\n",
			constant.AnsiColor.Green,
			constant.AnsiColor.Reset,
			constant.AnsiColor.Green,
			v.panes[v.commandOutputModal.callerPaneIndex].config.Name,
			constant.AnsiColor.Reset,
		)
		fmt.Printf(
			"%s+%s Command is executed in %s%s%s\n",
			constant.AnsiColor.Green,
			constant.AnsiColor.Reset,
			constant.AnsiColor.Green,
			v.panes[v.commandOutputModal.callerPaneIndex].config.Dir,
			constant.AnsiColor.Reset,
		)
		fmt.Printf(
			"%s+ %s -c %s%s\n\n",
			constant.AnsiColor.Green,
			sh,
			sanitizedCmd, // Use sanitized version for display
			constant.AnsiColor.Reset,
		)
		if err := cmd.Start(); err != nil {
			logger.Errorf(
				"error starting suspended custom command for pane %s: %v",
				v.panes[v.commandOutputModal.callerPaneIndex].config.Name,
				err,
			)
			fmt.Printf(
				"%sError starting command: %s%s\n",
				constant.AnsiColor.Red,
				err,
				constant.AnsiColor.Reset,
			)
			return
		}

		doneCh := make(chan error, 1)
		go func() { doneCh <- cmd.Wait() }()
	loop:
		for {
			select {
			case err := <-doneCh:
				// Check if process was killed by signal
				if err != nil {
					if cmd.ProcessState != nil {
						if status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok {
							// If killed by SIGKILL, suppress error message
							if status.Signaled() && status.Signal() == unix.SIGKILL {
								break loop
							}
						}
					}
					logger.Errorf("suspended custom command for pane %s exited with error: %v", v.panes[v.commandOutputModal.callerPaneIndex].config.Name, err)
					fmt.Printf("%sError running command: %s%s\n", constant.AnsiColor.Red, err, constant.AnsiColor.Reset)
				}
				break loop
			case <-sigCh:
				cancel()
				// kill the process group to ensure all child processes are terminated
				if cmd.Process != nil {
					pid := cmd.Process.Pid
					if err := unix.Kill(-pid, unix.SIGKILL); err != nil && err != syscall.ESRCH {
						logger.Warnf("failed to send SIGKILL during suspended command cancellation for pane %s (pid=%d, pgid=%d): %v", v.panes[v.commandOutputModal.callerPaneIndex].config.Name, pid, -pid, err)
					}
				}
			}
		}

		// When the external command is running, any keystrokes entered by the user are buffered by the terminal.
		// After the command completes and control returns to the Go program, these buffered inputs are immediately consumed by fmt.Scanln,
		// which can cause it to return without waiting for new user input. The flushInput() function clears any buffered input,
		// ensuring that fmt.Scanln waits for fresh input from the user.
		if err := flushInput(); err != nil {
			logger.Warnf("failed to flush terminal input after suspended custom command: %v", err)
		}

		// Wait for user input after command completes
		fmt.Printf(
			"\n%sPress Enter to return to the app...%s",
			constant.AnsiColor.Green,
			constant.AnsiColor.Reset,
		)
		var input string
		for {
			if _, err := fmt.Scanln(&input); err != nil {
				// If the user presses Enter without typing anything, we can break the loop
				if input == "" {
					fmt.Printf(
						"\n%s+ Returning to the app...%s\n",
						constant.AnsiColor.Green,
						constant.AnsiColor.Reset,
					)
					break
				}
			}
		}
	})
}

// runPaneUserCommand executes a user-defined command in a new process and captures its output
func (v *View) runPaneUserCommand(pane *Pane, generation int) (*exec.Cmd, error) {
	sh := shell.Current()
	cmd := exec.Command(sh, "-c", pane.config.Start)
	cmd.Env = append(os.Environ(), v.envVars...)
	cmd.Dir = pane.config.Dir
	cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}

	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		return nil, fmt.Errorf("error getting stdout pipe: %w", stdoutErr)
	}
	stderr, stderrErr := cmd.StderrPipe()
	if stderrErr != nil {
		return nil, fmt.Errorf("error getting stderr pipe: %w", stderrErr)
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	go func(gen int) {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			pane.mu.Lock()
			currentGen := pane.generation
			pane.mu.Unlock()
			if gen != currentGen {
				return
			}
			t := scanner.Text()
			v.tviewApp.QueueUpdate(func() {
				_, _ = pane.textView.Write([]byte(t + "\n"))
			})
		}
		if err := scanner.Err(); err != nil {
			logger.Errorf(
				"error reading stdout for pane %s during start command: %v",
				pane.config.Name,
				err,
			)
		}
	}(generation)

	go func(gen int) {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			pane.mu.Lock()
			currentGen := pane.generation
			pane.mu.Unlock()
			if gen != currentGen {
				return
			}
			t := scanner.Text()
			v.tviewApp.QueueUpdate(func() {
				_, _ = pane.textView.Write([]byte("[#8B4513]" + t + "[white]\n"))
			})
		}
		if err := scanner.Err(); err != nil {
			logger.Errorf(
				"error reading stderr for pane %s during start command: %v",
				pane.config.Name,
				err,
			)
		}
	}(generation)

	return cmd, nil
}

func (v *View) handlePaneCommandWaitError(pane *Pane, phase string, generation int, err error) {
	if err == nil {
		return
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.Signaled() &&
			(status.Signal() == unix.SIGINT || status.Signal() == unix.SIGKILL) &&
			pane.isExpectedStop(generation) {
			// Silently ignore expected SIGINT/SIGKILL for this process generation.
			return
		}
	}
	logger.Errorf("pane %s %s command exited with error: %v", pane.config.Name, phase, err)
	v.tviewApp.QueueUpdate(func() {
		_, _ = fmt.Fprintf(
			pane.textView,
			"[red]Pane %s %s command exited with error: %v[-]\n",
			pane.config.Name,
			phase,
			err,
		)
	})
}

// getPaneTitle generates the title for each pane in the grid
func getPaneTitle(
	paneIndex int,
	configPane config.ConfigPane,
	focused bool,
	isRunning bool,
) string {
	branch, err := command.GetCurrentBranch(configPane.Dir)
	branchInfo := branch
	// if git is not initialized, it will return an error
	if err != nil {
		branchInfo = "N/A"
	}
	status, err := command.GetBranchSyncStatus(configPane.Dir)
	// if git is not pushed to remote, it will return an error
	if err == nil {
		branchInfo += fmt.Sprintf(
			" [yellow]↑%d[white] [yellow]↓%d[white]",
			status.Ahead,
			status.Behind,
		)
	}

	statusIndicator := ""
	if isRunning {
		statusIndicator = "[green]●[white] "
	} else {
		statusIndicator = "[red]●[white] "
	}

	if focused {
		return fmt.Sprintf(
			"[green][%d] %s[green]%s[white] - %s",
			paneIndex+1,
			statusIndicator,
			configPane.Name,
			branchInfo,
		)
	}

	return fmt.Sprintf("[%d] %s%s - %s", paneIndex+1, statusIndicator, configPane.Name, branchInfo)
}

// GetEnvVars returns the environment variables captured after running the project command.
func (v *View) GetEnvVars() []string {
	return v.envVars
}

// Run initializes and starts the terminal UI with the given configuration.
func (v *View) Run(config config.Config) error {
	v.tviewApp = tview.NewApplication()
	v.tviewApp.EnableMouse(true).EnablePaste(true).SetInputCapture(v.keyMapping)
	v.tviewPages, v.panes = v.getRootView(config)
	projectCmd := config.GetProjectCommand()
	if projectCmd != "" {
		beforeCommandEnvVars, afterCommandEnvVars, err := env_vars.RunCommandAndCaptureEnvVars(
			projectCmd,
		)
		defer os.Remove(beforeCommandEnvVars)
		defer os.Remove(afterCommandEnvVars)
		if err != nil {
			return fmt.Errorf("error running project command: %w", err)
		}
		v.envVars, err = env_vars.GetDiffEnvVars(beforeCommandEnvVars, afterCommandEnvVars)
		if err != nil {
			return fmt.Errorf("error getting env vars diff: %w", err)
		}
	}
	for i, pane := range v.panes {
		pane.mu.Lock()
		pane.generation++
		gen := pane.generation
		pane.mu.Unlock()

		cmd, err := v.runPaneUserCommand(pane, gen)
		if err != nil {
			logger.Errorf(
				"error running initial start command for pane %s: %v",
				pane.config.Name,
				err,
			)
			return fmt.Errorf("error running command: %w", err)
		}

		pane.mu.Lock()
		pane.cmd = cmd
		pane.mu.Unlock()

		go func(p *Pane, c *exec.Cmd, gen int, idx int) {
			defer p.clearExpectedStop(gen)
			if err := c.Wait(); err != nil {
				v.handlePaneCommandWaitError(p, "start", gen, err)
			}
			p.mu.Lock()
			if p.cmd == c {
				p.cmd = nil
			}
			p.mu.Unlock()
			v.tviewApp.QueueUpdate(func() {
				v.updatePaneTitle(idx)
			})
		}(pane, cmd, gen, i)

		v.updatePaneTitle(i)
	}
	v.tviewApp.SetRoot(v.tviewPages, true)
	v.commandOutputModal = newCommandOutputModal()
	v.commandHelpModal = newCommandHelpModal()
	if err := v.tviewApp.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}
	return nil
}

func (v *View) getRootView(config config.Config) (*tview.Pages, []*Pane) {
	root := tview.NewPages()
	l := len(config.Panes)
	panes := make([]*Pane, l)
	rows, cols := getGridDimensions(l)
	grid := tview.NewGrid()
	grid.
		SetRows(makeFlexibleSlice(rows)...).
		SetColumns(makeFlexibleSlice(cols)...)
	row := 0
	col := 0
	for index, configPane := range config.Panes {
		projectDir := config.GetProjectDir()
		if projectDir != "" {
			configPane.Dir = filepath.Join(projectDir, configPane.Dir)
		}

		tv := tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetChangedFunc(func() {
				v.tviewApp.Draw()
			}).ScrollToEnd().SetMaxLines(constant.MaxPaneOutputLines)
		tv.
			SetBorder(true).
			SetTitle(getPaneTitle(index, configPane, tv.HasFocus(), false))

		panes[index] = &Pane{
			textView: tv,
			config:   configPane,
		}
		paneRef := panes[index]

		tv.SetBlurFunc(func() {
			tv.SetBorderColor(tcell.ColorWhite).
				SetTitle(getPaneTitle(index, configPane, false, paneRef.IsRunning()))
		})
		tv.SetFocusFunc(func() {
			tv.SetBorderColor(tcell.ColorGreen).
				SetTitle(getPaneTitle(index, configPane, true, paneRef.IsRunning()))
		})

		grid.AddItem(tv, row, col, 1, 1, 0, 0, true)
		if row == 1 {
			row = 0
			col++
		} else {
			row++
		}
	}
	root.AddPage(constant.Page.MainPage, grid, true, true)

	return root, panes
}

func (v *View) disablePanesMouse() {
	for _, pane := range v.panes {
		pane.textView.SetMouseCapture(
			func(_ tview.MouseAction, _ *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
				return tview.MouseConsumed, nil
			},
		)
	}
}

func (v *View) enablePanesMouse() {
	for _, pane := range v.panes {
		pane.textView.SetMouseCapture(nil)
	}
}

func (v *View) checkIsCommandOutputModalOpen() bool {
	return v.tviewPages.HasPage(constant.Page.CommandOutputModalPage)
}

func (v *View) checkIsCommandHelpModalOpen() bool {
	return v.tviewPages.HasPage(constant.Page.CommandHelpModalPage)
}

func (v *View) removeCommandOutputModal() {
	v.tviewPages.RemovePage(constant.Page.CommandOutputModalPage)
	v.commandOutputModal.reset()
	v.enablePanesMouse()
}

func (v *View) removeCommandHelpModal() {
	v.tviewPages.RemovePage(constant.Page.CommandHelpModalPage)
	v.commandHelpModal.reset()
	v.enablePanesMouse()
}

func (v *View) openCommandOutputModal() *tview.InputField {
	inputField := tview.NewInputField().
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)
	inputField.SetTitle("Command to run").
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				callerPaneTextView := v.panes[v.commandOutputModal.callerPaneIndex].textView
				v.removeCommandOutputModal()
				v.tviewApp.SetFocus(callerPaneTextView)
				return nil

			case tcell.KeyUp:
				pc := v.commandOutputModal.previousCommand()
				if pc != "" {
					inputField.SetText(pc)
				}
				return nil

			case tcell.KeyDown:
				nc := v.commandOutputModal.nextCommand()
				if nc != "" {
					inputField.SetText(nc)
				}
				return nil

			case tcell.KeyLeft:
			case tcell.KeyRight:
				break

			case tcell.KeyEnter:
				command := inputField.GetText()
				v.commandOutputModal.appendCommandHistory(command)
				v.commandOutputModal.resetCommandHistoryIndex()
				v.runCustomUserCommand(
					v.panes[v.commandOutputModal.callerPaneIndex].config.Dir,
					command,
				)
				inputField.SetText("")

			default:
				v.commandOutputModal.resetCommandHistoryIndex()
			}

			return event
		}).SetBorder(true)
	modal := func(p tview.Primitive) *tview.Grid {
		g := tview.NewGrid()
		g.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			if width > 90 {
				g.SetColumns(0, 80, 0)
			} else {
				g.SetColumns(2, 0, 2)
			}
			return x, y, width, height
		})

		return g.
			SetColumns(2, 0, 2).
			SetRows(5, 3, 0).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}
	v.tviewPages.AddPage(constant.Page.CommandOutputModalPage, modal(inputField), true, true)

	return inputField
}

func (v *View) openCommandHelpModal() *tview.TextView {
	textView := tview.NewTextView().ScrollToEnd().SetDynamicColors(true)
	textView.
		SetBorder(true).
		SetTitle("Command Help").
		SetMouseCapture(
			func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
				if action == tview.MouseScrollUp || action == tview.MouseScrollDown {
					return action, event
				}

				return tview.MouseConsumed, nil
			},
		).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyEsc:
				callerPaneTextView := v.panes[v.commandHelpModal.callerPaneIndex].textView
				v.removeCommandHelpModal()
				v.tviewApp.SetFocus(callerPaneTextView)
			}
			return nil
		})
	modal := func(p tview.Primitive) *tview.Grid {
		g := tview.NewGrid()
		g.SetDrawFunc(func(screen tcell.Screen, x, y, width, height int) (int, int, int, int) {
			if width > 90 {
				g.SetColumns(0, 80, 0)
			} else {
				g.SetColumns(2, 0, 2)
			}
			return x, y, width, height
		})

		return g.
			SetRows(2, 0, 2).
			AddItem(p, 1, 1, 1, 1, 0, 0, true)
	}

	v.tviewPages.AddPage(constant.Page.CommandHelpModalPage, modal(textView), true, true)

	return textView
}

func (v *View) setCommandHelpModalBodyText() {
	tv := v.commandHelpModal.textView

	tv.Clear()
	_, _ = tv.Write(fmt.Appendf(nil, "\n  [orange]===%s===[-]\n\n", "Local"))
	_, _ = tv.Write(fmt.Appendf(nil, "  [lightgreen]Silent[-] command\n"))
	_, _ = tv.Write(fmt.Appendf(nil, "  [green]Normal[-] command\n\n"))

	paneCommands := v.panes[v.commandHelpModal.callerPaneIndex].config.Commands
	if paneCommands == nil {
		_, _ = tv.Write(fmt.Appendf(nil, "  No commands available\n"))
		return
	}
	paneCommandsMap, err := util.YamlToMap[*config.ConfigCommands, *config.ConfigCommand](
		paneCommands,
	)
	if err != nil {
		logger.Errorf(
			"error rendering command help for pane %s: %v",
			v.panes[v.commandHelpModal.callerPaneIndex].config.Name,
			err,
		)
		_, _ = tv.Write(fmt.Appendf(nil, "  [red]Error[white]: %s\n", err))
		return
	}

	// Print non-reserved commands first
	for key, configCommand := range paneCommandsMap {
		if configCommand == nil {
			continue
		}
		if configCommand.Command == constant.ReservedCommand.TogglePaneSize ||
			configCommand.Command == constant.ReservedCommand.StartPane ||
			configCommand.Command == constant.ReservedCommand.StopPane {
			continue
		}
		c, err := convertCommandKeyToCharacter(key)
		if err != nil {
			logger.Errorf(
				"error rendering command help key %q for pane %s: %v",
				key,
				v.panes[v.commandHelpModal.callerPaneIndex].config.Name,
				err,
			)
			_, _ = tv.Write(fmt.Appendf(nil, "  [red]Error[white]: %s\n", err))
			continue
		}
		if configCommand.Silent {
			_, _ = tv.Write(
				fmt.Appendf(nil, "  [lightgreen]%s[white] %s\n", c, configCommand.Description),
			)
		} else {
			_, _ = tv.Write(fmt.Appendf(nil, "  [green]%s[white] %s\n", c, configCommand.Description))
		}
	}

	// Show reserved commands section if any are bound
	var reservedKeys []string
	for key, configCommand := range paneCommandsMap {
		if configCommand == nil {
			continue
		}
		if configCommand.Command == constant.ReservedCommand.TogglePaneSize ||
			configCommand.Command == constant.ReservedCommand.StartPane ||
			configCommand.Command == constant.ReservedCommand.StopPane {
			reservedKeys = append(reservedKeys, key)
		}
	}

	if len(reservedKeys) > 0 {
		_, _ = tv.Write(fmt.Appendf(nil, "\n  [orange]===%s===[-]\n\n", "Reserved"))

		for _, key := range reservedKeys {
			configCommand := paneCommandsMap[key]
			c, err := convertCommandKeyToCharacter(key)
			if err != nil {
				logger.Errorf(
					"error rendering reserved command help key %q for pane %s: %v",
					key,
					v.panes[v.commandHelpModal.callerPaneIndex].config.Name,
					err,
				)
				_, _ = tv.Write(fmt.Appendf(nil, "  [red]Error[white]: %s\n", err))
				continue
			}
			_, _ = tv.Write(
				fmt.Appendf(nil, "  [orange]%s[white] %s\n", c, configCommand.Description),
			)
		}
	}
}

// togglePaneSize maximizes the currently focused pane or restores it back to the grid view
func (v *View) togglePaneSize() {
	focusedPaneIndex := v.focusedViewIndex()
	if focusedPaneIndex == -1 {
		return
	}

	fpName, _ := v.tviewPages.GetFrontPage()
	if fpName == constant.Page.MainPage {
		v.tviewPages.AddAndSwitchToPage(
			constant.Page.MaximizedPane,
			v.panes[focusedPaneIndex].textView,
			true,
		)
	} else {
		v.tviewPages.SwitchToPage(constant.Page.MainPage)
		v.tviewPages.RemovePage(constant.Page.MaximizedPane)
		v.tviewApp.SetFocus(v.panes[focusedPaneIndex].textView)
	}
}

// checkIsPaneMaximized checks if any pane is currently maximized
func (v *View) checkIsPaneMaximized() bool {
	fpName, _ := v.tviewPages.GetFrontPage()
	return fpName == constant.Page.MaximizedPane
}

func (v *View) startPane(index int) {
	p := v.panes[index]

	go func() {
		p.mu.Lock()
		oldCmd := p.cmd
		oldGen := p.generation
		p.generation++
		gen := p.generation
		p.stopExecuted = false
		p.mu.Unlock()

		v.tviewApp.QueueUpdate(func() {
			_, _ = fmt.Fprintf(
				p.textView,
				"\n[gray]━━━ Started at %s ━━━[-]\n\n",
				time.Now().Format("15:04:05"),
			)
		})

		if oldCmd != nil && oldCmd.Process != nil {
			p.markExpectedStop(oldGen)
			pid := oldCmd.Process.Pid
			if err := unix.Kill(-pid, unix.SIGKILL); err != nil && err != syscall.ESRCH {
				logger.Warnf(
					"failed to send SIGKILL during restart cleanup for pane %s (pid=%d, pgid=%d): %v",
					p.config.Name,
					pid,
					-pid,
					err,
				)
			}
			for range 50 {
				if err := unix.Kill(-pid, 0); err == syscall.ESRCH {
					break
				}
				time.Sleep(100 * time.Millisecond)
			}
		}

		newCmd, err := v.runPaneUserCommand(p, gen)
		if err != nil {
			logger.Errorf("error restarting start command for pane %s: %v", p.config.Name, err)
			v.tviewApp.QueueUpdate(func() {
				_, _ = fmt.Fprintf(p.textView, "[red]Failed to start: %s[-]\n", err)
			})
			v.tviewApp.QueueUpdate(func() {
				v.updatePaneTitle(index)
			})
			return
		}

		p.mu.Lock()
		p.cmd = newCmd
		p.mu.Unlock()

		go func(p *Pane, c *exec.Cmd, gen int) {
			defer p.clearExpectedStop(gen)
			if err := c.Wait(); err != nil {
				v.handlePaneCommandWaitError(p, "start", gen, err)
			}
			p.mu.Lock()
			if p.cmd == c {
				p.cmd = nil
			}
			p.mu.Unlock()
			v.tviewApp.QueueUpdate(func() {
				v.updatePaneTitle(index)
			})
		}(p, newCmd, gen)

		v.tviewApp.QueueUpdate(func() {
			v.updatePaneTitle(index)
		})
	}()
}

func (v *View) stopPane(index int) {
	p := v.panes[index]

	go func() {
		p.mu.Lock()
		cmd := p.cmd
		gen := p.generation
		p.mu.Unlock()

		v.tviewApp.QueueUpdate(func() {
			_, _ = fmt.Fprintf(
				p.textView,
				"\n[gray]━━━ Stopping... %s ━━━[-]\n\n",
				time.Now().Format("15:04:05"),
			)
		})

		if cmd != nil && cmd.Process != nil {
			p.markExpectedStop(gen)
			pid := cmd.Process.Pid
			if err := unix.Kill(-pid, unix.SIGINT); err != nil && err != syscall.ESRCH {
				logger.Warnf(
					"failed to send SIGINT to process group for pane %s (pid=%d, pgid=%d): %v",
					p.config.Name,
					pid,
					-pid,
					err,
				)
			}

			exited := false
			for range 30 {
				if err := unix.Kill(-pid, 0); err == syscall.ESRCH {
					exited = true
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if !exited {
				if err := unix.Kill(-pid, unix.SIGKILL); err != nil && err != syscall.ESRCH {
					logger.Warnf(
						"failed to send SIGKILL to process group for pane %s (pid=%d, pgid=%d): %v",
						p.config.Name,
						pid,
						-pid,
						err,
					)
				}
				for range 20 {
					if err := unix.Kill(-pid, 0); err == syscall.ESRCH {
						break
					}
					time.Sleep(100 * time.Millisecond)
				}
			}
		}

		stopErrorCount := v.runPaneCommandToTextView(
			p.config.Name,
			p.config.Dir,
			p.config.Stop,
			p.textView,
		)

		p.mu.Lock()
		p.stopExecuted = stopErrorCount == 0
		p.mu.Unlock()

		v.tviewApp.QueueUpdate(func() {
			if stopErrorCount == 0 {
				_, _ = fmt.Fprintf(
					p.textView,
					"\n[gray]━━━ Stopped at %s ━━━[-]\n\n",
					time.Now().Format("15:04:05"),
				)
			} else {
				_, _ = fmt.Fprintf(p.textView, "\n[red]━━━ Stop command failed at %s (%d error(s)); final shutdown will retry cleanup ━━━[-]\n\n", time.Now().Format("15:04:05"), stopErrorCount)
			}
		})

		v.tviewApp.QueueUpdate(func() {
			v.updatePaneTitle(index)
		})
	}()
}

func (v *View) runPaneCommandToTextView(
	paneName, dir, userCmd string,
	textView *tview.TextView,
) int {
	var errorMu sync.Mutex
	errorCount := 0
	recordError := func() {
		errorMu.Lock()
		defer errorMu.Unlock()
		errorCount++
	}

	sh := shell.Current()
	cmd := exec.Command(sh, "-c", userCmd)
	cmd.Env = append(os.Environ(), v.envVars...)
	cmd.Dir = dir

	stdout, err1 := cmd.StdoutPipe()
	stderr, err2 := cmd.StderrPipe()
	if err1 != nil || err2 != nil {
		if err1 != nil {
			recordError()
		}
		if err2 != nil {
			recordError()
		}
		displayErr := err1
		if displayErr == nil {
			displayErr = err2
		}
		if err1 != nil {
			logger.Errorf(
				"error creating stdout pipe for stop command for pane %s: %v",
				paneName,
				err1,
			)
		}
		if err2 != nil {
			logger.Errorf(
				"error creating stderr pipe for stop command for pane %s: %v",
				paneName,
				err2,
			)
		}
		v.tviewApp.QueueUpdate(func() {
			_, _ = fmt.Fprintf(textView, "[red]Error piping stop command: %v[-]\n", displayErr)
		})
		return errorCount
	}

	if err := cmd.Start(); err != nil {
		recordError()
		logger.Errorf("error starting stop command for pane %s: %v", paneName, err)
		v.tviewApp.QueueUpdate(func() {
			_, _ = fmt.Fprintf(textView, "[red]Error starting stop command: %v[-]\n", err)
		})
		return errorCount
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			t := scanner.Text()
			v.tviewApp.QueueUpdate(func() {
				_, _ = textView.Write([]byte(t + "\n"))
			})
		}
		if err := scanner.Err(); err != nil {
			recordError()
			logger.Errorf("error reading stdout for pane %s during stop command: %v", paneName, err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			t := scanner.Text()
			v.tviewApp.QueueUpdate(func() {
				_, _ = textView.Write([]byte("[#8B4513]" + t + "[white]\n"))
			})
		}
		if err := scanner.Err(); err != nil {
			recordError()
			logger.Errorf("error reading stderr for pane %s during stop command: %v", paneName, err)
		}
	}()

	if err := cmd.Wait(); err != nil {
		recordError()
		logger.Errorf("stop command for pane %s exited with error: %v", paneName, err)
		v.tviewApp.QueueUpdate(func() {
			_, _ = fmt.Fprintf(textView, "[red]Pane command exited with error: %v[-]\n", err)
		})
	}
	wg.Wait()
	return errorCount
}

// GetManuallyStoppedPaneNames returns a set of pane names that were manually stopped.
func (v *View) GetManuallyStoppedPaneNames() map[string]bool {
	result := make(map[string]bool)
	for _, p := range v.panes {
		p.mu.Lock()
		if p.stopExecuted {
			result[p.config.Name] = true
		}
		p.mu.Unlock()
	}
	return result
}

func (v *View) updatePaneTitle(index int) {
	if index < 0 || index >= len(v.panes) {
		return
	}
	p := v.panes[index]
	focused := p.textView.HasFocus()
	p.textView.SetTitle(getPaneTitle(index, p.config, focused, p.IsRunning()))
}
