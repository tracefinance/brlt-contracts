package wallet

import (
	"database/sql"
	"encoding/json"
	"time"

	"vault0/internal/types"
)

// Wallet represents a wallet entity stored in the database
type Wallet struct {
	ID        string
	KeyID     string
	ChainType types.ChainType
	Address   string
	Name      string
	Tags      map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt sql.NullTime
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
