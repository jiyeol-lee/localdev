package constant

var Page = struct {
	MainPage               string
	CommandOutputModalPage string
	CommandHelpModalPage   string
}{
	MainPage:               "main",
	CommandOutputModalPage: "command_output_modal",
	CommandHelpModalPage:   "command_help_modal",
}

var ReservedCommand = struct {
	KillProcess string
}{
	KillProcess: "__KILL_PROCESS__",
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
