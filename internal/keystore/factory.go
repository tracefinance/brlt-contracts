package keystore

import (
	"database/sql"
	"errors"
	"fmt"

	"vault0/internal/config"
)

// KeyStoreType represents the type of key store to use
type KeyStoreType string

// Supported key store types
const (
	KeyStoreTypeDB  KeyStoreType = "db"
	KeyStoreTypeKMS KeyStoreType = "kms"
	// Add more key store types here as they are implemented
)

// Factory creates KeyStore instances based on the specified type
type Factory struct {
	cfg *config.Config
	db  *sql.DB
}

// NewFactory creates a new KeyStore factory
func NewFactory(cfg *config.Config, db *sql.DB) *Factory {
	return &Factory{
		cfg: cfg,
		db:  db,
	}
}

// Create creates a new KeyStore instance of the specified type
func (f *Factory) Create(keyStoreType KeyStoreType) (KeyStore, error) {
	switch keyStoreType {
	case KeyStoreTypeDB:
		return NewDBKeyStore(f.db, f.cfg)
	case KeyStoreTypeKMS:
		return nil, errors.New("KMS key store not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported key store type: %s", keyStoreType)
	}
}

// CreateDefault creates a new KeyStore instance with the default type (DB)
func (f *Factory) CreateDefault() (KeyStore, error) {
	return f.Create(KeyStoreTypeDB)
}
