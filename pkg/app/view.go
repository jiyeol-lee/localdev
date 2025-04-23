package app

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"syscall"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/pkg/command"
	"github.com/jiyeol-lee/localdev/pkg/constant"
	"github.com/rivo/tview"
)

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
func (a *App) runUserCommand(dir string, userCmd string, view *AppView) {
	cmd := exec.Command("sh", "-c", userCmd)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Panicln("Error starting command:", err)
	}

	view.processId = cmd.Process.Pid

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			a.tviewApp.QueueUpdateDraw(func() {
				view.textView.Write([]byte(line + "\n"))
			})
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			a.tviewApp.QueueUpdateDraw(func() {
				view.textView.Write([]byte("[#8B4513]" + line + "[white]\n"))
			})
		}
	}()
}

// getPaneTitle generates the title for each pane in the grid
func getPaneTitle(paneIndex int, pane ConfigPane, focused bool) string {
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

// getRootView creates the root view for the application
func (a *App) getRootView() *tview.Pages {
	l := len(a.config.Panes)
	a.views = make([]*AppView, l)
	rows, cols := getGridDimensions(l)

	root := tview.NewPages()
	grid := tview.NewGrid()
	grid.
		SetRows(makeFlexibleSlice(rows)...).
		SetColumns(makeFlexibleSlice(cols)...)

	row := 0
	col := 0
	for index, pane := range a.config.Panes {
		tv := tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetChangedFunc(func() {
				a.tviewApp.Draw()
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

		a.views[index] = &AppView{
			textView:  tv,
			processId: 0,
		}

		a.runUserCommand(pane.Dir, pane.Start, a.views[index])
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
