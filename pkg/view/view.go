package view

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/pkg/command"
	"github.com/jiyeol-lee/localdev/pkg/config"
	"github.com/jiyeol-lee/localdev/pkg/constant"
	"github.com/jiyeol-lee/localdev/pkg/internal/shell"
	"github.com/jiyeol-lee/localdev/pkg/util"
	"github.com/rivo/tview"
	"golang.org/x/sys/unix"
)

type PanePreview struct {
	Name   string
	Dir    string
	CmdPid int
}

type Pane struct {
	textView *tview.TextView
	config   config.ConfigPane
	cmdPid   int
}

type View struct {
	tviewApp           *tview.Application
	tviewPages         *tview.Pages
	panes              []*Pane
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
	s := make([]int, size)
	for i := range s {
		s[i] = 0
	}
	return s
}

// sanitizeForDisplay removes or escapes ANSI escape sequences from a string for safe display
func sanitizeForDisplay(s string) string {
	// Remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(s, "")
}

// newPane creates a new Pane instance with the provided text view, configuration, and command PID
func newPane(textView *tview.TextView, config config.ConfigPane, cmdPid int) *Pane {
	return &Pane{
		textView: textView,
		config:   config,
		cmdPid:   cmdPid,
	}
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
					fmt.Printf("%sError running command: %s%s\n", constant.AnsiColor.Red, err, constant.AnsiColor.Reset)
				}
				break loop
			case <-sigCh:
				cancel()
				// kill the process group to ensure all child processes are terminated
				if cmd.Process != nil {
					unix.Kill(-cmd.Process.Pid, unix.SIGKILL)
				}
			}
		}

		// When the external command is running, any keystrokes entered by the user are buffered by the terminal.
		// After the command completes and control returns to the Go program, these buffered inputs are immediately consumed by fmt.Scanln,
		// which can cause it to return without waiting for new user input. The flushInput() function clears any buffered input,
		// ensuring that fmt.Scanln waits for fresh input from the user.
		flushInput()

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
func (v *View) runPaneUserCommand(paneIdx int) error {
	pane := v.panes[paneIdx]
	dir := pane.config.Dir
	userCmd := pane.config.Start
	textView := pane.textView
	sh := shell.Current()
	cmd := exec.Command(sh, "-c", userCmd)
	cmd.Dir = dir
	cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}

	stdout, stdoutErr := cmd.StdoutPipe()
	if stdoutErr != nil {
		return fmt.Errorf("error getting stdout pipe: %w", stdoutErr)
	}
	stderr, stderrErr := cmd.StderrPipe()
	if stderrErr != nil {
		return fmt.Errorf("error getting stderr pipe: %w", stderrErr)
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	pane.cmdPid = cmd.Process.Pid

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			t := scanner.Text()
			v.tviewApp.QueueUpdate(func() {
				textView.Write([]byte(t + "\n"))
			})
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			t := scanner.Text()
			v.tviewApp.QueueUpdate(func() {
				textView.Write([]byte("[#8B4513]" + t + "[white]\n"))
			})
		}
	}()

	return nil
}

// getPaneTitle generates the title for each pane in the grid
func getPaneTitle(paneIndex int, configPane config.ConfigPane, focused bool) string {
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

	if focused {
		return fmt.Sprintf("[green][%d] %s[white] - %s", paneIndex+1, configPane.Name, branchInfo)
	}

	return fmt.Sprintf("[%d] %s - %s", paneIndex+1, configPane.Name, branchInfo)
}

func (v *View) Run(configPanes []config.ConfigPane) error {
	v.tviewApp = tview.NewApplication()
	v.tviewApp.EnableMouse(true).EnablePaste(true).SetInputCapture(v.keyMapping)
	v.tviewPages = v.getRootView(configPanes)
	for idx := range v.panes {
		err := v.runPaneUserCommand(idx)
		if err != nil {
			return fmt.Errorf("error running command: %w", err)
		}
	}
	v.tviewApp.SetRoot(v.tviewPages, true)
	v.commandOutputModal = newCommandOutputModal()
	v.commandHelpModal = newCommandHelpModal()
	if err := v.tviewApp.Run(); err != nil {
		return fmt.Errorf("error running app: %w", err)
	}
	return nil
}

func (v *View) GetPanePreviews() []PanePreview {
	pids := make([]PanePreview, len(v.panes))
	for i, pane := range v.panes {
		pids[i] = PanePreview{
			Name:   pane.config.Name,
			Dir:    pane.config.Dir,
			CmdPid: pane.cmdPid,
		}
	}
	return pids
}

func (v *View) getRootView(configPanes []config.ConfigPane) *tview.Pages {
	root := tview.NewPages()
	l := len(configPanes)
	v.panes = make([]*Pane, l)
	rows, cols := getGridDimensions(l)
	grid := tview.NewGrid()
	grid.
		SetRows(makeFlexibleSlice(rows)...).
		SetColumns(makeFlexibleSlice(cols)...)
	row := 0
	col := 0
	for index, configPane := range configPanes {
		tv := tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetChangedFunc(func() {
				v.tviewApp.Draw()
			}).ScrollToEnd().SetMaxLines(constant.MaxPaneOutputLines)
		tv.
			SetBorder(true).
			SetTitle(getPaneTitle(index, configPane, tv.HasFocus()))

		tv.SetBlurFunc(func() {
			tv.SetBorderColor(tcell.ColorWhite).
				SetTitle(getPaneTitle(index, configPane, false))
		})
		tv.SetFocusFunc(func() {
			tv.SetBorderColor(tcell.ColorGreen).
				SetTitle(getPaneTitle(index, configPane, true))
		})

		v.panes[index] = newPane(
			tv, configPane, 0,
		)

		grid.AddItem(tv, row, col, 1, 1, 0, 0, true)
		if row == 1 {
			row = 0
			col++
		} else {
			row++
		}
	}
	root.AddPage(constant.Page.MainPage, grid, true, true)

	return root
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
	tv.Write([]byte(fmt.Appendf(nil, "\n  [orange]===%s===[-]\n\n", "Local")))
	tv.Write([]byte(fmt.Appendf(nil, "  [lightgreen]Silent[-] command\n")))
	tv.Write([]byte(fmt.Appendf(nil, "  [green]Normal[-] command\n\n")))

	paneCommands := v.panes[v.commandHelpModal.callerPaneIndex].config.Commands
	if paneCommands == nil {
		tv.Write(fmt.Appendf(nil, "  No commands available\n"))
		return
	}
	paneCommandsMap, err := util.YamlToMap[*config.ConfigCommands, *config.ConfigCommand](
		paneCommands,
	)
	if err != nil {
		tv.Write(fmt.Appendf(nil, "  [red]Error[white]: %s\n", err))
		return
	}

	for key, configCommand := range paneCommandsMap {
		if configCommand == nil {
			continue
		}
		c, err := convertCommandKeyToCharacter(key)
		if err != nil {
			tv.Write(fmt.Appendf(nil, "  [red]Error[white]: %s\n", err))
			continue
		}
		if configCommand.Silent {
			tv.Write(
				fmt.Appendf(nil, "  [lightgreen]%s[white] %s\n", c, configCommand.Description),
			)
		} else {
			tv.Write(fmt.Appendf(nil, "  [green]%s[white] %s\n", c, configCommand.Description))
		}
	}
}
