package view

import (
	"fmt"
	"os/exec"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/pkg/config"
	"github.com/jiyeol-lee/localdev/pkg/constant"
	"github.com/jiyeol-lee/localdev/pkg/internal/shell"
	"github.com/jiyeol-lee/localdev/pkg/util"
)

// keyToFocusAction maps key inputs to focus actions for text views.
func keyToFocusAction(key rune, textViewsLength int) (int, bool) {
	switch key {
	case 49: // 1
		return 0, textViewsLength > 0
	case 50: // 2
		return 1, textViewsLength > 1
	case 51: // 3
		return 2, textViewsLength > 2
	case 52: // 4
		return 3, textViewsLength > 3
	case 53: // 5
		return 4, textViewsLength > 4
	case 54: // 6
		return 5, textViewsLength > 5
	case 55: // 7
		return 6, textViewsLength > 6
	case 56: // 8
		return 7, textViewsLength > 7
	case 57: // 9
		return 8, textViewsLength > 8
	case 48: // 0
		return 9, textViewsLength > 9
	default:
		return -1, false
	}
}

// keyToCommand maps key inputs to command actions defined in the configuration.
func (v *View) keyToCommand(
	keyRune rune,
	configPane config.ConfigPane,
) (*config.ConfigCommand, error) {
	if configPane.Commands == nil {
		return nil, fmt.Errorf("no commands defined for pane: %s", configPane.Name)
	}

	keyName := ""
	if keyRune >= 97 && keyRune <= 122 {
		keyName = fmt.Sprintf("lower%c", keyRune-32)
	} else if keyRune >= 65 && keyRune <= 90 {
		keyName = fmt.Sprintf("upper%c", keyRune)
	}
	if keyName == "" {
		return nil, fmt.Errorf("invalid keyRune: %c", keyRune)
	}

	paneCommandsMap, err := util.YamlToMap[*config.ConfigCommands, *config.ConfigCommand](
		configPane.Commands,
	)
	if err != nil {
		return nil, err
	}

	var configCommand *config.ConfigCommand
	for k, cc := range paneCommandsMap {
		if k == keyName {
			configCommand = cc
			break
		}
	}

	if configCommand == nil {
		return nil, fmt.Errorf("no command found for key: %s", keyName)
	}

	return configCommand, nil
}

func (v *View) focusedViewIndex() int {
	for i, pane := range v.panes {
		if pane.textView.HasFocus() {
			return i
		}
	}

	return -1
}

// keyMapping handles key events for switching focus between text views.
func (v *View) keyMapping(event *tcell.EventKey) *tcell.EventKey {
	panesLength := len(v.panes)
	focusedViewIndex := v.focusedViewIndex()

	if focusedViewIndex != -1 {
		// handle pane focus switching when there is no modal focused and no pane is maximized
		if action, ok := keyToFocusAction(event.Rune(), panesLength); ok &&
			!v.checkIsPaneMaximized() {
			v.tviewApp.SetFocus(v.panes[action].textView)
		}

		// open command help modal when '?' is pressed
		if event.Rune() == 63 {
			if !v.checkIsCommandHelpModalOpen() && !v.checkIsCommandOutputModalOpen() {
				v.commandHelpModal.callerPaneIndex = focusedViewIndex
				v.commandHelpModal.textView = v.openCommandHelpModal()
				v.setCommandHelpModalBodyText()
				v.disablePanesMouse()
			}
			return event
		}

		configPane := v.panes[focusedViewIndex].config
		if configCommand, err := v.keyToCommand(event.Rune(), configPane); err == nil &&
			configCommand.Command != "" {
			if configCommand.Command == constant.ReservedCommand.TogglePaneSize {
				v.togglePaneSize()
				return event
			}
			if configCommand.Silent {
				sh := shell.Current()
				cmd := exec.Command(sh, "-c", configCommand.Command)
				cmd.Dir = configPane.Dir
				err := cmd.Start()
				if err != nil {
					v.panes[focusedViewIndex].textView.Write(
						fmt.Appendf(nil, "[red]Command execution failed: %s[white]\n",
							err,
						),
					)
				} else {
					v.panes[focusedViewIndex].textView.Write(
						fmt.Appendf(nil, "[green]Command started successfully: %s[white]\n",
							configCommand.Command,
						),
					)
				}
				return event
			}
			if !v.checkIsCommandHelpModalOpen() && !v.checkIsCommandOutputModalOpen() {
				v.commandOutputModal.callerPaneIndex = focusedViewIndex
				v.commandOutputModal.appendCommandHistory(configCommand.Command)
				if configCommand.AutoExecute {
					v.runCustomUserCommand(
						v.panes[v.commandOutputModal.callerPaneIndex].config.Dir,
						configCommand.Command,
					)
				} else {
					v.commandOutputModal.inputField = v.openCommandOutputModal()
					v.commandOutputModal.inputField.SetText(configCommand.Command)
				}
				v.disablePanesMouse()
			}
		}
	}

	return event
}
