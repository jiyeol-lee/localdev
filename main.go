package main

import (
	"fmt"
	"log"

	"github.com/jiyeol-lee/localdev/pkg/app"
)

func main() {
	a, err := app.Run()
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
