package wallet

import (
	"context"
	"crypto/elliptic"

	"vault0/internal/keygen"
	"vault0/internal/keystore"
)

// MockKeyStore is a mock implementation of the keystore.KeyStore interface for testing
type MockKeyStore struct {
	GetPublicKeyFunc func(ctx context.Context, id string) (*keystore.Key, error)
}

func (m *MockKeyStore) Create(ctx context.Context, id, name string, keyType keygen.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
	return &keystore.Key{
		ID:    id,
		Name:  name,
		Type:  keyType,
		Tags:  tags,
		Curve: curve,
	}, nil
}

func (m *MockKeyStore) Import(ctx context.Context, id, name string, keyType keygen.KeyType, curve elliptic.Curve, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	return &keystore.Key{
		ID:        id,
		Name:      name,
		Type:      keyType,
		Tags:      tags,
		PublicKey: publicKey,
		Curve:     curve,
	}, nil
}

func (m *MockKeyStore) GetPublicKey(ctx context.Context, id string) (*keystore.Key, error) {
	if m.GetPublicKeyFunc != nil {
		return m.GetPublicKeyFunc(ctx, id)
	}
	return &keystore.Key{ID: id}, nil
}

func (m *MockKeyStore) List(ctx context.Context) ([]*keystore.Key, error) {
	return nil, nil
}

func (m *MockKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error) {
	return nil, nil
}

func (m *MockKeyStore) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockKeyStore) Sign(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
	return nil, nil
}
