package keystore

import (
	"context"
	"fmt"
	"testing"

	"vault0/internal/keygen"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBKeyStore_Create(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Create_Valid_ECDSA_Key", func(t *testing.T) {
		// Arrange
		id := "test-key-1"
		name := "Test ECDSA Key"
		keyType := keygen.KeyTypeECDSA
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, tags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
		assert.True(t, key.CreatedAt > 0)
	})

	t.Run("Create_Valid_RSA_Key", func(t *testing.T) {
		// Arrange
		id := "test-key-2"
		name := "Test RSA Key"
		keyType := keygen.KeyTypeRSA
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, tags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
	})

	t.Run("Create_Valid_Ed25519_Key", func(t *testing.T) {
		// Arrange
		id := "test-key-3"
		name := "Test Ed25519 Key"
		keyType := keygen.KeyTypeEd25519
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, tags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
	})

	t.Run("Create_Valid_Symmetric_Key", func(t *testing.T) {
		// Arrange
		id := "test-key-4"
		name := "Test Symmetric Key"
		keyType := keygen.KeyTypeSymmetric
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, tags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.Nil(t, key.PublicKey)
	})

	t.Run("Create_DuplicateID", func(t *testing.T) {
		// Arrange
		id := "test-key-duplicate"
		name := "Test Duplicate Key"
		keyType := keygen.KeyTypeECDSA
		tags := map[string]string{"purpose": "testing"}

		// First creation should succeed
		_, err := keystore.Create(ctx, id, name, keyType, tags)
		require.NoError(t, err)

		// Act - Second creation with same ID should fail
		_, err = keystore.Create(ctx, id, "Another Key", keyType, tags)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrKeyAlreadyExists, err)
	})
}

func TestDBKeyStore_GetPublicKey(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("GetPublicKey_ExistingKey", func(t *testing.T) {
		// Arrange - Create a key
		id := "test-key-get"
		name := "Test Get Key"
		keyType := keygen.KeyTypeECDSA
		tags := map[string]string{"purpose": "testing"}

		originalKey, err := keystore.Create(ctx, id, name, keyType, tags)
		require.NoError(t, err)

		// Act
		retrievedKey, err := keystore.GetPublicKey(ctx, id)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, originalKey.ID, retrievedKey.ID)
		assert.Equal(t, originalKey.Name, retrievedKey.Name)
		assert.Equal(t, originalKey.Type, retrievedKey.Type)
		assert.Equal(t, originalKey.Tags, retrievedKey.Tags)
		assert.Equal(t, originalKey.CreatedAt, retrievedKey.CreatedAt)
		assert.Equal(t, originalKey.PublicKey, retrievedKey.PublicKey)
		// Private key should never be exposed by GetPublicKey
		assert.Nil(t, retrievedKey.PrivateKey)
	})

	t.Run("GetPublicKey_NonExistentKey", func(t *testing.T) {
		// Act
		_, err := keystore.GetPublicKey(ctx, "non-existent-key")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrKeyNotFound, err)
	})
}

func TestDBKeyStore_List(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("List_EmptyKeyStore", func(t *testing.T) {
		// Act
		keys, err := keystore.List(ctx)

		// Assert
		require.NoError(t, err)
		assert.Empty(t, keys)
	})

	t.Run("List_MultipleKeys", func(t *testing.T) {
		// Arrange - Create a few keys
		keyIDs := []string{"list-key-1", "list-key-2", "list-key-3"}
		for i, id := range keyIDs {
			keyName := fmt.Sprintf("List Key %d", i)
			_, err := keystore.Create(ctx, id, keyName, keygen.KeyTypeECDSA, nil)
			require.NoError(t, err)
		}

		// Act
		keys, err := keystore.List(ctx)

		// Assert
		require.NoError(t, err)
		assert.Len(t, keys, len(keyIDs))

		// Verify the keys are returned with correct IDs and no private keys
		for _, key := range keys {
			assert.Contains(t, keyIDs, key.ID)
			assert.Nil(t, key.PrivateKey)
		}
	})
}

