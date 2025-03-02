package config

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	// DBPath is the path to the SQLite database file
	DBPath string
	// Port is the server port to listen on
	Port string
	// UIPath is the path to the static React UI files
	UIPath string
	// MigrationsPath is the path to the migration files
	MigrationsPath string
}

// LoadConfig loads the application configuration from environment variables
// or falls back to default values
func LoadConfig() *Config {
	// Get base directory for paths
	baseDir := os.Getenv("APP_BASE_DIR")
	if baseDir == "" {
		// Default to current working directory if not specified
		currentDir, _ := os.Getwd()
		baseDir = currentDir
	}

	// Get the database path from environment or use default
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		// Default to a db file in the project root
		dbPath = filepath.Join(baseDir, "vault0.db")
	}

	// Get the port from environment or use default
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// Set UI path, default to ui/dist
	uiPath := os.Getenv("UI_PATH")
	if uiPath == "" {
		uiPath = filepath.Join(baseDir, "ui", "dist")
	}

	// Set migrations path
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = filepath.Join(baseDir, "migrations")
	}

	return &Config{
		DBPath:         dbPath,
		Port:           port,
		UIPath:         uiPath,
		MigrationsPath: migrationsPath,
	}
}
