package keystore

import (
	"errors"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/db"
)

type Factory interface {
	NewKeyStore() (KeyStore, error)
}

// KeyStoreType represents the type of key store to use
type KeyStoreType string

// Supported key store types
const (
	KeyStoreTypeDB  KeyStoreType = "db"
	KeyStoreTypeKMS KeyStoreType = "kms"
	// Add more key store types here as they are implemented
)

// Factory creates KeyStore instances based on the specified type
type factory struct {
	cfg *config.Config
	db  *db.DB
}

// NewFactory creates a new KeyStore factory
func NewFactory(db *db.DB, cfg *config.Config) Factory {
	return &factory{
		cfg: cfg,
		db:  db,
	}
}

// NewKeyStore creates a new KeyStore instance based on the type specified in config
func (f *factory) NewKeyStore() (KeyStore, error) {
	keyStoreType := KeyStoreType(f.cfg.KeyStoreType)

	switch keyStoreType {
	case KeyStoreTypeDB:
		return NewDBKeyStore(f.db.GetConnection(), f.cfg)
	case KeyStoreTypeKMS:
		return nil, errors.New("KMS key store not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported key store type: %s", keyStoreType)
	}
}