func TestDBKeyStore_Update(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Update_ExistingKey", func(t *testing.T) {
		// Arrange - Create a key
		id := "update-key-1"
		originalName := "Original Name"
		originalTags := map[string]string{"purpose": "original", "env": "test"}

		originalKey, err := keystore.Create(ctx, id, originalName, keygen.KeyTypeECDSA, originalTags)
		require.NoError(t, err)

		// Act - Update the key
		newName := "Updated Name"
		newTags := map[string]string{"purpose": "updated", "env": "test", "new": "tag"}

		updatedKey, err := keystore.Update(ctx, id, newName, newTags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, updatedKey.ID)
		assert.Equal(t, newName, updatedKey.Name)
		assert.Equal(t, newTags, updatedKey.Tags)
		assert.Equal(t, originalKey.Type, updatedKey.Type)
		assert.Equal(t, originalKey.CreatedAt, updatedKey.CreatedAt)
		assert.Equal(t, originalKey.PublicKey, updatedKey.PublicKey)
	})

	t.Run("Update_NonExistentKey", func(t *testing.T) {
		// Act
		_, err := keystore.Update(ctx, "non-existent-key", "New Name", nil)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrKeyNotFound, err)
	})
}

func TestDBKeyStore_Delete(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Delete_ExistingKey", func(t *testing.T) {
		// Arrange - Create a key
		id := "delete-key-1"
		_, err := keystore.Create(ctx, id, "Delete Key", keygen.KeyTypeECDSA, nil)
		require.NoError(t, err)

		// Act - Delete the key
		err = keystore.Delete(ctx, id)

		// Assert
		require.NoError(t, err)

		// Verify the key is deleted
		_, err = keystore.GetPublicKey(ctx, id)
		assert.Error(t, err)
		assert.Equal(t, ErrKeyNotFound, err)
	})

	t.Run("Delete_NonExistentKey", func(t *testing.T) {
		// Act
		err := keystore.Delete(ctx, "non-existent-key")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrKeyNotFound, err)
	})
}

func TestDBKeyStore_Sign(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()
	data := []byte("test data to sign")

	tests := []struct {
		name    string
		keyType keygen.KeyType
		wantErr bool
	}{
		{
			name:    "ECDSA",
			keyType: keygen.KeyTypeECDSA,
			wantErr: false,
		},
		{
			name:    "RSA",
			keyType: keygen.KeyTypeRSA,
			wantErr: false,
		},
		{
			name:    "Ed25519",
			keyType: keygen.KeyTypeEd25519,
			wantErr: false,
		},
		{
			name:    "Symmetric",
			keyType: keygen.KeyTypeSymmetric,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a key
			keyID := "test-sign-key-" + tt.name
			key, err := keystore.Create(ctx, keyID, "Test Signing Key "+tt.name, tt.keyType, map[string]string{"purpose": "signing"})
			require.NoError(t, err)
			require.NotNil(t, key)

			// Sign data
			signature, err := keystore.Sign(ctx, keyID, data)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, signature)
			require.NotEmpty(t, signature)
		})
	}
}

func TestDBKeyStore_Import(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	// Generate a key pair for import
	keyGen := keygen.NewKeyGenerator()
	privateKey, publicKey, err := keyGen.GenerateKeyPair(keygen.KeyTypeECDSA, nil)
	require.NoError(t, err)

	// Import the key
	keyID := "imported-key"
	keyName := "Imported ECDSA Key"
	keyType := keygen.KeyTypeECDSA
	tags := map[string]string{"purpose": "imported"}

	key, err := keystore.Import(ctx, keyID, keyName, keyType, privateKey, publicKey, tags)

	require.NoError(t, err)
	require.NotNil(t, key)
	require.Equal(t, keyID, key.ID)
	require.Equal(t, keyName, key.Name)
	require.Equal(t, keyType, key.Type)
	require.Equal(t, tags, key.Tags)
}
