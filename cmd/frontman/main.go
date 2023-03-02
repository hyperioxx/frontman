package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/hyperioxx/frontman"
)



func main() {
    
    // Retrieve the database URI from the environment variables
    uri := os.Getenv("DATABASE_URI")
    if uri == "" {
        log.Fatal("DATABASE_URI environment variable is not set")
    }

    // Create a new Gateway instance
    gateway, err := frontman.NewGateway(func() (*sql.DB, error) {
        return frontman.NewDB(uri)
    })
    if err != nil {
        log.Fatal(err)
    }

    // Start the server
    if err := gateway.Start(); err != nil {
        log.Fatal(err)
    }
}