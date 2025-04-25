package view

import (
	"fmt"
	"os/exec"
	"reflect"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/pkg/config"
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
func (v *View) keyToCommand(key rune, configPane config.ConfigPane) (string, error) {
	fieldName := ""
	if key >= 97 && key <= 122 {
		fieldName = fmt.Sprintf("Lower%c", key-32)
	} else if key >= 65 && key <= 90 {
		fieldName = fmt.Sprintf("Upper%c", key)
	}

	if fieldName == "" {
		return "", fmt.Errorf("invalid key: %c", key)
	}

	if configPane.Commands == nil {
		return "", fmt.Errorf("commands not found for key: %c", key)
	}

	vo := reflect.ValueOf(configPane.Commands).Elem()
	field := vo.FieldByName(fieldName)

	if !field.IsValid() || field.IsNil() {
		return "", fmt.Errorf("command not found for key: %c", key)
	}

	cmdStruct := field.Interface().(*config.ConfigCommand)
	if cmdStruct == nil {
		return "", fmt.Errorf("command not found for key: %c", key)
	}

	if cmdStruct.Silent {
		cmd := exec.Command("sh", "-c", cmdStruct.Command)
		_, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return cmdStruct.Command, nil
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
		// handle pane focus switching when there is no modal focused
		if action, ok := keyToFocusAction(event.Rune(), panesLength); ok {
			v.tviewApp.SetFocus(v.panes[action].textView)
		}

		if event.Rune() == 63 {
			if !v.checkIsCommandHelpModalOpen() {
				v.commandHelpModal.callerPaneIndex = focusedViewIndex
				v.commandHelpModal.textView = v.openCommandHelpModal()
				v.setCommandHelpModalBodyText()
			}
			return event
		}

		if command, err := v.keyToCommand(event.Rune(), v.panes[focusedViewIndex].config); err == nil &&
			command != "" {
			if !v.checkIsCommandOutputModalOpen() {
				v.commandOutputModal.callerPaneIndex = focusedViewIndex
				v.commandOutputModal.inputField, v.commandOutputModal.textView = v.openCommandOutputModal()
				v.commandOutputModal.appendCommandHistory(command)
				v.runUserCommand(
					v.panes[v.commandOutputModal.callerPaneIndex].config.Dir,
					command,
					v.commandOutputModal.textView,
				)
			}
		}
	}

	return event
}
