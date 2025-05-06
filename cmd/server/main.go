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

	// Register wallet service as a transaction transformer
	if err := container.Services.Transaction.TransformerService.RegisterTransformer("wallet", container.Services.WalletService); err != nil {
		log.Error("Failed to register wallet transformer", logger.Error(err))
	}

	// Register blockchain transformer for blockchain data enrichment
	if err := container.Services.Transaction.TransformerService.RegisterTransformer("blockchain_data", container.Services.Transaction.BlockchainTransformer); err != nil {
		log.Error("Failed to register blockchain transformer", logger.Error(err))
	}

	// Start pending transaction polling
	container.Services.Transaction.PoolingService.StartPendingTransactionPolling(ctx)

	// Start wallet monitoring to track balances and transactions
	if err := container.Services.WalletMonitorService.StartWalletMonitoring(ctx); err != nil {
		log.Error("Failed to start wallet monitoring", logger.Error(err))
	}

	// Start token transaction monitoring
	if err := container.Services.TokenMonitorService.StartTokenTransactionMonitoring(ctx); err != nil {
		log.Error("Failed to start token transaction monitoring", logger.Error(err))
	}

	// Start transaction monitoring
	if err := container.Services.Transaction.MonitorService.StartTransactionMonitoring(ctx); err != nil {
		log.Error("Failed to start transaction monitoring", logger.Error(err))
	}

	// Start transaction history synchronization
	if err := container.Services.Transaction.HistoryService.StartTransactionSyncing(ctx); err != nil {
		log.Error("Failed to start transaction history syncing", logger.Error(err))
	}

	// Start token price update job
	container.Services.TokenPricePollingService.StartPricePolling(ctx)

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

	// Stop wallet monitoring
	if err := container.Services.WalletMonitorService.StopWalletMonitoring(ctx); err != nil {
		log.Error("Failed to stop wallet monitoring", logger.Error(err))
	}

	// Stop token transaction monitoring
	if err := container.Services.TokenMonitorService.StopTokenTransactionMonitoring(); err != nil {
		log.Error("Failed to stop token transaction monitoring", logger.Error(err))
	}

	// Unsubscribe from transaction events first to close the channel
	// before any services that might use it
	container.Services.Transaction.MonitorService.StopTransactionMonitoring()

	// Stop transaction history synchronization
	container.Services.Transaction.HistoryService.StopTransactionSyncing()

	// Stop pending transaction polling
	container.Services.Transaction.PoolingService.StopPendingTransactionPolling()

	// Stop vault recovery polling
	container.Services.VaultService.StopRecoveryPolling()

	// Stop vault deployment monitoring
	container.Services.VaultService.StopDeploymentMonitoring()

	// Stop token price update job
	container.Services.TokenPricePollingService.StopPricePolling()

	// Perform cleanup
	container.Server.Shutdown()

	// Close the database connection
	if container.Core.DB != nil {
		if err := container.Core.DB.Close(); err != nil {
			log.Error("Failed to close database connection", logger.Error(err))
		}
	}

	log.Info("Server gracefully stopped")
}
