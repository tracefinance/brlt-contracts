package signer

import (
	"time"
)

// SignerType represents the type of signer
type SignerType string

const (
	// Internal indicates a signer tied to a user
	Internal SignerType = "internal"
	// External indicates a standalone signer
	External SignerType = "external"
)

// Address represents a blockchain address associated with a signer
type Address struct {
	ID        int64     `db:"id"`
	SignerID  int64     `db:"signer_id"`
	ChainType string    `db:"chain_type"`
	Address   string    `db:"address"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Signer represents a entity responsible for signing transactions
type Signer struct {
	ID        int64      `db:"id"`
	Name      string     `db:"name"`
	Type      SignerType `db:"type"`
	UserID    *int64     `db:"user_id"` // NULL for external signers
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	// Not stored in database, populated by repository when needed
	Addresses []*Address `db:"-"`
}
