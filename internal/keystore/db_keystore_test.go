package keystore

import (
	"context"
	"crypto/elliptic"
	"crypto/sha256"
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

	t.Run("Create_Valid_ECDSA_P256_Key", func(t *testing.T) {
		// Arrange
		id := "test-key-p256"
		name := "Test ECDSA P-256 Key"
		keyType := keygen.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, curve, tags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
		assert.True(t, key.CreatedAt > 0)
		assert.Equal(t, curve, key.Curve)
	})

	t.Run("Create_Valid_ECDSA_Secp256k1_Key", func(t *testing.T) {
		// Arrange
		id := "test-key-secp256k1"
		name := "Test ECDSA secp256k1 Key"
		keyType := keygen.KeyTypeECDSA
		curve := keygen.Secp256k1Curve
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, curve, tags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
		assert.True(t, key.CreatedAt > 0)
		assert.Equal(t, curve, key.Curve)
	})

	t.Run("Create_Valid_ECDSA_Key", func(t *testing.T) {
		// Arrange
		id := "test-key-1"
		name := "Test ECDSA Key"
		keyType := keygen.KeyTypeECDSA
		curve := elliptic.P256() // Default to P-256 for backward compatibility
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, curve, tags)

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
		key, err := keystore.Create(ctx, id, name, keyType, nil, tags)

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
		key, err := keystore.Create(ctx, id, name, keyType, nil, tags)

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
		key, err := keystore.Create(ctx, id, name, keyType, nil, tags)

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
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		// First creation should succeed
		_, err := keystore.Create(ctx, id, name, keyType, curve, tags)
		require.NoError(t, err)

		// Act - Second creation with same ID should fail
		_, err = keystore.Create(ctx, id, "Another Key", keyType, curve, tags)

		// Assert
		assert.Error(t, err)
		assert.Equal(t, ErrKeyAlreadyExists, err)
	})

	t.Run("Create_ECDSA_NilCurve", func(t *testing.T) {
		// Arrange
		id := "test-key-nil-curve"
		name := "Test ECDSA Key with Nil Curve"
		keyType := keygen.KeyTypeECDSA
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, id, name, keyType, nil, tags)

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, key)
		assert.Equal(t, elliptic.P256(), key.Curve) // Should default to P-256
	})
}

