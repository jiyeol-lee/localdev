package view

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/pkg/command"
	"github.com/jiyeol-lee/localdev/pkg/config"
	"github.com/jiyeol-lee/localdev/pkg/constant"
	"github.com/rivo/tview"
)

type Pane struct {
	textView *tview.TextView
}

type View struct {
	tviewApp   *tview.Application
	tviewPages *tview.Pages
	panes      []*Pane
}

func (v *View) EnableMouse(enable bool) {
	v.tviewApp.EnableMouse(enable)
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
func (v *View) runUserCommand(dir string, userCmd string, textView *tview.TextView) {
	cmd := exec.Command("sh", "-c", userCmd)
	cmd.Dir = dir

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Panicln("Error starting command:", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			t := scanner.Text()
			v.tviewApp.QueueUpdateDraw(func() {
				textView.Write([]byte(t + "\n"))
			})
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			t := scanner.Text()
			v.tviewApp.QueueUpdateDraw(func() {
				textView.Write([]byte("[#8B4513]" + t + "[white]\n"))
			})
		}
	}()
}

// getPaneTitle generates the title for each pane in the grid
func getPaneTitle(paneIndex int, pane config.ConfigPane, focused bool) string {
	branch, err := command.GetCurrentBranch(pane.Dir)
	branchInfo := branch
	// if git is not initialized, it will return an error
	if err != nil {
		branchInfo = "N/A"
	}
	status, err := command.GetBranchSyncStatus(pane.Dir)
	// if git is not pushed to remote, it will return an error
	if err == nil {
		branchInfo += fmt.Sprintf(
			" [yellow]↑%d[white] [yellow]↓%d[white]",
			status.Ahead,
			status.Behind,
		)
	}

	if focused {
		return fmt.Sprintf("[green][%d] %s[white] - %s", paneIndex+1, pane.Name, branchInfo)
	}

	return fmt.Sprintf("[%d] %s - %s", paneIndex+1, pane.Name, branchInfo)
}

func (v *View) Run(configPanes []config.ConfigPane) error {
	v.tviewApp = tview.NewApplication()
	v.tviewApp.EnableMouse(true).EnablePaste(true).SetInputCapture(v.keyMapping(configPanes))
	v.tviewPages = v.getRootView(configPanes)
	v.tviewApp.SetRoot(v.tviewPages, true)
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
	for index, pane := range configPanes {
		tv := tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetChangedFunc(func() {
				v.tviewApp.Draw()
			}).ScrollToEnd()
		tv.
			SetBorder(true).
			SetTitle(getPaneTitle(index, pane, tv.HasFocus()))

		tv.SetBlurFunc(func() {
			tv.SetBorderColor(tcell.ColorWhite).
				SetTitle(getPaneTitle(index, pane, false))
		})
		tv.SetFocusFunc(func() {
			tv.SetBorderColor(tcell.ColorGreen).
				SetTitle(getPaneTitle(index, pane, true))
		})

		v.panes[index] = &Pane{
			textView: tv,
		}

		v.runUserCommand(pane.Dir, pane.Start, v.panes[index].textView)
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

func (v *View) openCommandOutputModal() {
	tv := tview.NewTextView()
	inputF := tview.NewInputField().
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)
	modal := func(p tview.Primitive) tview.Primitive {
		inputF.
			SetDoneFunc(func(key tcell.Key) {
				if key == tcell.KeyEnter {
					tv.Write([]byte(inputF.GetText() + "\n"))
					inputF.SetText("")
				}
			}).SetBorder(true)

		return tview.NewGrid().
			SetColumns(0).
			SetRows(0, 3).
			AddItem(p, 0, 0, 1, 1, 0, 0, true).
			AddItem(inputF, 1, 0, 1, 1, 0, 0, true)
	}
	tv.
		SetBorder(true).
		SetTitle("Command Output")

	v.tviewPages.AddPage(constant.Page.CommandOutputModalPage, modal(tv), true, true)
	v.tviewApp.SetFocus(inputF)
}
