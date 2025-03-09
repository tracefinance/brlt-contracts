package keystore

import (
	"database/sql"
	"testing"

	"vault0/internal/config"
	"vault0/internal/crypto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

// testConfig creates a config for testing purposes
func testConfig() *config.Config {
	// Generate a random encryption key for tests
	encKey, _ := crypto.GenerateEncryptionKeyBase64(32)

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
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			key_type TEXT NOT NULL,
			curve TEXT,
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
		// No need to close the keystore anymore
		dbCleanup()
	}

	return dbKeyStore, db, cleanup
}

func TestConfig(t *testing.T) {
	t.Run("Create config with valid encryption key", func(t *testing.T) {
		// Generate a valid encryption key
		key, err := crypto.GenerateEncryptionKeyBase64(32)
		assert.NoError(t, err)

		// Create config
		cfg := &config.Config{
			DBPath:          ":memory:",
			DBEncryptionKey: key,
		}

		assert.NotNil(t, cfg)
		assert.Equal(t, ":memory:", cfg.DBPath)
		assert.Equal(t, key, cfg.DBEncryptionKey)
	})
}
