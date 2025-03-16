package db

import (
	"context"
	"database/sql"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	conn   *sql.DB
	config *config.Config
	logger logger.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config, log logger.Logger) (*DB, error) {
	// Create a database connection string for SQLite
	connStr := fmt.Sprintf("file:%s", cfg.DBPath)

	// Connect to the database
	conn, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// Verify the connection works
	if err := conn.Ping(); err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	log.Info("Connected to database", logger.String("path", cfg.DBPath))
	return &DB{conn: conn, config: cfg, logger: log}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		if err := db.conn.Close(); err != nil {
			return errors.NewDatabaseError(err)
		}
	}
	return nil
}

// GetConnection returns the underlying database connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// ExecuteQuery executes a query with parameters and returns the result
func (db *DB) ExecuteQuery(query string, args ...any) (*sql.Rows, error) {
	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return rows, nil
}

// ExecuteQueryContext executes a query with context and parameters and returns the result
func (db *DB) ExecuteQueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := db.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return rows, nil
}

// ExecuteStatement executes a statement with parameters
func (db *DB) ExecuteStatement(query string, args ...any) (sql.Result, error) {
	result, err := db.conn.Exec(query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return result, nil
}

// ExecuteStatementContext executes a statement with context and parameters
func (db *DB) ExecuteStatementContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	result, err := db.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return result, nil
}
