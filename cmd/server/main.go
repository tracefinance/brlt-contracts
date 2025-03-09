package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"vault0/internal/api"
	"vault0/internal/config"
	"vault0/internal/db"
	"vault0/internal/keystore"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Apply database migrations
	if err := db.MigrateDatabase(database, cfg); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}

	// Initialize keystore
	keystoreFactory := keystore.NewFactory(cfg, database.GetConnection())
	keyStore, err := keystoreFactory.NewDefaultKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize keystore: %v", err)
	}

	// Initialize and start the server
	server := api.New(database, keyStore, cfg)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		if err := server.Run(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server is running on port %s", cfg.Port)
	log.Printf("Press Ctrl+C to shutdown")

	// Wait for interrupt signal
	<-quit
	log.Println("Shutting down server...")

	// Perform cleanup
	server.Shutdown()

	log.Println("Server gracefully stopped")
}
