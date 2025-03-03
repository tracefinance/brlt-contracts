package keymanagement

import (
	"database/sql"
	"errors"
	"fmt"

	"vault0/internal/config"
)

// KeyManagerType represents the type of key manager to use
type KeyManagerType string

// Supported key manager types
const (
	KeyManagerTypeDB  KeyManagerType = "db"
	KeyManagerTypeKMS KeyManagerType = "kms"
	// Add more key manager types here as they are implemented
)

// Factory creates KeyManager instances based on the specified type
type Factory struct {
	cfg *config.Config
	db  *sql.DB
}

// NewFactory creates a new KeyManager factory
func NewFactory(cfg *config.Config, db *sql.DB) *Factory {
	return &Factory{
		cfg: cfg,
		db:  db,
	}
}

// Create creates a new KeyManager instance of the specified type
func (f *Factory) Create(keyManagerType KeyManagerType) (KeyManager, error) {
	switch keyManagerType {
	case KeyManagerTypeDB:
		return NewDBKeyManager(f.db, f.cfg)
	case KeyManagerTypeKMS:
		return nil, errors.New("KMS key manager not implemented yet")
	default:
		return nil, fmt.Errorf("unsupported key manager type: %s", keyManagerType)
	}
}

// CreateDefault creates a new KeyManager instance with the default type (DB)
func (f *Factory) CreateDefault() (KeyManager, error) {
	return f.Create(KeyManagerTypeDB)
}
