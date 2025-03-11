package wallet

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"vault0/internal/types"
)

// Common errors
var (
	ErrInvalidWalletType   = errors.New("invalid wallet type")
	ErrInvalidWalletSource = errors.New("invalid wallet source")
	ErrMissingUserID       = errors.New("user wallet requires a user ID")
	ErrMissingKeyID        = errors.New("internal wallet requires a key ID")
	ErrMissingAddress      = errors.New("external wallet requires an address")
)

// WalletType represents the type of wallet (system or user)
type WalletType string

const (
	// WalletTypeSystem represents a system wallet (one per chain type)
	WalletTypeSystem WalletType = "system"
	// WalletTypeUser represents a user wallet
	WalletTypeUser WalletType = "user"
)

// WalletSource represents the source of the wallet (internal or external)
type WalletSource string

const (
	// WalletSourceInternal represents an internal wallet with a keyID
	WalletSourceInternal WalletSource = "internal"
	// WalletSourceExternal represents an external wallet with only an address
	WalletSourceExternal WalletSource = "external"
)

// Wallet represents a wallet entity stored in the database
type Wallet struct {
	ID        string            `db:"id"`
	KeyID     string            `db:"key_id"`
	UserID    string            `db:"user_id"`
	ChainType types.ChainType   `db:"chain_type"`
	Address   string            `db:"address"`
	Name      string            `db:"name"`
	Tags      map[string]string `db:"tags"`
	Type      WalletType        `db:"type"`
	Source    WalletSource      `db:"source"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
}

// IsSystem returns true if the wallet is a system wallet
func (w *Wallet) IsSystem() bool {
	return w.Type == WalletTypeSystem
}

// IsUser returns true if the wallet is a user wallet
func (w *Wallet) IsUser() bool {
	return w.Type == WalletTypeUser
}

// IsInternal returns true if the wallet is an internal wallet
func (w *Wallet) IsInternal() bool {
	return w.Source == WalletSourceInternal
}

// IsExternal returns true if the wallet is an external wallet
func (w *Wallet) IsExternal() bool {
	return w.Source == WalletSourceExternal
}

// Validate validates the wallet based on its type and source
func (w *Wallet) Validate() error {
	// Validate wallet type
	if w.Type != WalletTypeSystem && w.Type != WalletTypeUser {
		return ErrInvalidWalletType
	}

	// Validate wallet source
	if w.Source != WalletSourceInternal && w.Source != WalletSourceExternal {
		return ErrInvalidWalletSource
	}

	// User wallets must have a user ID
	if w.Type == WalletTypeUser && w.UserID == "" {
		return ErrMissingUserID
	}

	// Internal wallets must have a key ID
	if w.Source == WalletSourceInternal && w.KeyID == "" {
		return ErrMissingKeyID
	}

	// All wallets must have an address
	if w.Address == "" {
		return ErrMissingAddress
	}

	return nil
}

// ScanWallet scans a database row into a Wallet struct
func ScanWallet(row interface {
	Scan(dest ...any) error
}) (*Wallet, error) {
	wallet := &Wallet{}
	var tagsJSON sql.NullString
	var userID sql.NullString
	var walletType, walletSource sql.NullString

	err := row.Scan(
		&wallet.ID,
		&wallet.KeyID,
		&userID,
		&wallet.ChainType,
		&wallet.Address,
		&wallet.Name,
		&tagsJSON,
		&walletType,
		&walletSource,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Set UserID if present
	if userID.Valid {
		wallet.UserID = userID.String
	}

	// Set Type if present, default to user
	if walletType.Valid && walletType.String != "" {
		wallet.Type = WalletType(walletType.String)
	} else {
		wallet.Type = WalletTypeUser
	}

	// Set Source if present, determine based on KeyID
	if walletSource.Valid && walletSource.String != "" {
		wallet.Source = WalletSource(walletSource.String)
	} else if wallet.KeyID != "" {
		wallet.Source = WalletSourceInternal
	} else {
		wallet.Source = WalletSourceExternal
	}

	// Parse tags JSON if present
	if tagsJSON.Valid && tagsJSON.String != "" {
		err = json.Unmarshal([]byte(tagsJSON.String), &wallet.Tags)
		if err != nil {
			// If we can't parse the tags, initialize an empty map
			wallet.Tags = make(map[string]string)
		}
	} else {
		wallet.Tags = make(map[string]string)
	}

	return wallet, nil
}
