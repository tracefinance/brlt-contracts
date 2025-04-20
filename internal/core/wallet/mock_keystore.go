package wallet

import (
	"context"
	"crypto/elliptic"

	"vault0/internal/core/keystore"
	"vault0/internal/types"
)

// MockKeyStore is a mock implementation of the keystore.KeyStore interface for testing
type MockKeyStore struct {
	GetPublicKeyFunc func(ctx context.Context, id string) (*keystore.Key, error)
	SignFunc         func(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error)
	CreateFunc       func(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error)
}

func (m *MockKeyStore) Create(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, tags map[string]string) (*keystore.Key, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, name, keyType, curve, tags)
	}

	// Generate a mock ID
	id := "mock-key-id"

	return &keystore.Key{
		ID:    id,
		Name:  name,
		Type:  keyType,
		Tags:  tags,
		Curve: curve,
	}, nil
}

func (m *MockKeyStore) Import(ctx context.Context, name string, keyType types.KeyType, curve elliptic.Curve, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	// Generate a mock ID
	id := "mock-key-id"

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

func (m *MockKeyStore) List(ctx context.Context, limit int, nextToken string) (*types.Page[*keystore.Key], error) {
	// For mock implementation, ignore pagination and just return an empty page
	return &types.Page[*keystore.Key]{
		Items:     []*keystore.Key{},
		NextToken: "",
		Limit:     limit,
	}, nil
}

func (m *MockKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error) {
	return nil, nil
}

func (m *MockKeyStore) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *MockKeyStore) Sign(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
	if m.SignFunc != nil {
		return m.SignFunc(ctx, id, data, dataType)
	}
	return nil, nil
}
