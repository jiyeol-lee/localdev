package app

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	"github.com/gdamore/tcell/v2"
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

// runUserCommand executes a user command in the specified directory and writes the output to the provided text view
func (a *App) runUserCommand(userCmd string, textView *tview.TextView) {
	cmd := exec.Command("sh", "-c", userCmd)

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Panicln("Error starting command:", err)
	}

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			a.tviewApp.QueueUpdateDraw(func() {
				textView.Write([]byte(line + "\n"))
			})
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			a.tviewApp.QueueUpdateDraw(func() {
				textView.Write([]byte("[#8B4513]" + line + "[white]\n"))
			})
		}
	}()
}

// getRootView creates the root view for the application
func (a *App) getRootView() *tview.Pages {
	l := len(a.config.Spaces)
	a.textViews = make([]*tview.TextView, l)
	rows, cols := getGridDimensions(l)

	root := tview.NewPages()
	grid := tview.NewGrid()
	grid.
		SetRows(makeFlexibleSlice(rows)...).
		SetColumns(makeFlexibleSlice(cols)...)

	row := 0
	col := 0
	for index, space := range a.config.Spaces {
		tv := tview.NewTextView().
			SetDynamicColors(true).
			SetScrollable(true).
			SetChangedFunc(func() {
				a.tviewApp.Draw()
			}).ScrollToEnd()
		tv.
			SetBorder(true).
			SetTitle(fmt.Sprintf("[%d] %s", index+1, space.Name))

		tv.SetBlurFunc(func() {
			tv.SetBorderColor(tcell.ColorWhite)
		})
		tv.SetFocusFunc(func() {
			tv.SetBorderColor(tcell.ColorGreen)
		})

		a.textViews[index] = tv

		a.runUserCommand("cd "+space.Dir+" && "+space.Start, tv)
		grid.AddItem(tv, row, col, 1, 1, 0, 0, true)
		if row == 1 {
			row = 0
			col++
		} else {
			row++
		}
	}
	root.AddPage("grid", grid, true, true)

	return root
}
