package app

import "github.com/gdamore/tcell/v2"

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

// keyMapping handles key events for switching focus between text views.
func (a *App) keyMapping(event *tcell.EventKey) *tcell.EventKey {
	textViewsLength := len(a.textViews)

	if action, ok := keyToFocusAction(event.Rune(), textViewsLength); ok {
		a.tviewApp.SetFocus(a.textViews[action])
	}

	return event
}
