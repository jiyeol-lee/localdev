package view

import (
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/jiyeol-lee/localdev/pkg/config"
	"github.com/rivo/tview"
)

func Test_getGridDimensions(t *testing.T) {
	type args struct {
		length int
	}
	tests := []struct {
		name     string
		args     args
		wantRows int
		wantCols int
	}{
		{
			name:     "0 length",
			args:     args{length: 0},
			wantRows: 0,
			wantCols: 0,
		},
		{
			name:     "1 length",
			args:     args{length: 1},
			wantRows: 1,
			wantCols: 1,
		},
		{
			name:     "2 length",
			args:     args{length: 2},
			wantRows: 1,
			wantCols: 2,
		},
		{
			name:     "3 length",
			args:     args{length: 3},
			wantRows: 2,
			wantCols: 2,
		},
		{
			name:     "4 length",
			args:     args{length: 4},
			wantRows: 2,
			wantCols: 2,
		},
		{
			name:     "5 length",
			args:     args{length: 5},
			wantRows: 2,
			wantCols: 3,
		},
		{
			name:     "6 length",
			args:     args{length: 6},
			wantRows: 2,
			wantCols: 3,
		},
		{
			name:     "7 length",
			args:     args{length: 7},
			wantRows: 2,
			wantCols: 4,
		},
		{
			name:     "8 length",
			args:     args{length: 8},
			wantRows: 2,
			wantCols: 4,
		},
		{
			name:     "9 length",
			args:     args{length: 9},
			wantRows: 2,
			wantCols: 5,
		},
		{
			name:     "10 length",
			args:     args{length: 10},
			wantRows: 2,
			wantCols: 5,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotRows, gotCols := getGridDimensions(tt.args.length)
			if gotRows != tt.wantRows {
				t.Errorf("getGridDimensions() gotRows = %v, want %v", gotRows, tt.wantRows)
			}
			if gotCols != tt.wantCols {
				t.Errorf("getGridDimensions() gotCols = %v, want %v", gotCols, tt.wantCols)
			}
		})
	}
}

func TestView_getRootView_resolvesPaneDirWithProjectSettingsDir(t *testing.T) {
	projectDir := t.TempDir()
	paneDir := "service"
	wantDir := filepath.Join(projectDir, paneDir)

	v := &View{}
	_, panes := v.getRootView(config.Config{
		ProjectSettings: &config.ProjectSettings{Dir: projectDir},
		Panes: []config.ConfigPane{
			{
				Name:  "api",
				Dir:   paneDir,
				Start: "sleep 10",
				Stop:  "true",
			},
		},
	})

	if got := panes[0].config.Dir; got != wantDir {
		t.Fatalf("pane dir = %q, want %q", got, wantDir)
	}
}

func TestPaneExpectedStopGenerationHelpers(t *testing.T) {
	p := &Pane{}

	if p.isExpectedStop(1) {
		t.Fatal("generation 1 should not be expected before marking")
	}

	p.markExpectedStop(1)
	p.markExpectedStop(3)
	if !p.isExpectedStop(1) {
		t.Fatal("generation 1 should be expected after marking")
	}
	if p.isExpectedStop(2) {
		t.Fatal("generation 2 should not be expected")
	}

	p.clearExpectedStop(1)
	if p.isExpectedStop(1) {
		t.Fatal("generation 1 should not be expected after clearing")
	}
	if !p.isExpectedStop(3) {
		t.Fatal("clearing generation 1 should not clear generation 3")
	}
}

func TestView_runPaneCommandToTextView_ReturnsErrorCount(t *testing.T) {
	tests := []struct {
		name    string
		dir     string
		command string
		wantErr bool
	}{
		{name: "success", dir: t.TempDir(), command: "true", wantErr: false},
		{name: "non-zero exit", dir: t.TempDir(), command: "exit 1", wantErr: true},
		{
			name:    "bad working directory",
			dir:     filepath.Join(t.TempDir(), "missing"),
			command: "true",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := startTestTviewApplication(t)
			v := &View{tviewApp: app}
			got := v.runPaneCommandToTextView("test-pane", tt.dir, tt.command, tview.NewTextView())
			if tt.wantErr && got == 0 {
				t.Fatalf("error count = %d, want > 0", got)
			}
			if !tt.wantErr && got != 0 {
				t.Fatalf("error count = %d, want 0", got)
			}
		})
	}
}

func TestView_GetManuallyStoppedPaneNamesOnlyIncludesSuccessfulStops(t *testing.T) {
	v := &View{panes: []*Pane{
		{config: config.ConfigPane{Name: "stopped"}, stopExecuted: true},
		{config: config.ConfigPane{Name: "failed"}, stopExecuted: false},
	}}

	got := v.GetManuallyStoppedPaneNames()
	if !got["stopped"] {
		t.Fatal("expected stopped pane to be included")
	}
	if got["failed"] {
		t.Fatal("expected failed stop pane to be excluded")
	}
}

func startTestTviewApplication(t *testing.T) *tview.Application {
	t.Helper()
	screen := tcell.NewSimulationScreen("")
	if err := screen.Init(); err != nil {
		t.Fatalf("failed to initialize simulation screen: %v", err)
	}
	app := tview.NewApplication().SetScreen(screen)
	done := make(chan error, 1)
	go func() {
		done <- app.Run()
	}()
	t.Cleanup(func() {
		app.Stop()
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("test tview application exited with error: %v", err)
			}
		case <-time.After(time.Second):
			t.Error("timed out waiting for test tview application to stop")
		}
	})
	app.QueueUpdate(func() {})
	return app
}

func Test_makeFlexibleSlice(t *testing.T) {
	type args struct {
		size int
	}
	tests := []struct {
		name string
		args args
		want []int
	}{
		{
			name: "size 0",
			args: args{size: 0},
			want: []int{},
		},
		{
			name: "size 1",
			args: args{size: 1},
			want: []int{0},
		},
		{
			name: "size 2",
			args: args{size: 2},
			want: []int{0, 0},
		},
		{
			name: "size 10",
			args: args{size: 10},
			want: []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeFlexibleSlice(tt.args.size); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeFlexibleSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}
