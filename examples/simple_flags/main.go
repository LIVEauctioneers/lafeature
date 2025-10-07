package main

import (
	"fmt"
	"time"

	"github.com/LIVEauctioneers/lafeature/pkg/experiment/local"
)

func main() {
	// Create flag manager with API key and config
	manager := local.NewSimpleFlagManager("YOUR_API_KEY", local.SimpleFlagManagerConfig{
		Timeout: 10 * time.Second,
		Debug:   false,
	})

	// Start fetching and streaming flags
	err := manager.Start()
	if err != nil {
		panic(err)
	}
	defer manager.Stop()

	// Give it a moment to fetch initial flags
	time.Sleep(2 * time.Second)

	// Check if individual flags are enabled
	if manager.MustEnabled("my-feature-flag") {
		fmt.Println("Feature is enabled!")
	} else {
		fmt.Println("Feature is disabled")
	}

	// Get all flags
	allFlags := manager.GetAll()
	fmt.Printf("All flags: %+v\n", allFlags)

	// Keep running to receive streaming updates
	fmt.Println("Listening for flag updates...")
	time.Sleep(60 * time.Second)
}
