package keystore

import (
	"database/sql"
	"os"
	"testing"

	"vault0/internal/config"
	"vault0/internal/keygen"

	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// testConfig creates a config for testing purposes
func testConfig() *config.Config {
	// Generate a random encryption key for tests
	encKey, _ := keygen.GenerateEncryptionKeyBase64(32)

	return &config.Config{
		DBPath:          ":memory:", // Use in-memory SQLite for tests
		DBEncryptionKey: encKey,
	}
}

// setupTestDB sets up a test database with the keys table
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)

	// Create the keys table
	// Note: Using the same schema as expected by the DBKeyStore implementation
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS keys (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			key_type TEXT NOT NULL,
			tags TEXT,
			private_key BLOB,
			public_key BLOB,
			created_at INTEGER NOT NULL
		)
	`)
	require.NoError(t, err)

	// Return a cleanup function
	cleanup := func() {
		db.Close()
	}

	return db, cleanup
}

// setupTestKeyStore creates a test KeyStore instance
func setupTestKeyStore(t *testing.T) (KeyStore, *sql.DB, func()) {
	db, dbCleanup := setupTestDB(t)
	cfg := testConfig()

	// Create a DBKeyStore
	dbKeyStore, err := NewDBKeyStore(db, cfg)
	require.NoError(t, err)

	cleanup := func() {
		dbKeyStore.Close()
		dbCleanup()
	}

	return dbKeyStore, db, cleanup
}

// TestMain is used for setup before running all tests
func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()

	// Exit with the same code
	os.Exit(exitCode)
}
