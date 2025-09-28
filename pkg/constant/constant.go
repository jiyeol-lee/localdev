package constant

var Page = struct {
	MainPage               string
	CommandOutputModalPage string
	CommandHelpModalPage   string
	MaximizedPane          string
}{
	MainPage:               "main",
	CommandOutputModalPage: "command_output_modal",
	CommandHelpModalPage:   "command_help_modal",
	MaximizedPane:          "maximized_pane",
}

var ReservedCommand = struct {
	TogglePaneSize string
}{
	TogglePaneSize: "<toggle_pane_size>",
}

var AnsiColor = struct {
	Red   string
	Green string
	Reset string
}{
	Red:   "\033[31m",
	Green: "\033[32m",
	Reset: "\033[0m",
}

// MaxPaneOutputLines defines the maximum number of lines that can be displayed in a pane.
// The limit of 500 was chosen to balance performance and usability:
// - Performance: Rendering too many lines can degrade UI responsiveness.
// - Usability: Limiting the output prevents overwhelming the user with excessive information.
var MaxPaneOutputLines = 500
