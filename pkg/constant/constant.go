package constant

var Page = struct {
	MainPage  string
	ModalPage string
}{
	MainPage:  "main",
	ModalPage: "modal",
}

var ReservedCommand = struct {
	KillProcess string
}{
	KillProcess: "__KILL_PROCESS__",
}
