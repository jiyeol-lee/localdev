package view

import "github.com/rivo/tview"

type commandHelpModal struct {
	callerPaneIndex int
	textView        *tview.TextView
}

func newCommandHelpModal() *commandHelpModal {
	return &commandHelpModal{
		textView: tview.NewTextView(),
	}
}

func (c *commandHelpModal) reset() {
	c.textView = nil
}
