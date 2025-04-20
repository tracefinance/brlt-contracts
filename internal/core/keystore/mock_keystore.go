package keystore

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"math/big"
	"sync"
	"time"

	"vault0/internal/core/crypto"
	"vault0/internal/core/keygen"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// MockKeyStore provides a mock implementation of the KeyStore interface for testing
type MockKeyStore struct {
	Keys   map[string]*Key
	keyGen keygen.KeyGenerator
	mutex  sync.RWMutex
}

// NewMockKeyStore creates a new instance of the MockKeyStore
func NewMockKeyStore() *MockKeyStore {
	return &MockKeyStore{
		Keys:   make(map[string]*Key),
		keyGen: keygen.NewKeyGenerator(),
	}
}

// Create creates a new key with the given ID, name, and type
func (ks *MockKeyStore) Create(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, tags map[string]string) (*Key, error) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Use a simple ID generation for mocks
	id := fmt.Sprintf("key-%d", len(ks.Keys)+1)

	// Check if key exists
	if _, exists := ks.Keys[id]; exists {
		return nil, errors.NewResourceAlreadyExistsError("key", "id", id)
	}

	// Create key
	key := &Key{
		ID:        id,
		Name:      name,
		Type:      keyType,
		Curve:     curve,
		Tags:      tags,
		CreatedAt: time.Now(),
	}

	// Generate key material
	privateKey, publicKey, err := ks.keyGen.GenerateKeyPair(keyType, curve)
	if err != nil {
		return nil, err
	}

	// Store key material (in mock we don't encrypt)
	key.PrivateKey = privateKey
	key.PublicKey = publicKey

	// Save key
	ks.Keys[id] = key

	return key, nil
}

// Import imports an existing key
func (ks *MockKeyStore) Import(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, privateKey, publicKey []byte, tags map[string]string) (*Key, error) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Use a simple ID generation for mocks
	id := fmt.Sprintf("key-%d", len(ks.Keys)+1)

	// Check if key exists
	if _, exists := ks.Keys[id]; exists {
		return nil, errors.NewResourceAlreadyExistsError("key", "id", id)
	}

	// Create key
	key := &Key{
		ID:         id,
		Name:       name,
		Type:       keyType,
		Curve:      curve,
		Tags:       tags,
		CreatedAt:  time.Now(),
		PrivateKey: privateKey, // In mock we store unencrypted
		PublicKey:  publicKey,
	}

	// Save key
	ks.Keys[id] = key

	return key, nil
}

// Sign signs the provided data using the key identified by id
func (ks *MockKeyStore) Sign(ctx context.Context, id string, data []byte, dataType DataType) ([]byte, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	// Get the key
	key, exists := ks.Keys[id]
	if !exists {
		return nil, errors.NewKeyNotFoundError(id)
	}

	// For ECDSA keys, use proper signing
	if key.Type == types.KeyTypeECDSA && len(key.PrivateKey) > 0 {
		var privateKey *ecdsa.PrivateKey
		var err error

		// Handle secp256k1 curve specially
		if key.Curve == crypto.Secp256k1Curve {
			privateKey, err = crypto.UnmarshalPrivateKey(key.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal secp256k1 private key: %w", err)
			}
		} else {
			// For other curves, use standard x509 parsing
			privateKey, err = x509.ParseECPrivateKey(key.PrivateKey)
			if err != nil {
				return nil, fmt.Errorf("failed to parse ECDSA private key: %w", err)
			}
		}

		var digest []byte
		if dataType == DataTypeRaw {
			// Create a hash of the data
			hash := sha256.Sum256(data)
			digest = hash[:]
		} else {
			// Use the data as-is (it's already hashed)
			digest = data
		}

		// Sign the hash
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest)
		if err != nil {
			return nil, err
		}

		// ASN.1 DER encoding for ECDSA signatures
		type ECDSASignature struct {
			R, S *big.Int
		}
		signature, err := asn1.Marshal(ECDSASignature{R: r, S: s})
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ECDSA signature: %w", err)
		}

		return signature, nil
	}

	// For other key types in the mock, just return some mock data
	mockSignature := make([]byte, 64)
	_, err := rand.Read(mockSignature)
	if err != nil {
		return nil, err
	}
	return mockSignature, nil
}

// GetPublicKey retrieves only the public part of a key by its ID
func (ks *MockKeyStore) GetPublicKey(ctx context.Context, id string) (*Key, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	// Get the key
	key, exists := ks.Keys[id]
	if !exists {
		return nil, errors.NewKeyNotFoundError(id)
	}

	// Return a copy of the key without the private key material
	return &Key{
		ID:        key.ID,
		Name:      key.Name,
		Type:      key.Type,
		Tags:      key.Tags,
		CreatedAt: key.CreatedAt,
		PublicKey: key.PublicKey,
		Curve:     key.Curve,
	}, nil
}

// List lists all keys
func (ks *MockKeyStore) List(ctx context.Context, limit int, nextToken string) (*types.Page[*Key], error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	// Create a list of keys without private key material
	keys := make([]*Key, 0, len(ks.Keys))
	for _, key := range ks.Keys {
		keys = append(keys, &Key{
			ID:        key.ID,
			Name:      key.Name,
			Type:      key.Type,
			Tags:      key.Tags,
			CreatedAt: key.CreatedAt,
			PublicKey: key.PublicKey,
			Curve:     key.Curve,
		})
	}

	// For tests, we ignore pagination and just return all keys
	return &types.Page[*Key]{
		Items:     keys,
		NextToken: "",
		Limit:     limit,
	}, nil
}

// Update updates a key's metadata
func (ks *MockKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*Key, error) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Get the key
	key, exists := ks.Keys[id]
	if !exists {
		return nil, errors.NewKeyNotFoundError(id)
	}

	// Update the key
	key.Name = name
	key.Tags = tags

	// Return a copy of the key without the private key material
	return &Key{
		ID:        key.ID,
		Name:      key.Name,
		Type:      key.Type,
		Tags:      key.Tags,
		CreatedAt: key.CreatedAt,
		PublicKey: key.PublicKey,
		Curve:     key.Curve,
	}, nil
}

// Delete deletes a key by its ID
func (ks *MockKeyStore) Delete(ctx context.Context, id string) error {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Check if key exists
	if _, exists := ks.Keys[id]; !exists {
		return errors.NewKeyNotFoundError(id)
	}

	// Delete the key
	delete(ks.Keys, id)
	return nil
}
