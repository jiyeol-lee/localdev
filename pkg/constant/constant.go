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
