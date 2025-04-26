package view

import "github.com/rivo/tview"

type commandOutputModal struct {
	callerPaneIndex     int
	textView            *tview.TextView
	inputField          *tview.InputField
	commandHistoryIndex int
	commandHistory      []string
}

func newCommandOutputModal() *commandOutputModal {
	return &commandOutputModal{
		callerPaneIndex:     -1,
		textView:            tview.NewTextView(),
		inputField:          tview.NewInputField(),
		commandHistoryIndex: -1,
		commandHistory:      make([]string, 0),
	}
}

func (c *commandOutputModal) appendCommandHistory(history string) {
	c.commandHistory = append(c.commandHistory, history)
}

func (c *commandOutputModal) resetCommandHistoryIndex() {
	c.commandHistoryIndex = -1
}

func (c *commandOutputModal) resetCommandHistory() {
	c.commandHistory = make([]string, 0)
	c.resetCommandHistoryIndex()
}

func (c *commandOutputModal) reset() {
	c.callerPaneIndex = -1
	c.textView = nil
	c.inputField = nil
	c.resetCommandHistory()
}

func (c *commandOutputModal) previousCommand() string {
	cHistoryLen := len(c.commandHistory)
	if cHistoryLen == 0 {
		return ""
	}
	if c.commandHistoryIndex > 0 {
		c.commandHistoryIndex--
	} else if c.commandHistoryIndex <= 0 {
		c.commandHistoryIndex = len(c.commandHistory) - 1
	}
	return c.commandHistory[c.commandHistoryIndex]
}

func (c *commandOutputModal) nextCommand() string {
	cHistoryLen := len(c.commandHistory)
	if cHistoryLen == 0 {
		return ""
	}
	if c.commandHistoryIndex < cHistoryLen-1 {
		c.commandHistoryIndex++
	} else if c.commandHistoryIndex >= cHistoryLen-1 {
		c.commandHistoryIndex = 0
	}
	return c.commandHistory[c.commandHistoryIndex]
}
