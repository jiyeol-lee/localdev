package view

import (
	"bufio"
	"fmt"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/pkg/command"
	"github.com/jiyeol-lee/localdev/pkg/config"
	"github.com/jiyeol-lee/localdev/pkg/constant"
	"github.com/jiyeol-lee/localdev/pkg/util"
	"github.com/rivo/tview"
)

type Pane struct {
	textView *tview.TextView
	config   config.ConfigPane
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

// runUserCommand executes a user-defined command in a new process and captures its output
func (v *View) runUserCommand(dir string, userCmd string, textView *tview.TextView) error {
	cmd := exec.Command("sh", "-c", userCmd)
	cmd.Dir = dir

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
	for _, pane := range v.panes {
		err := v.runUserCommand(pane.config.Dir, pane.config.Start, pane.textView)
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
			}).ScrollToEnd()
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

		v.panes[index] = &Pane{
			textView: tv,
			config:   configPane,
		}

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

func (v *View) checkIsCommandOutputModalOpen() bool {
	return v.tviewPages.HasPage(constant.Page.CommandOutputModalPage)
}

func (v *View) checkIsCommandHelpModalOpen() bool {
	return v.tviewPages.HasPage(constant.Page.CommandHelpModalPage)
}

func (v *View) removeCommandOutputModal() {
	v.tviewPages.RemovePage(constant.Page.CommandOutputModalPage)
	v.commandOutputModal.reset()
}

func (v *View) removeCommandHelpModal() {
	v.tviewPages.RemovePage(constant.Page.CommandHelpModalPage)
	v.commandHelpModal.reset()
}

func (v *View) openCommandOutputModal() (*tview.InputField, *tview.TextView) {
	textView := tview.NewTextView().ScrollToEnd().SetDynamicColors(true)
	inputField := tview.NewInputField().
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)
	modal := func(p tview.Primitive) tview.Primitive {
		return tview.NewGrid().
			SetColumns(0).
			SetRows(0, 3).
			AddItem(p, 0, 0, 1, 1, 0, 0, false).
			AddItem(inputField, 1, 0, 1, 1, 0, 0, true)
	}
	textView.
		SetBorder(true).
		SetTitle("Command Output").
		SetMouseCapture(
			func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
				if action == tview.MouseScrollUp || action == tview.MouseScrollDown {
					return action, event
				}

				return tview.MouseConsumed, nil
			},
		)
	inputField.
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
				v.runUserCommand(
					v.panes[v.commandOutputModal.callerPaneIndex].config.Dir,
					command,
					v.commandOutputModal.textView,
				)
				inputField.SetText("")
				break

			default:
				v.commandOutputModal.resetCommandHistoryIndex()
				break
			}

			return event
		}).SetBorder(true)
	v.tviewPages.AddPage(constant.Page.CommandOutputModalPage, modal(textView), true, true)

	return inputField, textView
}

func (v *View) openCommandHelpModal() *tview.TextView {
	textView := tview.NewTextView().ScrollToEnd().SetDynamicColors(true)
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
