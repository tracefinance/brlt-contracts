package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"

	_ "github.com/mattn/go-sqlite3"
)

// DB represents the database connection
type DB struct {
	Conn      *sql.DB
	Config    *config.Config
	Snowflake *Snowflake
	Logger    logger.Logger
}

// NewDatabase creates a new database connection
func NewDatabase(cfg *config.Config, snowflake *Snowflake, log logger.Logger) (*DB, error) {
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
	return &DB{Conn: conn, Config: cfg, Logger: log}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	if db.Conn != nil {
		if err := db.Conn.Close(); err != nil {
			return errors.NewDatabaseError(err)
		}
	}
	return nil
}

// GetConnection returns the underlying database connection
func (db *DB) GetConnection() *sql.DB {
	return db.Conn
}

// ExecuteQuery executes a query with parameters and returns the result
func (db *DB) ExecuteQuery(query string, args ...any) (*sql.Rows, error) {
	rows, err := db.Conn.Query(query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return rows, nil
}

// ExecuteQueryContext executes a query with context and parameters and returns the result
func (db *DB) ExecuteQueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	rows, err := db.Conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return rows, nil
}

// ExecuteStatement executes a statement with parameters
func (db *DB) ExecuteStatement(query string, args ...any) (sql.Result, error) {
	result, err := db.Conn.Exec(query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return result, nil
}

// ExecuteStatementContext executes a statement with context and parameters
func (db *DB) ExecuteStatementContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	result, err := db.Conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	return result, nil
}

// UnmarshalJSONToMap unmarshals a SQL JSON string to a map[string]string
// If the JSON is invalid or empty, it returns an initialized empty map
func UnmarshalJSONToMap(jsonStr sql.NullString) map[string]string {
	result := make(map[string]string)

	if jsonStr.Valid && jsonStr.String != "" {
		err := json.Unmarshal([]byte(jsonStr.String), &result)
		if err != nil {
			// If we can't parse the JSON, return the empty map
			return result
		}
	}

	return result
}

func (db *DB) GenerateID() (int64, error) {
	id, err := db.Snowflake.GenerateID()
	if err != nil {
		return 0, errors.NewDatabaseError(err)
	}
	return id, nil
}
