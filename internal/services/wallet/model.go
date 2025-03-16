package wallet

import (
	"database/sql"
	"time"

	"vault0/internal/core/db"
	"vault0/internal/types"
)

// Wallet represents a wallet entity stored in the database
type Wallet struct {
	ID              string            `db:"id"`
	KeyID           string            `db:"key_id"`
	ChainType       types.ChainType   `db:"chain_type"`
	Address         string            `db:"address"`
	Name            string            `db:"name"`
	Tags            map[string]string `db:"tags"`
	LastBlockNumber int64             `db:"last_block_number"`
	CreatedAt       time.Time         `db:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"`
	DeletedAt       sql.NullTime      `db:"deleted_at"`
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
		&wallet.LastBlockNumber,
		&wallet.CreatedAt,
		&wallet.UpdatedAt,
		&wallet.DeletedAt,
	)
	if err != nil {
		return nil, err
	}

	// Use the utility function to unmarshal tags
	wallet.Tags = db.UnmarshalJSONToMap(tagsJSON)

	return wallet, nil
}
