package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"

	"vault0/internal/errors"
)

// Encryptor defines the interface for encryption and decryption operations
type Encryptor interface {
	// Encrypt encrypts the plaintext data
	Encrypt(plaintext []byte) ([]byte, error)

	// Decrypt decrypts the ciphertext data
	Decrypt(ciphertext []byte) ([]byte, error)
}

// AESEncryptor implements the Encryptor interface using AES-GCM
type AESEncryptor struct {
	key []byte
}

// NewAESEncryptor creates a new AES-GCM encryptor with the provided key
// The key must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256
func NewAESEncryptor(key []byte) (*AESEncryptor, error) {
	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, errors.NewInvalidEncryptionKeyError("key length must be 16, 24, or 32 bytes")
	}
	return &AESEncryptor{key: key}, nil
}

// NewAESEncryptorFromBase64 creates a new AES-GCM encryptor with a base64 encoded key
func NewAESEncryptorFromBase64(encodedKey string) (*AESEncryptor, error) {
	key, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, errors.NewInvalidEncryptionKeyError("invalid base64 encoded key")
	}
	return NewAESEncryptor(key)
}

// Encrypt encrypts the plaintext data using AES-GCM
func (e *AESEncryptor) Encrypt(plaintext []byte) ([]byte, error) {
	// Create a new AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, errors.NewEncryptionError(err)
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.NewEncryptionError(err)
	}

	// Create a nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.NewEncryptionError(err)
	}

	// Encrypt the plaintext
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// Decrypt decrypts the ciphertext data using AES-GCM
func (e *AESEncryptor) Decrypt(ciphertext []byte) ([]byte, error) {
	// Create a new AES cipher block
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return nil, errors.NewDecryptionError(err)
	}

	// Create a new GCM cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.NewDecryptionError(err)
	}

	// Get the nonce size
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.NewDecryptionError(err)
	}

	// Extract the nonce and ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the ciphertext
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.NewDecryptionError(err)
	}

	return plaintext, nil
}

// GenerateEncryptionKey generates a random encryption key of the specified size in bytes
func GenerateEncryptionKey(size int) ([]byte, error) {
	if size != 16 && size != 24 && size != 32 {
		return nil, errors.NewInvalidEncryptionKeyError("key size must be 16, 24, or 32 bytes")
	}

	key := make([]byte, size)
	_, err := io.ReadFull(rand.Reader, key)
	if err != nil {
		return nil, errors.NewCryptoError(err)
	}

	return key, nil
}

// GenerateEncryptionKeyBase64 generates a random encryption key encoded as base64
func GenerateEncryptionKeyBase64(size int) (string, error) {
	key, err := GenerateEncryptionKey(size)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(key), nil
}
