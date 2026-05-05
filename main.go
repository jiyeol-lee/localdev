package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jiyeol-lee/localdev/internal/logger"
	"github.com/jiyeol-lee/localdev/pkg/app"
)

func main() {
	if err := logger.Init(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Local Dev failed to initialize logging: %v\n", err)
		os.Exit(1)
	}

	configFileName := flag.String(
		"config",
		"config.yml",
		"Config file name under the Local Dev config directory",
	)
	flag.Parse()

	a, err := app.Run(*configFileName)
	if err != nil {
		logger.Errorf("error initializing app: %v", err)
		_, _ = fmt.Fprintf(
			os.Stderr,
			"Local Dev failed: %v; see log file: %s\n",
			err,
			logger.Path(),
		)
		_ = logger.Close()
		os.Exit(1)
	}

	defer func() {
		fmt.Println("🛑 Stopping all panes...")
		stopErrorCount := a.StopPanes()
		if stopErrorCount > 0 {
			logger.Errorf("shutdown completed with %d stop command error(s)", stopErrorCount)
			_, _ = fmt.Fprintf(
				os.Stderr,
				"Some stop commands failed (%d); see log file: %s\n",
				stopErrorCount,
				logger.Path(),
			)
			fmt.Println("⚠️ Pane shutdown completed with errors.")
		} else {
			fmt.Println("✅ All panes stopped.")
		}
		_ = logger.Close()
	}()
}
