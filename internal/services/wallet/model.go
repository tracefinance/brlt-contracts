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
	ErrMissingKeyID   = errors.New("internal wallet requires a key ID")
	ErrMissingAddress = errors.New("external wallet requires an address")
)

// Wallet represents a wallet entity stored in the database
type Wallet struct {
	ID        string            `db:"id"`
	KeyID     string            `db:"key_id"`
	ChainType types.ChainType   `db:"chain_type"`
	Address   string            `db:"address"`
	Name      string            `db:"name"`
	Tags      map[string]string `db:"tags"`
	CreatedAt time.Time         `db:"created_at"`
	UpdatedAt time.Time         `db:"updated_at"`
	DeletedAt sql.NullTime      `db:"deleted_at"`
}

// ScanWallet scans a database row into a Wallet struct
func ScanWallet(row interface {
	Scan(dest ...any) error
}) (*Wallet, error) {
	wallet := &Wallet{}
	var tagsJSON sql.NullString

	err := row.Scan(
		&wallet.ID,
		&wallet.KeyID,
		&wallet.ChainType,
		&wallet.Address,
		&wallet.Name,
		&tagsJSON,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
		&wallet.DeletedAt,
	)
	if err != nil {
		return nil, err
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
