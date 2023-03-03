package main

import (
	"github.com/hyperioxx/frontman"
	"log"
)

func main() {

	// Create a new Gateway instance
	gateway, err := frontman.NewGateway(frontman.NewRedisClient)
	if err != nil {
		log.Fatal(err)
	}

	// Start the server
	if err := gateway.Start(); err != nil {
		log.Fatal(err)
	}
}
