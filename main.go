package lafeature

import "github.com/LIVEauctioneers/lafeature/pkg/experiment/local"

type SimpleFlagManager = local.SimpleFlagManager
type SimpleFlagManagerConfig = local.SimpleFlagManagerConfig

// NewSimpleFlagManager creates a new SimpleFlagManager with default configuration.
// You can customize the configuration by modifying the returned manager's fields.
// func main() {
// 	// Create flag manager with API key and config
// 	manager := NewSimpleFlagManager("YOUR_API_KEY", SimpleFlagManagerConfig{
// 		Timeout: 10 * time.Second,
// 		Debug:   false,
// 	})

// 	// Start fetching and streaming flags
// 	err := manager.Start()
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer manager.Stop()

// 	// Check if individual flags are enabled
// 	if manager.MustEnabled("my-feature-flag") {
// 		fmt.Println("Feature is enabled!")
// 	} else {
// 		fmt.Println("Feature is disabled")
// 	}

// 	// Get all flags
// 	allFlags := manager.GetAll()
// 	fmt.Printf("All flags: %+v\n", allFlags)

//		// Keep running to receive streaming updates
//		fmt.Println("Listening for flag updates...")
//		time.Sleep(60 * time.Second)
//	}
func NewSimpleFlagManager(config SimpleFlagManagerConfig) *SimpleFlagManager {
	return local.NewSimpleFlagManager("", config)
}
