package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"vault0/internal/config"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	conn   *sql.DB
	config *config.Config
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config) (*DB, error) {
	// Create a database connection string for SQLite
	connStr := fmt.Sprintf("file:%s", cfg.DBPath)

	// Connect to the database
	conn, err := sql.Open("sqlite3", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Verify the connection works
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Connected to database at %s", cfg.DBPath)
	return &DB{conn: conn, config: cfg}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// GetConnection returns the underlying database connection
func (db *DB) GetConnection() *sql.DB {
	return db.conn
}

// ExecuteQuery executes a query with parameters and returns the result
func (db *DB) ExecuteQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.Query(query, args...)
}

// ExecuteQueryContext executes a query with context and parameters and returns the result
func (db *DB) ExecuteQueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return db.conn.QueryContext(ctx, query, args...)
}

// ExecuteStatement executes a statement with parameters
func (db *DB) ExecuteStatement(query string, args ...interface{}) (sql.Result, error) {
	return db.conn.Exec(query, args...)
}

// ExecuteStatementContext executes a statement with context and parameters
func (db *DB) ExecuteStatementContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return db.conn.ExecContext(ctx, query, args...)
}
