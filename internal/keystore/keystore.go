package keystore

import (
	"context"
	"crypto/elliptic"
	"errors"

	"vault0/internal/keygen"
)

// Common errors
var (
	ErrKeyNotFound      = errors.New("key not found")
	ErrKeyAlreadyExists = errors.New("key already exists")
	ErrInvalidKey       = errors.New("invalid key")
	ErrEncryptionFailed = errors.New("encryption failed")
	ErrDecryptionFailed = errors.New("decryption failed")
	ErrInvalidCurve     = errors.New("invalid curve")
)

// DataType indicates how the data should be treated in cryptographic operations
type DataType string

const (
	// DataTypeRaw indicates that data has not been hashed and requires hashing during operations
	DataTypeRaw DataType = "raw"
	// DataTypeDigest indicates that data is already hashed/digested
	DataTypeDigest DataType = "digest"
)

// Key struct represents a cryptographic key.
type Key struct {
	// ID is the unique identifier for the key
	ID string
	// Name is a human-readable name for the key
	Name string
	// Type is the type of cryptographic key
	Type keygen.KeyType
	// Curve is the elliptic curve used for ECDSA keys
	Curve elliptic.Curve
	// Tags are optional metadata for the key
	Tags map[string]string
	// CreatedAt is the timestamp when the key was created
	CreatedAt int64
	// PrivateKey is the encrypted private key material
	// Note: This field should only be populated during key creation and import operations
	// and should never be exposed outside the keystore package
	PrivateKey []byte
	// PublicKey is the public key material (if applicable)
	PublicKey []byte
}

// KeyStore defines the interface for key management operations
type KeyStore interface {
	// Create creates a new key with the given name and type
	// For ECDSA keys, curve specifies which elliptic curve to use (e.g., P256, P384, P521)
	// For other key types, curve parameter is ignored
	Create(ctx context.Context, name string, keyType keygen.KeyType, curve elliptic.Curve, tags map[string]string) (*Key, error)

	// Import imports an existing key
	// For ECDSA keys, curve must match the curve used to generate the key
	// For other key types, curve parameter is ignored
	Import(ctx context.Context, name string, keyType keygen.KeyType, curve elliptic.Curve, privateKey, publicKey []byte, tags map[string]string) (*Key, error)

	// Sign signs the provided data using the key identified by id
	// This method uses the private key internally without exposing it
	// The dataType parameter indicates whether the data needs to be hashed as part of the signing algorithm
	Sign(ctx context.Context, id string, data []byte, dataType DataType) ([]byte, error)

	// GetPublicKey retrieves only the public part of a key by its ID
	GetPublicKey(ctx context.Context, id string) (*Key, error)

	// List lists all keys
	List(ctx context.Context) ([]*Key, error)

	// Update updates a key's metadata
	Update(ctx context.Context, id string, name string, tags map[string]string) (*Key, error)

	// Delete deletes a key by its ID
	Delete(ctx context.Context, id string) error
}
