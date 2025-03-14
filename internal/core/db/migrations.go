package db

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateDatabase applies all available migrations to the database
func (db *DB) MigrateDatabase() error {
	// Get the database driver
	driver, err := sqlite3.WithInstance(db.GetConnection(), &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create database driver: %w", err)
	}

	// Create the migrate instance
	sourceURL := fmt.Sprintf("file://%s", db.config.MigrationsPath)
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"sqlite3", // Database name (just a label for migrate)
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	// Apply all up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}
