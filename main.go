package main

import (
	"fmt"
	"jiyeol-lee/localdev/pkg/app"
	"log"
)

func main() {
	a, err := app.Run()
	if err != nil {
		log.Fatalf("Error initializing app: %v", err)
	}

	// When app.Stop() is called, this will run afterward
	defer func() {
		fmt.Println("🛑 Stopping all spaces...")
		a.StopSpaces()
		fmt.Println("✅ All spaces stopped.")
	}()
}
