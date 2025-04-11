package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"vault0/internal/logger"
	"vault0/internal/wire"
)

func main() {
	// Build the dependency injection container
	container, err := wire.BuildContainer()
	if err != nil {
		// Since we don't have the logger yet, use os.Exit
		os.Stderr.WriteString("Failed to build dependency container: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Get the logger from the container
	log := container.Core.Logger

	// Migrate the database
	if err := container.Core.DB.MigrateDatabase(); err != nil {
		log.Fatal("Failed to migrate database", logger.Error(err))
	}

	// Initialize token store with tokens from database
	log.Info("Initializing token store")

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start transaction event subscriptions
	container.Services.TransactionService.SubscribeToTransactionEvents(ctx)

	// Start pending transaction polling
	container.Services.TransactionService.StartPendingTransactionPolling(ctx)

	// Start token price update job
	container.Services.TokenPriceService.StartPriceUpdateJob(ctx)

	// Setup routes
	container.Server.SetupRoutes()

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Run the server in a goroutine
	go func() {
		if err := container.Server.Run(); err != nil {
			log.Fatal("Failed to start server", logger.Error(err))
		}
	}()

	log.Info("Server is running",
		logger.String("port", container.Core.Config.Port),
		logger.String("message", "Press Ctrl+C to shutdown"),
	)

	// Wait for interrupt signal
	<-quit
	log.Info("Received shutdown signal")

	// Cancel the root context
	cancel()

	// Unsubscribe from events
	container.Services.TransactionService.UnsubscribeFromTransactionEvents()

	// Stop pending transaction polling
	container.Services.TransactionService.StopPendingTransactionPolling()

	// Stop token price update job
	container.Services.TokenPriceService.StopPriceUpdateJob()

	// Perform cleanup
	container.Server.Shutdown()

	// Close the database connection
	if container.Core.DB != nil {
		container.Core.DB.Close()
	}

	log.Info("Server gracefully stopped")
}
