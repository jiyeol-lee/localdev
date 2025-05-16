package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/jiyeol-lee/localdev/pkg/app"
)

func main() {
	configFileName := flag.String("config", "config.yml", "Path to config file")
	flag.Parse()

	a, err := app.Run(*configFileName)
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	// When app.Stop() is called, this will run afterward
	defer func() {
		fmt.Println("🛑 Stopping all panes...")
		a.StopPanes()
		fmt.Println("✅ All panes stopped.")
	}()
}
