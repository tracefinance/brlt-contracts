package main

import (
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
	log := container.Logger

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
		logger.String("port", container.Config.Port),
		logger.String("message", "Press Ctrl+C to shutdown"),
	)

	// Wait for interrupt signal
	<-quit
	log.Info("Received shutdown signal")

	// Perform cleanup
	container.Server.Shutdown()

	// Close the database connection
	if container.DB != nil {
		container.DB.Close()
	}

	log.Info("Server gracefully stopped")
}
