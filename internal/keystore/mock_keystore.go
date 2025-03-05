package keystore

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sync"
	"time"

	"vault0/internal/keygen"
)

// MockKeyStore provides a mock implementation of the KeyStore interface for testing
type MockKeyStore struct {
	keys   map[string]*Key
	keyGen keygen.KeyGenerator
	mutex  sync.RWMutex
}

// NewMockKeyStore creates a new instance of the MockKeyStore
func NewMockKeyStore() *MockKeyStore {
	return &MockKeyStore{
		keys:   make(map[string]*Key),
		keyGen: keygen.NewKeyGenerator(),
	}
}

// Create creates a new key with the given ID, name, and type
func (ks *MockKeyStore) Create(ctx context.Context, id, name string, keyType keygen.KeyType, tags map[string]string) (*Key, error) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Check if key with same ID already exists
	if _, exists := ks.keys[id]; exists {
		return nil, ErrKeyAlreadyExists
	}

	// Generate key pair
	privateKey, publicKey, err := ks.keyGen.GenerateKeyPair(keyType)
	if err != nil {
		return nil, err
	}

	// Create the key
	key := &Key{
		ID:         id,
		Name:       name,
		Type:       keyType,
		Tags:       tags,
		CreatedAt:  time.Now().Unix(),
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}

	// Store the key
	ks.keys[id] = key

	// Return a copy of the key without the private key material
	return &Key{
		ID:        key.ID,
		Name:      key.Name,
		Type:      key.Type,
		Tags:      key.Tags,
		CreatedAt: key.CreatedAt,
		PublicKey: key.PublicKey,
	}, nil
}

// Import imports an existing key
func (ks *MockKeyStore) Import(ctx context.Context, id, name string, keyType keygen.KeyType, privateKey, publicKey []byte, tags map[string]string) (*Key, error) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Check if key with same ID already exists
	if _, exists := ks.keys[id]; exists {
		return nil, ErrKeyAlreadyExists
	}

	// Create the key
	key := &Key{
		ID:         id,
		Name:       name,
		Type:       keyType,
		Tags:       tags,
		CreatedAt:  time.Now().Unix(),
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}

	// Store the key
	ks.keys[id] = key

	// Return a copy of the key without the private key material
	return &Key{
		ID:        key.ID,
		Name:      key.Name,
		Type:      key.Type,
		Tags:      key.Tags,
		CreatedAt: key.CreatedAt,
		PublicKey: key.PublicKey,
	}, nil
}

// Sign signs the provided data using the key identified by id
func (ks *MockKeyStore) Sign(ctx context.Context, id string, data []byte) ([]byte, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	// Get the key
	key, exists := ks.keys[id]
	if !exists {
		return nil, ErrKeyNotFound
	}

	// For mock implementation, we'll just do a simple ECDSA signing for all key types
	// In a real implementation, you'd use the appropriate signing method based on key type
	if key.Type == keygen.KeyTypeECDSA && len(key.PrivateKey) > 0 {
		// Parse the PEM-encoded private key
		block, _ := pem.Decode(key.PrivateKey)
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block containing private key")
		}

		// Parse the private key
		privateKey, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}

		// Hash the data
		hash := sha256.Sum256(data)

		// Sign the hash
		r, s, err := ecdsa.Sign(rand.Reader, privateKey, hash[:])
		if err != nil {
			return nil, err
		}

		// Concatenate r and s
		signature := append(r.Bytes(), s.Bytes()...)
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
	key, exists := ks.keys[id]
	if !exists {
		return nil, ErrKeyNotFound
	}

	// Return a copy of the key without the private key material
	return &Key{
		ID:        key.ID,
		Name:      key.Name,
		Type:      key.Type,
		Tags:      key.Tags,
		CreatedAt: key.CreatedAt,
		PublicKey: key.PublicKey,
	}, nil
}

// List lists all keys
func (ks *MockKeyStore) List(ctx context.Context) ([]*Key, error) {
	ks.mutex.RLock()
	defer ks.mutex.RUnlock()

	// Create a list of keys without private key material
	keys := make([]*Key, 0, len(ks.keys))
	for _, key := range ks.keys {
		keys = append(keys, &Key{
			ID:        key.ID,
			Name:      key.Name,
			Type:      key.Type,
			Tags:      key.Tags,
			CreatedAt: key.CreatedAt,
			PublicKey: key.PublicKey,
		})
	}

	return keys, nil
}

// Update updates a key's metadata
func (ks *MockKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*Key, error) {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Get the key
	key, exists := ks.keys[id]
	if !exists {
		return nil, ErrKeyNotFound
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
	}, nil
}

// Delete deletes a key by its ID
func (ks *MockKeyStore) Delete(ctx context.Context, id string) error {
	ks.mutex.Lock()
	defer ks.mutex.Unlock()

	// Check if key exists
	if _, exists := ks.keys[id]; !exists {
		return ErrKeyNotFound
	}

	// Delete the key
	delete(ks.keys, id)
	return nil
}

// Close releases any resources used by the key store
func (ks *MockKeyStore) Close() error {
	// No resources to release in the mock implementation
	return nil
}
