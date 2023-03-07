package main

import (
	"flag"
	"log"

	"github.com/Frontman-Labs/frontman"
	"github.com/Frontman-Labs/frontman/config"
)

func main() {

	// Define command-line flags
	var configFile string
	flag.StringVar(&configFile, "config", "", "path to configuration file")

	// Parse command-line flags
	flag.Parse()

	// Load configuration from file or use default
	configPath := configFile
	if configPath == "" {
		configPath = "frontman.yaml"
	}

	config, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load configuration: %v", err)
	}

	// Create a new Gateway instance
	gateway, err := frontman.NewGateway(config)
	if err != nil {
		log.Fatalf("failed to create gateway: %v", err)
	}

	// Start the server
	log.Fatal(gateway.Start())
}
