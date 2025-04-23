package app

import (
	"fmt"
	"os/exec"
	"reflect"

	"github.com/gdamore/tcell/v2"
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

// keyToCommandAction maps key inputs to command actions defined in the configuration.
func keyToCommandAction(key rune, configCommands *ConfigCommands) ([]byte, error) {
	if configCommands == nil {
		return nil, fmt.Errorf("configCommands is nil")
	}

	fieldName := ""

	if key >= 97 && key <= 122 {
		fieldName = fmt.Sprintf("Lower%c", key-32)
	} else if key >= 65 && key <= 90 {
		fieldName = fmt.Sprintf("Upper%c", key)
	}

	if fieldName == "" {
		return nil, fmt.Errorf("invalid key: %c", key)
	}

	v := reflect.ValueOf(configCommands).Elem()
	field := v.FieldByName(fieldName)

	if !field.IsValid() || field.IsNil() {
		return nil, fmt.Errorf("command not found for key: %c", key)
	}

	cmdStruct := field.Interface().(*ConfigCommand)
	if cmdStruct == nil {
		return nil, fmt.Errorf("command not found for key: %c", key)
	}

	cmd := exec.Command("sh", "-c", cmdStruct.Command)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	if cmdStruct.Silent {
		return nil, nil
	}

	return output, nil
}

func (a *App) focusedViewIndex() int {
	for i, view := range a.views {
		if view.textView.HasFocus() {
			return i
		}
	}

	return -1
}

// keyMapping handles key events for switching focus between text views.
func (a *App) keyMapping(event *tcell.EventKey) *tcell.EventKey {
	viewsLength := len(a.views)

	if action, ok := keyToFocusAction(event.Rune(), viewsLength); ok {
		a.tviewApp.SetFocus(a.views[action].textView)
	}

	focusedViewIndex := a.focusedViewIndex()
	if focusedViewIndex != -1 {
		if output, err := keyToCommandAction(event.Rune(), a.config.Panes[focusedViewIndex].Commands); err == nil &&
			output != nil {
			a.views[a.focusedViewIndex()].textView.Write(output)
		}
	}

	return event
}
