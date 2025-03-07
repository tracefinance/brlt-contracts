package keystore

import (
	"context"
	"database/sql"
	"testing"

	"vault0/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFactory_Create(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create a test config
	cfg := testConfig()

	// Create a factory
	factory := NewFactory(cfg, db)

	t.Run("CreateDBKeyStore", func(t *testing.T) {
		// Create a DB keystore
		keystore, err := factory.Create(KeyStoreTypeDB)
		require.NoError(t, err)

		// Validate it's a DBKeyStore
		_, ok := keystore.(*DBKeyStore)
		assert.True(t, ok, "Keystore should be a DBKeyStore")
	})

	t.Run("CreateDefaultKeyStore", func(t *testing.T) {
		// Create a default keystore (should be DB)
		keystore, err := factory.CreateDefault()
		require.NoError(t, err)

		// Validate it's a DBKeyStore
		_, ok := keystore.(*DBKeyStore)
		assert.True(t, ok, "Default keystore should be a DBKeyStore")
	})

	t.Run("CreateKMSKeyStore", func(t *testing.T) {
		// Attempt to create a KMS keystore (not implemented yet)
		_, err := factory.Create(KeyStoreTypeKMS)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not implemented yet")
	})

	t.Run("CreateUnsupportedKeyStore", func(t *testing.T) {
		// Attempt to create an unsupported keystore type
		_, err := factory.Create("unsupported")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported key store type")
	})
}

func TestFactory_InvalidConfig(t *testing.T) {
	// Setup test database
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create an invalid config (missing encryption key)
	invalidCfg := &config.Config{
		DBPath: ":memory:",
		// DBEncryptionKey is intentionally missing
	}

	// Create a factory
	factory := NewFactory(invalidCfg, db)

	// Attempt to create a DB keystore with invalid config
	_, err := factory.Create(KeyStoreTypeDB)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "DB_ENCRYPTION_KEY environment variable is required")
}

// Instead of testing for nil DB directly, let's test for db query execution failure
func TestFactory_ErrorHandling(t *testing.T) {
	// Create a valid config
	cfg := testConfig()

	// Create a DB that will be closed before use to simulate errors
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	db.Close() // Close it immediately to force query errors

	// Create a factory with the closed DB
	factory := NewFactory(cfg, db)

	// Create the keystore (should succeed initially)
	keystore, err := factory.Create(KeyStoreTypeDB)
	require.NoError(t, err)

	// But operations on it should fail due to closed DB
	_, err = keystore.List(context.Background())
	assert.Error(t, err)
}
