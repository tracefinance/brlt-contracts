package main

import (
	"os"
	"os/signal"
	"syscall"

	"vault0/internal/api"
	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	"vault0/internal/core/wallet"
	"vault0/internal/logger"
	"vault0/internal/types"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	log, err := logger.NewLogger(cfg.Log)
	if err != nil {
		// Since we don't have the logger yet, use os.Exit
		os.Exit(1)
	}

	// Initialize database
	database, err := db.NewDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to initialize database", logger.Error(err))
	}
	defer database.Close()

	// Apply database migrations
	if err := db.MigrateDatabase(database, cfg); err != nil {
		log.Fatal("Failed to apply database migrations", logger.Error(err))
	}

	// Initialize keystore factory and create keystore
	keystoreFactory := keystore.NewFactory(database, cfg)
	ks, err := keystoreFactory.NewKeyStore()
	if err != nil {
		log.Fatal("Failed to initialize keystore", logger.Error(err))
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
		log,
	)

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		if err := server.Run(); err != nil {
			log.Fatal("Failed to start server", logger.Error(err))
		}
	}()

	log.Info("Server is running",
		logger.String("port", cfg.Port),
		logger.String("message", "Press Ctrl+C to shutdown"),
	)

	// Wait for interrupt signal
	<-quit
	log.Info("Received shutdown signal")

	// Perform cleanup
	server.Shutdown()

	log.Info("Server gracefully stopped")
}
