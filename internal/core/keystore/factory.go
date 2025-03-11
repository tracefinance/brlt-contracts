package keystore

import (
	"database/sql"
	"errors"
	"fmt"

	"vault0/internal/config"
)

type KeyStoreFactory interface {
	NewKeyStore(keyStoreType KeyStoreType) (KeyStore, error)
	NewDefaultKeyStore() (KeyStore, error)
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
	db  *sql.DB
}

// NewFactory creates a new KeyStore factory
func NewFactory(cfg *config.Config, db *sql.DB) KeyStoreFactory {
	return &factory{
		cfg: cfg,
		db:  db,
	}
}

// NewKeyStore creates a new KeyStore instance of the specified type
func (f *factory) NewKeyStore(keyStoreType KeyStoreType) (KeyStore, error) {
	switch keyStoreType {
	case KeyStoreTypeDB:
		return NewDBKeyStore(f.db, f.cfg)
	case KeyStoreTypeKMS:
		return nil, errors.New("KMS key store not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported key store type: %s", keyStoreType)
	}
}

// NewDefaultKeyStore creates a new KeyStore instance with the default type (DB)
func (f *factory) NewDefaultKeyStore() (KeyStore, error) {
	return f.NewKeyStore(KeyStoreTypeDB)
}
