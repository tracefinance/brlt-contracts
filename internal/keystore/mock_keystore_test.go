package keystore

import (
	"context"
	"testing"

	"vault0/internal/keygen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockKeyStore(t *testing.T) {
	mockKeyStore := NewMockKeyStore()
	ctx := context.Background()

	t.Run("ImplementsKeyStoreInterface", func(t *testing.T) {
		// This is a compile-time check to ensure MockKeyStore implements KeyStore
		var _ KeyStore = (*MockKeyStore)(nil)
	})

	t.Run("Create", func(t *testing.T) {
		// Create a key
		key, err := mockKeyStore.Create(ctx, "test-key", "Test Key", keygen.KeyTypeECDSA, nil)
		require.NoError(t, err)
		assert.Equal(t, "test-key", key.ID)
		assert.Equal(t, "Test Key", key.Name)
		assert.Equal(t, keygen.KeyTypeECDSA, key.Type)
		assert.NotNil(t, key.PublicKey)
		assert.Nil(t, key.PrivateKey) // Public method should not expose private key
	})

	t.Run("GetPublicKey", func(t *testing.T) {
		// Get a key
		key, err := mockKeyStore.GetPublicKey(ctx, "test-key")
		require.NoError(t, err)
		assert.Equal(t, "test-key", key.ID)
		assert.Equal(t, "Test Key", key.Name)
		assert.Nil(t, key.PrivateKey) // Should never expose private key
	})

	t.Run("GetPublicKey_NotFound", func(t *testing.T) {
		// Get a non-existent key
		_, err := mockKeyStore.GetPublicKey(ctx, "non-existent")
		assert.Error(t, err)
		assert.Equal(t, ErrKeyNotFound, err)
	})

	t.Run("Sign", func(t *testing.T) {
		// Sign data
		signature, err := mockKeyStore.Sign(ctx, "test-key", []byte("test data"))
		require.NoError(t, err)
		assert.NotEmpty(t, signature)
	})

	t.Run("List", func(t *testing.T) {
		// Create a second key
		_, err := mockKeyStore.Create(ctx, "test-key-2", "Test Key 2", keygen.KeyTypeRSA, nil)
		require.NoError(t, err)

		// List keys
		keys, err := mockKeyStore.List(ctx)
		require.NoError(t, err)
		assert.Len(t, keys, 2)

		// Verify keys don't have private key material
		for _, key := range keys {
			assert.Nil(t, key.PrivateKey)
		}
	})

	t.Run("Update", func(t *testing.T) {
		// Update a key
		updatedKey, err := mockKeyStore.Update(ctx, "test-key", "Updated Name", map[string]string{"tag": "value"})
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updatedKey.Name)
		assert.Equal(t, map[string]string{"tag": "value"}, updatedKey.Tags)
	})

	t.Run("Delete", func(t *testing.T) {
		// Delete a key
		err := mockKeyStore.Delete(ctx, "test-key")
		require.NoError(t, err)

		// Verify key is deleted
		_, err = mockKeyStore.GetPublicKey(ctx, "test-key")
		assert.Error(t, err)
		assert.Equal(t, ErrKeyNotFound, err)

		// List should only have one key now
		keys, err := mockKeyStore.List(ctx)
		require.NoError(t, err)
		assert.Len(t, keys, 1)
	})

	t.Run("Import", func(t *testing.T) {
		// Generate key material for import
		keyGen := keygen.NewKeyGenerator()
		privateKey, publicKey, err := keyGen.GenerateKeyPair(keygen.KeyTypeEd25519)
		require.NoError(t, err)

		// Import the key
		importedKey, err := mockKeyStore.Import(ctx, "imported-key", "Imported Key", keygen.KeyTypeEd25519, privateKey, publicKey, nil)
		require.NoError(t, err)
		assert.Equal(t, "imported-key", importedKey.ID)
		assert.Equal(t, "Imported Key", importedKey.Name)
		assert.Equal(t, keygen.KeyTypeEd25519, importedKey.Type)

		// Verify we can sign with the imported key
		signature, err := mockKeyStore.Sign(ctx, "imported-key", []byte("test data"))
		require.NoError(t, err)
		assert.NotEmpty(t, signature)
	})

	t.Run("Close", func(t *testing.T) {
		// Close should not return an error
		err := mockKeyStore.Close()
		assert.NoError(t, err)
	})
}
