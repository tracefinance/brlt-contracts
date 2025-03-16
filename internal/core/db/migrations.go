package db

import (
	"vault0/internal/errors"

	"vault0/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// MigrateDatabase applies all available migrations to the database
func (db *DB) MigrateDatabase() error {
	// Get the database driver
	driver, err := sqlite3.WithInstance(db.GetConnection(), &sqlite3.Config{})
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	// Create the migrate instance
	sourceURL := "file://" + db.config.MigrationsPath
	m, err := migrate.NewWithDatabaseInstance(
		sourceURL,
		"sqlite3", // Database name (just a label for migrate)
		driver,
	)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	// Apply all up migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return errors.NewDatabaseError(err)
	}

	db.logger.Info("Database migrations applied successfully",
		logger.String("path", db.config.MigrationsPath))
	return nil
}
