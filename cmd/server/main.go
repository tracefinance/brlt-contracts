package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"vault0/internal/logger"
	"vault0/internal/wire"
)

// @title Vault0 API
// @version 1.0
// @description API Server for Vault0 application

// @contact.name Vault0 API Support
// @contact.email support@vault0.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
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
	container.Services.TokenPriceService.StartPricePolling(ctx)

	// Start transaction monitoring
	container.Services.WalletService.StartTransactionMonitoring(ctx)

	// Start wallet history syncing
	container.Services.WalletService.StartWalletHistorySyncing(ctx)

	// Start vault recovery polling
	container.Services.VaultService.StartRecoveryPolling(ctx)

	// Start vault deployment monitoring
	container.Services.VaultService.StartDeploymentMonitoring(ctx)

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

	// Unsubscribe from transaction events first to close the channel
	// before any services that might use it
	container.Services.TransactionService.UnsubscribeFromTransactionEvents()

	// Stop pending transaction polling
	container.Services.TransactionService.StopPendingTransactionPolling()

	// Stop transaction monitoring in wallet service
	container.Services.WalletService.StopTransactionMonitoring()

	// Stop wallet history syncing
	container.Services.WalletService.StopWalletHistorySyncing()

	// Stop vault recovery polling
	container.Services.VaultService.StopRecoveryPolling()

	// Stop vault deployment monitoring
	container.Services.VaultService.StopDeploymentMonitoring()

	// Stop token price update job
	container.Services.TokenPriceService.StopPricePolling()

	// Perform cleanup
	container.Server.Shutdown()

	// Close the database connection
	if container.Core.DB != nil {
		container.Core.DB.Close()
	}

	log.Info("Server gracefully stopped")
}
