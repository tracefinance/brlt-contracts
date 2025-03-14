// Package keystore provides secure key management operations for cryptographic keys.
//
// The keystore package is part of the Core/Infrastructure Layer and provides
// functionality for creating, storing, and managing cryptographic keys securely.
// It supports multiple key types (ECDSA, RSA, Ed25519, Symmetric) and different
// storage backends (DB, KMS).
//
// Key Management:
//   - Secure storage of private keys (encrypted at rest)
//   - Public key distribution
//   - Key metadata management
//   - Signing operations without exposing private keys
package keystore

import (
	"context"
	"crypto/elliptic"
	"errors"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/db"
	"vault0/internal/types"
)

// Common errors returned by keystore operations.
var (
	// ErrKeyNotFound is returned when attempting to access a non-existent key.
	ErrKeyNotFound = errors.New("key not found")

	// ErrKeyAlreadyExists is returned when attempting to create a key with a name
	// that is already in use.
	ErrKeyAlreadyExists = errors.New("key already exists")

	// ErrInvalidKey is returned when a key's format or content is invalid.
	ErrInvalidKey = errors.New("invalid key")

	// ErrEncryptionFailed is returned when key material encryption fails.
	ErrEncryptionFailed = errors.New("encryption failed")

	// ErrDecryptionFailed is returned when key material decryption fails.
	ErrDecryptionFailed = errors.New("decryption failed")

	// ErrInvalidCurve is returned when an unsupported elliptic curve is specified
	// for ECDSA key generation.
	ErrInvalidCurve = errors.New("invalid curve")
)

// KeyStoreType represents the type of key store implementation to use.
// Different implementations may provide different security guarantees
// and integration options.
type KeyStoreType string

// Supported key store types.
const (
	// KeyStoreTypeDB stores keys in an encrypted format in a SQL database.
	KeyStoreTypeDB KeyStoreType = "db"

	// KeyStoreTypeKMS uses a Key Management Service for key operations.
	// This provides hardware security module (HSM) backing when available.
	KeyStoreTypeKMS KeyStoreType = "kms"
)

// DataType indicates how input data should be processed during cryptographic operations.
type DataType string

// Supported data types for cryptographic operations.
const (
	// DataTypeRaw indicates that data has not been hashed and requires hashing
	// during cryptographic operations. This is the most common case for general
	// signing operations.
	DataTypeRaw DataType = "raw"

	// DataTypeDigest indicates that data is already hashed/digested and should
	// be used as-is in cryptographic operations. This is useful when working
	// with standardized protocols that specify their own hashing requirements.
	DataTypeDigest DataType = "digest"
)

// Key represents a cryptographic key and its associated metadata.
// Private key material is always encrypted at rest and is only accessible
// through the KeyStore interface's cryptographic operations.
type Key struct {
	// ID is the unique identifier for the key
	ID string

	// Name is a human-readable identifier for the key
	Name string

	// Type specifies the cryptographic algorithm family (e.g., ECDSA, RSA)
	Type types.KeyType

	// Curve specifies the elliptic curve parameters for ECDSA keys
	// This field is only relevant for ECDSA keys and is nil for other key types
	Curve elliptic.Curve

	// Tags store arbitrary metadata associated with the key
	// Common uses include: environment, application, owner, purpose
	Tags map[string]string

	// CreatedAt is the Unix timestamp when the key was created
	CreatedAt int64

	// PrivateKey contains the encrypted private key material
	// This field is only populated during key creation and import operations
	// and is never exposed outside the keystore package
	PrivateKey []byte

	// PublicKey contains the public key material
	// This is always available for asymmetric keys and is nil for symmetric keys
	PublicKey []byte
}

// KeyStore defines the interface for secure key management operations.
// Implementations must ensure that private key material is properly protected
// and that all operations are performed securely.
type KeyStore interface {
	// Create generates a new cryptographic key with the specified parameters.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - name: Human-readable identifier for the key
	//   - keyType: The type of key to create (ECDSA, RSA, Ed25519, Symmetric)
	//   - curve: For ECDSA keys, specifies the curve to use (e.g., P256, P384)
	//   - tags: Optional metadata to associate with the key
	//
	// Returns:
	//   - *Key: The created key (private key material is not included)
	//   - error: ErrKeyAlreadyExists if name is in use, or other creation errors
	Create(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, tags map[string]string) (*Key, error)

	// Import stores an existing key in the keystore.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - name: Human-readable identifier for the key
	//   - keyType: The type of key being imported
	//   - curve: For ECDSA keys, must match the curve used to generate the key
	//   - privateKey: The private key material to import
	//   - publicKey: The public key material (required for asymmetric keys)
	//   - tags: Optional metadata to associate with the key
	//
	// Returns:
	//   - *Key: The imported key (private key material is not included)
	//   - error: ErrKeyAlreadyExists if name is in use, or other import errors
	Import(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, privateKey, publicKey []byte, tags map[string]string) (*Key, error)

	// Sign performs a cryptographic signing operation using the specified key.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The ID of the key to use for signing
	//   - data: The data to sign
	//   - dataType: Specifies if the data needs to be hashed before signing
	//
	// Returns:
	//   - []byte: The cryptographic signature
	//   - error: ErrKeyNotFound if key doesn't exist, or other signing errors
	Sign(ctx context.Context, id string, data []byte, dataType DataType) ([]byte, error)

	// GetPublicKey retrieves the public portion of a key.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The ID of the key to retrieve
	//
	// Returns:
	//   - *Key: The key with only public information
	//   - error: ErrKeyNotFound if key doesn't exist
	GetPublicKey(ctx context.Context, id string) (*Key, error)

	// List retrieves all keys in the keystore.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//
	// Returns:
	//   - []*Key: List of keys (private key material is not included)
	//   - error: Any error that occurred during retrieval
	List(ctx context.Context) ([]*Key, error)

	// Update modifies a key's metadata.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The ID of the key to update
	//   - name: New name for the key
	//   - tags: New tags for the key
	//
	// Returns:
	//   - *Key: The updated key
	//   - error: ErrKeyNotFound if key doesn't exist
	Update(ctx context.Context, id string, name string, tags map[string]string) (*Key, error)

	// Delete removes a key from the keystore.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The ID of the key to delete
	//
	// Returns:
	//   - error: ErrKeyNotFound if key doesn't exist
	Delete(ctx context.Context, id string) error
}

// NewKeyStore creates a new KeyStore instance based on the configuration.
//
// Parameters:
//   - db: Database connection for storing keys
//   - cfg: Configuration specifying the type of keystore to create
//
// Returns:
//   - KeyStore: The configured keystore implementation
//   - error: Any error that occurred during creation
func NewKeyStore(db *db.DB, cfg *config.Config) (KeyStore, error) {
	keyStoreType := KeyStoreType(cfg.KeyStoreType)

	switch keyStoreType {
	case KeyStoreTypeDB:
		return NewDBKeyStore(db.GetConnection(), cfg)
	case KeyStoreTypeKMS:
		return nil, errors.New("KMS key store not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported key store type: %s", keyStoreType)
	}
}
