package constant

var Page = struct {
	MainPage               string
	CommandOutputModalPage string
}{
	MainPage:               "main",
	CommandOutputModalPage: "command_output_modal",
}

var ReservedCommand = struct {
	KillProcess string
}{
	KillProcess: "__KILL_PROCESS__",
}
