package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hardikphalet/go-redis/internal/server"
)

func main() {
	// Create a new server instance
	srv := server.New(":6379") // Default Redis port

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start the server in a goroutine
	go func() {
		log.Printf("Starting Redis server on port 6379...")
		if err := srv.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sig := <-sigChan
	log.Printf("Received signal %v, initiating graceful shutdown...", sig)

	// Initiate graceful shutdown
	if err := srv.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	log.Println("Server shutdown complete")
}
