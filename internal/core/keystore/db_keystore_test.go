package keystore

import (
	"context"
	"crypto/elliptic"
	"crypto/sha256"
	"testing"

	"vault0/internal/core/crypto"
	"vault0/internal/core/keygen"
	"vault0/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDBKeyStore_Create(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Create_Valid_ECDSA_P256_Key", func(t *testing.T) {
		// Arrange
		name := "Test ECDSA P-256 Key"
		keyType := types.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, name, keyType, curve, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
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
		name := "Test ECDSA secp256k1 Key"
		keyType := types.KeyTypeECDSA
		curve := crypto.Secp256k1Curve
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, name, keyType, curve, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
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
		name := "Test ECDSA Key"
		keyType := types.KeyTypeECDSA
		curve := elliptic.P256() // Default to P-256 for backward compatibility
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, name, keyType, curve, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
		assert.True(t, key.CreatedAt > 0)
	})

	t.Run("Create_Valid_RSA_Key", func(t *testing.T) {
		// Arrange
		name := "Test RSA Key"
		keyType := types.KeyTypeRSA
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, name, keyType, nil, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
	})

	t.Run("Create_Valid_Ed25519_Key", func(t *testing.T) {
		// Arrange
		name := "Test Ed25519 Key"
		keyType := types.KeyTypeEd25519
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, name, keyType, nil, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.NotNil(t, key.PublicKey)
		assert.NotEmpty(t, key.PublicKey)
	})

	t.Run("Create_Valid_Symmetric_Key", func(t *testing.T) {
		// Arrange
		name := "Test Symmetric Key"
		keyType := types.KeyTypeSymmetric
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, name, keyType, nil, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.Nil(t, key.PublicKey)
	})

	t.Run("Create_DuplicateName", func(t *testing.T) {
		// Arrange
		name := "Test Duplicate Name Key"
		keyType := types.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		// First creation should succeed
		_, err := keystore.Create(ctx, name, keyType, curve, tags)
		require.NoError(t, err)

		// Act - Second creation with same name should fail
		_, err = keystore.Create(ctx, name, keyType, curve, tags)

		// Assert
		assert.Error(t, err)
		// The error may not be exactly ErrKeyAlreadyExists since SQLite will return a constraint violation
		assert.Equal(t, ErrKeyAlreadyExists, err)
	})

	t.Run("Create_ECDSA_NilCurve", func(t *testing.T) {
		// Arrange
		name := "Test ECDSA Key with Nil Curve"
		keyType := types.KeyTypeECDSA
		tags := map[string]string{"purpose": "testing"}

		// Act
		key, err := keystore.Create(ctx, name, keyType, nil, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
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
		name := "Test Get P-256 Key"
		keyType := types.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		originalKey, err := keystore.Create(ctx, name, keyType, curve, tags)
		require.NoError(t, err)

		// Act
		retrievedKey, err := keystore.GetPublicKey(ctx, originalKey.ID)

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
		name := "Test Get secp256k1 Key"
		keyType := types.KeyTypeECDSA
		curve := crypto.Secp256k1Curve
		tags := map[string]string{"purpose": "testing"}

		originalKey, err := keystore.Create(ctx, name, keyType, curve, tags)
		require.NoError(t, err)

		// Act
		retrievedKey, err := keystore.GetPublicKey(ctx, originalKey.ID)

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
		name := "Test Get Key"
		keyType := types.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing"}

		originalKey, err := keystore.Create(ctx, name, keyType, curve, tags)
		require.NoError(t, err)

		// Act
		retrievedKey, err := keystore.GetPublicKey(ctx, originalKey.ID)

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
		keyNames := []string{"list-key-1", "list-key-2", "list-key-3"}
		for _, name := range keyNames {
			_, err := keystore.Create(ctx, name, types.KeyTypeECDSA, elliptic.P256(), nil)
			require.NoError(t, err)
		}

		// Act
		keys, err := keystore.List(ctx)

		// Assert
		require.NoError(t, err)
		assert.Len(t, keys, len(keyNames))

		// Verify the keys are returned with correct IDs and no private keys
		for _, key := range keys {
			assert.Contains(t, keyNames, key.Name)
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
		name := "Original Name"
		originalTags := map[string]string{"purpose": "original", "env": "test"}

		originalKey, err := keystore.Create(ctx, name, types.KeyTypeECDSA, elliptic.P256(), originalTags)
		require.NoError(t, err)

		// Act - Update the key
		newName := "Updated Name"
		newTags := map[string]string{"purpose": "updated", "env": "test", "new": "tag"}

		updatedKey, err := keystore.Update(ctx, originalKey.ID, newName, newTags)

		// Assert
		require.NoError(t, err)
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
		name := "Delete Key"
		key, err := keystore.Create(ctx, name, types.KeyTypeECDSA, elliptic.P256(), nil)
		require.NoError(t, err)

		// Act - Delete the key
		err = keystore.Delete(ctx, key.ID)

		// Assert
		require.NoError(t, err)

		// Verify the key is deleted
		_, err = keystore.GetPublicKey(ctx, key.ID)
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
		keyType  types.KeyType
		curve    elliptic.Curve
		dataType DataType
		testData []byte
		wantErr  bool
	}{
		{
			name:     "ECDSA_Raw",
			keyType:  types.KeyTypeECDSA,
			curve:    elliptic.P256(),
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "ECDSA_Digest",
			keyType:  types.KeyTypeECDSA,
			curve:    elliptic.P256(),
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
		{
			name:     "RSA_Raw",
			keyType:  types.KeyTypeRSA,
			curve:    nil,
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "RSA_Digest",
			keyType:  types.KeyTypeRSA,
			curve:    nil,
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
		{
			name:     "Ed25519_Raw",
			keyType:  types.KeyTypeEd25519,
			curve:    nil,
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "Ed25519_Digest",
			keyType:  types.KeyTypeEd25519,
			curve:    nil,
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
		{
			name:     "Symmetric_Raw",
			keyType:  types.KeyTypeSymmetric,
			curve:    nil,
			dataType: DataTypeRaw,
			testData: data,
			wantErr:  false,
		},
		{
			name:     "Symmetric_Digest",
			keyType:  types.KeyTypeSymmetric,
			curve:    nil,
			dataType: DataTypeDigest,
			testData: hashedData,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a key for this test
			name := "Test Signing Key " + tt.name
			key, err := keystore.Create(ctx, name, tt.keyType, tt.curve, map[string]string{"purpose": "signing"})
			require.NoError(t, err)
			require.NotNil(t, key)

			// Sign data
			signature, err := keystore.Sign(ctx, key.ID, tt.testData, tt.dataType)

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
		name := "Test Pre-hashed Signing"
		key, err := keystore.Create(ctx, name, types.KeyTypeECDSA, elliptic.P256(), nil)
		require.NoError(t, err)
		require.NotNil(t, key)

		// Hash the data manually
		hasher := sha256.New()
		hasher.Write(data)
		hashedData := hasher.Sum(nil)

		// Sign the pre-hashed data
		signature, err := keystore.Sign(ctx, key.ID, hashedData, DataTypeDigest)
		require.NoError(t, err)
		require.NotNil(t, signature)
		require.NotEmpty(t, signature)

		// Also test signing the raw data for comparison
		signatureRaw, err := keystore.Sign(ctx, key.ID, data, DataTypeRaw)
		require.NoError(t, err)
		require.NotNil(t, signatureRaw)
		require.NotEmpty(t, signatureRaw)

		// The signatures should be different because the input to the signing algorithm is different
		require.NotEqual(t, signature, signatureRaw)
	})

	t.Run("Sign_ECDSA_P256", func(t *testing.T) {
		// Arrange - Create a P-256 key
		name := "Test P-256 Key"
		key, err := keystore.Create(ctx, name, types.KeyTypeECDSA, elliptic.P256(), nil)
		require.NoError(t, err)

		// Test data
		data := []byte("test data")
		hash := sha256.Sum256(data)

		// Act - Sign raw data
		signature1, err := keystore.Sign(ctx, key.ID, data, DataTypeRaw)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		// Act - Sign pre-hashed data
		signature2, err := keystore.Sign(ctx, key.ID, hash[:], DataTypeDigest)
		require.NoError(t, err)
		assert.NotEmpty(t, signature2)
	})

	t.Run("Sign_ECDSA_Secp256k1", func(t *testing.T) {
		// Arrange - Create a secp256k1 key
		name := "Test secp256k1 Key"
		key, err := keystore.Create(ctx, name, types.KeyTypeECDSA, crypto.Secp256k1Curve, nil)
		require.NoError(t, err)

		// Test data
		data := []byte("test data")
		hash := sha256.Sum256(data)

		// Act - Sign raw data
		signature1, err := keystore.Sign(ctx, key.ID, data, DataTypeRaw)
		require.NoError(t, err)
		assert.NotEmpty(t, signature1)

		// Act - Sign pre-hashed data
		signature2, err := keystore.Sign(ctx, key.ID, hash[:], DataTypeDigest)
		require.NoError(t, err)
		assert.NotEmpty(t, signature2)
	})
}

func TestDBKeyStore_Import(t *testing.T) {
	keystore, _, cleanup := setupTestKeyStore(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("Import_Valid_ECDSA_Key", func(t *testing.T) {
		// Arrange
		name := "Imported ECDSA Key"
		keyType := types.KeyTypeECDSA
		curve := elliptic.P256()
		tags := map[string]string{"purpose": "testing", "source": "import"}

		// Generate a keypair to import
		privateKey, publicKey, err := generateTestKey(keyType, curve)
		require.NoError(t, err)

		// Act
		key, err := keystore.Import(ctx, name, keyType, curve, privateKey, publicKey, tags)

		// Assert
		require.NoError(t, err)
		assert.NotEmpty(t, key.ID)
		assert.Equal(t, name, key.Name)
		assert.Equal(t, keyType, key.Type)
		assert.Equal(t, tags, key.Tags)
		assert.Equal(t, publicKey, key.PublicKey)
	})
}

// Helper function to generate a test key pair
func generateTestKey(keyType types.KeyType, curve elliptic.Curve) ([]byte, []byte, error) {
	keyGen := keygen.NewKeyGenerator()
	privateKey, publicKey, err := keyGen.GenerateKeyPair(keyType, curve)
	if err != nil {
		return nil, nil, err
	}
	return privateKey, publicKey, nil
}
