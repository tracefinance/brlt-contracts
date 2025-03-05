package keygen

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAESEncryptor_EncryptDecrypt(t *testing.T) {
	t.Run("AES128", func(t *testing.T) {
		// Generate a 16-byte key for AES-128
		key, err := GenerateEncryptionKey(16)
		require.NoError(t, err)

		encryptor, err := NewAESEncryptor(key)
		require.NoError(t, err)

		// Test encrypting and decrypting data
		plaintext := []byte("This is a test message")
		ciphertext, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		// Ciphertext should be different from plaintext
		assert.NotEqual(t, plaintext, ciphertext)

		// Decrypt the ciphertext
		decrypted, err := encryptor.Decrypt(ciphertext)
		require.NoError(t, err)

		// Decrypted data should match original plaintext
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("AES256", func(t *testing.T) {
		// Generate a 32-byte key for AES-256
		key, err := GenerateEncryptionKey(32)
		require.NoError(t, err)

		encryptor, err := NewAESEncryptor(key)
		require.NoError(t, err)

		// Test encrypting and decrypting data
		plaintext := []byte("This is a test message for AES-256")
		ciphertext, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		// Ciphertext should be different from plaintext
		assert.NotEqual(t, plaintext, ciphertext)

		// Decrypt the ciphertext
		decrypted, err := encryptor.Decrypt(ciphertext)
		require.NoError(t, err)

		// Decrypted data should match original plaintext
		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("InvalidKeySize", func(t *testing.T) {
		// Try to create an encryptor with an invalid key size
		_, err := NewAESEncryptor([]byte("too-short"))
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidEncryptionKey, err)
	})

	t.Run("Base64KeyEncoding", func(t *testing.T) {
		// Generate a base64-encoded key
		encoded, err := GenerateEncryptionKeyBase64(32)
		require.NoError(t, err)

		// Create an encryptor from the base64 key
		encryptor, err := NewAESEncryptorFromBase64(encoded)
		require.NoError(t, err)

		// Test encrypting and decrypting
		plaintext := []byte("Test with base64 key")
		ciphertext, err := encryptor.Encrypt(plaintext)
		require.NoError(t, err)

		decrypted, err := encryptor.Decrypt(ciphertext)
		require.NoError(t, err)

		assert.Equal(t, plaintext, decrypted)
	})

	t.Run("InvalidBase64Key", func(t *testing.T) {
		// Try to create an encryptor with an invalid base64 key
		_, err := NewAESEncryptorFromBase64("not-valid-base64!@#$")
		assert.Error(t, err)
	})

	t.Run("DecryptionFailure", func(t *testing.T) {
		// Generate a key and encryptor
		key, err := GenerateEncryptionKey(32)
		require.NoError(t, err)

		encryptor, err := NewAESEncryptor(key)
		require.NoError(t, err)

		// Try to decrypt invalid ciphertext
		invalidCiphertext := []byte("not valid ciphertext")
		_, err = encryptor.Decrypt(invalidCiphertext)
		assert.Error(t, err)
		assert.Equal(t, ErrDecryptionError, err)
	})
}

func TestGenerateEncryptionKey(t *testing.T) {
	t.Run("ValidKeySizes", func(t *testing.T) {
		// Test generating keys of valid sizes
		validSizes := []int{16, 24, 32}

		for _, size := range validSizes {
			key, err := GenerateEncryptionKey(size)
			require.NoError(t, err)
			assert.Equal(t, size, len(key))
		}
	})

	t.Run("InvalidKeySize", func(t *testing.T) {
		// Test generating a key with an invalid size
		_, err := GenerateEncryptionKey(15)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidEncryptionKey, err)
	})

	t.Run("Base64Encoding", func(t *testing.T) {
		// Test generating a base64-encoded key
		encoded, err := GenerateEncryptionKeyBase64(32)
		require.NoError(t, err)
		assert.NotEmpty(t, encoded)

		// Should be able to create an encryptor from it
		encryptor, err := NewAESEncryptorFromBase64(encoded)
		require.NoError(t, err)
		assert.NotNil(t, encryptor)
	})
}
