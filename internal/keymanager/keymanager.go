package keymanager

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrKeyAlreadyExists = errors.New("key already exists")
	ErrInvalidKey       = errors.New("invalid key")
	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")
)

// KeyType represents the type of cryptographic key
type KeyType string

// Supported key types
const (
	KeyTypeECDSA     KeyType = "ECDSA"
	KeyTypeRSA       KeyType = "RSA"
	KeyTypeEd25519   KeyType = "Ed25519"
	KeyTypeSymmetric KeyType = "Symmetric"
)

// Key represents a cryptographic key stored in the key management system
type Key struct {
	// ID is the unique identifier for the key
	ID string
	// Name is a human-readable name for the key
	Name string
	// Type is the type of cryptographic key
	Type KeyType
	// Tags are optional metadata for the key
	Tags map[string]string
	// CreatedAt is the timestamp when the key was created
	CreatedAt int64
	// PrivateKey is the encrypted private key material (empty when retrieved without secret access)
	PrivateKey []byte
	// PublicKey is the public key material (if applicable)
	PublicKey []byte
}

// KeyManager defines the interface for key management operations
type KeyManager interface {
	// Create creates a new key with the given ID, name, and type
	Create(ctx context.Context, id, name string, keyType KeyType, tags map[string]string) (*Key, error)

	// Import imports an existing key
	Import(ctx context.Context, id, name string, keyType KeyType, privateKey, publicKey []byte, tags map[string]string) (*Key, error)

	// Get retrieves a key by its ID
	Get(ctx context.Context, id string) (*Key, error)

	// GetPublicOnly retrieves only the public part of a key by its ID
	GetPublicOnly(ctx context.Context, id string) (*Key, error)

	// List lists all keys
	List(ctx context.Context) ([]*Key, error)

	// Update updates a key's metadata
	Update(ctx context.Context, id string, name string, tags map[string]string) (*Key, error)

	// Delete deletes a key by its ID
	Delete(ctx context.Context, id string) error

	// Close releases any resources used by the key manager
	Close() error
}