func TestDBKeyStore_GetPublicKey(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("GetPublicKey_ECDSA_P256", func(t *testing.T) {
		// Arrange - Create a P-256 key
		id := "test-key-get-p256"
		name := "Test Get P-256 Key"
		keyType := keygen.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		originalKey, err := keystore.Create(ctx, id, name, keyType, curve, tags)
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
		assert.Equal(t, curve, retrievedKey.Curve)
		assert.Nil(t, retrievedKey.PrivateKey)
	})

	t.Run("GetPublicKey_ECDSA_Secp256k1", func(t *testing.T) {
		// Arrange - Create a secp256k1 key
		id := "test-key-get-secp256k1"
		name := "Test Get secp256k1 Key"
		keyType := keygen.KeyTypeECDSA
		curve := keygen.Secp256k1Curve
		tags := map[string]string{"purpose": "testing"}

		originalKey, err := keystore.Create(ctx, id, name, keyType, curve, tags)
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
		assert.Equal(t, curve, retrievedKey.Curve)
		assert.Nil(t, retrievedKey.PrivateKey)
	})

	t.Run("GetPublicKey_ExistingKey", func(t *testing.T) {
		// Arrange - Create a key
		id := "test-key-get"
		name := "Test Get Key"
		keyType := keygen.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		originalKey, err := keystore.Create(ctx, id, name, keyType, curve, tags)
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
			_, err := keystore.Create(ctx, id, keyName, keygen.KeyTypeECDSA, elliptic.P256(), nil)
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

		originalKey, err := keystore.Create(ctx, id, originalName, keygen.KeyTypeECDSA, elliptic.P256(), originalTags)
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
		_, err := keystore.Create(ctx, id, "Delete Key", keygen.KeyTypeECDSA, elliptic.P256(), nil)
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

	// Pre-hash data for digest tests
	hasher := sha256.New()
	hasher.Write(data)
	hashedData := hasher.Sum(nil)

	tests := []struct {
		name     string
		keyType  keygen.KeyType
		curve    elliptic.Curve
		dataType DataType
		testData []byte
		wantErr  bool
	}{
		{
			name:     "ECDSA_Raw",
			keyType:  keygen.KeyTypeECDSA,
			curve:    elliptic.P256(),
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "ECDSA_Digest",
			keyType:  keygen.KeyTypeECDSA,
			curve:    elliptic.P256(),
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
		{
			name:     "RSA_Raw",
			keyType:  keygen.KeyTypeRSA,
			curve:    nil,
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "RSA_Digest",
			keyType:  keygen.KeyTypeRSA,
			curve:    nil,
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
		{
			name:     "Ed25519_Raw",
			keyType:  keygen.KeyTypeEd25519,
			curve:    nil,
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "Ed25519_Digest",
			keyType:  keygen.KeyTypeEd25519,
			curve:    nil,
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
		{
			name:     "Symmetric_Raw",
			keyType:  keygen.KeyTypeSymmetric,
			curve:    nil,
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "Symmetric_Digest",
			keyType:  keygen.KeyTypeSymmetric,
			curve:    nil,
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a key
			keyID := "test-sign-key-" + tt.name
			key, err := keystore.Create(ctx, keyID, "Test Signing Key "+tt.name, tt.keyType, tt.curve, map[string]string{"purpose": "signing"})
			require.NoError(t, err)
			require.NotNil(t, key)

			// Sign data
			signature, err := keystore.Sign(ctx, keyID, tt.testData, tt.dataType)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, signature)
			require.NotEmpty(t, signature)
		})
	}

	t.Run("NonExistentKey", func(t *testing.T) {
		// Try to sign with a non-existent key
		_, err := keystore.Sign(ctx, "non-existent-key", data, DataTypeRaw)
		require.Error(t, err)
	})

	// Test with pre-hashed data specifically
	t.Run("PreHashedData", func(t *testing.T) {
		// Create an ECDSA key
		keyID := "test-sign-prehashed"
		key, err := keystore.Create(ctx, keyID, "Test Pre-hashed Signing", keygen.KeyTypeECDSA, elliptic.P256(), nil)
		require.NoError(t, err)
		require.NotNil(t, key)

		// Hash the data manually
		hasher := sha256.New()
		hasher.Write(data)
		hashedData := hasher.Sum(nil)

		// Sign the pre-hashed data
		signature, err := keystore.Sign(ctx, keyID, hashedData, DataTypeDigest)
		require.NoError(t, err)
		require.NotNil(t, signature)
		require.NotEmpty(t, signature)

		// Also test signing the raw data for comparison
		signatureRaw, err := keystore.Sign(ctx, keyID, data, DataTypeRaw)
		require.NoError(t, err)
		require.NotNil(t, signatureRaw)
		require.NotEmpty(t, signatureRaw)

		// The signatures should be different because the input to the signing algorithm is different
		require.NotEqual(t, signature, signatureRaw)
	})

	t.Run("Sign_ECDSA_P256", func(t *testing.T) {
		// Arrange - Create a P-256 key
		id := "sign-key-p256"
		_, err := keystore.Create(ctx, id, "Test P-256 Key", keygen.KeyTypeECDSA, elliptic.P256(), nil)
		require.NoError(t, err)

		// Test data
		data := []byte("test data")
		hash := sha256.Sum256(data)

		// Act - Sign raw data
		signature1, err := keystore.Sign(ctx, id, data, DataTypeRaw)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		// Act - Sign pre-hashed data
		signature2, err := keystore.Sign(ctx, id, hash[:], DataTypeDigest)
		require.NoError(t, err)
		assert.NotEmpty(t, signature2)
	})

	t.Run("Sign_ECDSA_Secp256k1", func(t *testing.T) {
		// Arrange - Create a secp256k1 key
		id := "sign-key-secp256k1"
		_, err := keystore.Create(ctx, id, "Test secp256k1 Key", keygen.KeyTypeECDSA, keygen.Secp256k1Curve, nil)
		require.NoError(t, err)

		// Test data
		data := []byte("test data")
		hash := sha256.Sum256(data)

		// Act - Sign raw data
		signature1, err := keystore.Sign(ctx, id, data, DataTypeRaw)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		// Act - Sign pre-hashed data
		signature2, err := keystore.Sign(ctx, id, hash[:], DataTypeDigest)
		require.NoError(t, err)
		assert.NotEmpty(t, signature2)
	})
}

func TestDBKeyStore_Import(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Import_ECDSA_Key", func(t *testing.T) {
		// Arrange
		id := "import-key-1"
		name := "Imported ECDSA Key"
		keyType := keygen.KeyTypeECDSA
		curve := elliptic.P256()
		privateKey := []byte("test-private-key")
		publicKey := []byte("test-public-key")
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Import(ctx, id, name, keyType, curve, privateKey, publicKey, tags)

		// Assert
		require.NoError(t, err)
		assert.Equal(t, id, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.Equal(t, publicKey, key.PublicKey)
	})
}
