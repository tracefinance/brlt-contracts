package crypto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESEncryptor(t *testing.T) {
	t.Run("Create new encryptor with valid key sizes", func(t *testing.T) {
		// Test AES-128
		key128, err := GenerateEncryptionKey(16)
		require.NoError(t, err)
		encryptor128, err := NewAESEncryptor(key128)
		assert.NoError(t, err)
		assert.NotNil(t, encryptor128)

		// Test AES-192
		key192, err := GenerateEncryptionKey(24)
		require.NoError(t, err)
		encryptor192, err := NewAESEncryptor(key192)
		assert.NoError(t, err)
		assert.NotNil(t, encryptor192)

		// Test AES-256
		key256, err := GenerateEncryptionKey(32)
		require.NoError(t, err)
		encryptor256, err := NewAESEncryptor(key256)
		assert.NoError(t, err)
		assert.NotNil(t, encryptor256)
	})

	t.Run("Create new encryptor with invalid key size", func(t *testing.T) {
		invalidKey := make([]byte, 20)
		encryptor, err := NewAESEncryptor(invalidKey)
		assert.Error(t, err)
		assert.Nil(t, encryptor)
		assert.ErrorIs(t, err, ErrInvalidEncryptionKey)
	})

	t.Run("Create new encryptor from base64", func(t *testing.T) {
		key, err := GenerateEncryptionKeyBase64(32)
		require.NoError(t, err)
		encryptor, err := NewAESEncryptorFromBase64(key)
		assert.NoError(t, err)
		assert.NotNil(t, encryptor)
	})

	t.Run("Create new encryptor from invalid base64", func(t *testing.T) {
		encryptor, err := NewAESEncryptorFromBase64("invalid-base64")
		assert.Error(t, err)
		assert.Nil(t, encryptor)
		assert.ErrorIs(t, err, ErrInvalidEncryptionKey)
	})

	t.Run("Encrypt and decrypt data", func(t *testing.T) {
		// Create encryptor
		key, err := GenerateEncryptionKey(32)
		require.NoError(t, err)
		encryptor, err := NewAESEncryptor(key)
		require.NoError(t, err)

		// Test data
		plaintext := []byte("Hello, World!")

		// Encrypt
		ciphertext, err := encryptor.Encrypt(plaintext)
		assert.NoError(t, err)
		assert.NotNil(t, ciphertext)
		assert.NotEqual(t, plaintext, ciphertext)

		// Decrypt
		decrypted, err := encryptor.Decrypt(ciphertext)
		assert.NoError(t, err)
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("Decrypt invalid ciphertext", func(t *testing.T) {
		// Create encryptor
		key, err := GenerateEncryptionKey(32)
		require.NoError(t, err)
		encryptor, err := NewAESEncryptor(key)
		require.NoError(t, err)

		// Test with invalid ciphertext
		invalidCiphertext := []byte("invalid-ciphertext")
		decrypted, err := encryptor.Decrypt(invalidCiphertext)
		assert.Error(t, err)
		assert.Nil(t, decrypted)
		assert.ErrorIs(t, err, ErrDecryptionError)
	})

	t.Run("Generate encryption key with invalid size", func(t *testing.T) {
		key, err := GenerateEncryptionKey(15)
		assert.Error(t, err)
		assert.Nil(t, key)
		assert.ErrorIs(t, err, ErrInvalidEncryptionKey)
	})

	t.Run("Generate base64 encryption key with invalid size", func(t *testing.T) {
		key, err := GenerateEncryptionKeyBase64(15)
		assert.Error(t, err)
		assert.Empty(t, key)
		assert.ErrorIs(t, err, ErrInvalidEncryptionKey)
	})
}
