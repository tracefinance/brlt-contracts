package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"vault0/internal/api"
	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	"vault0/internal/core/wallet"
	"vault0/internal/types"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	// Apply database migrations
	if err := db.MigrateDatabase(database, cfg); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}

	// Initialize keystore factory and create keystore
	keystoreFactory := keystore.NewFactory(database, cfg)
	ks, err := keystoreFactory.NewKeyStore()
	if err != nil {
		log.Fatalf("Failed to initialize keystore: %v", err)
	}

	// Initialize chain factory
	chainFactory := types.NewChainFactory(cfg)

	// Initialize wallet factory
	walletFactory := wallet.NewFactory(ks, cfg)

	// Initialize blockchain factory
	blockchainFactory := blockchain.NewFactory(cfg)

	// Initialize and start the server
	server := api.NewServer(
		database,
		cfg,
		ks,
		chainFactory,
		walletFactory,
		blockchainFactory,
	)

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
