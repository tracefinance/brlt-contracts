package wallet

import (
	"database/sql"
	"time"

	"github.com/govalues/decimal"

	"vault0/internal/db"
	"vault0/internal/types"
)

// Wallet represents a wallet entity stored in the database
type Wallet struct {
	ID              int64             `db:"id"`
	KeyID           string            `db:"key_id"`
	ChainType       types.ChainType   `db:"chain_type"`
	Address         string            `db:"address"`
	Name            string            `db:"name"`
	Tags            map[string]string `db:"tags"`
	Balance         decimal.Decimal   `db:"balance"`
	LastBlockNumber int64             `db:"last_block_number"`
	CreatedAt       time.Time         `db:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"`
	DeletedAt       sql.NullTime      `db:"deleted_at"`
}

// TokenBalance represents a token balance for a wallet
type TokenBalance struct {
	WalletID  int64           `db:"wallet_id"`
	TokenID   int64           `db:"token_id"`
	Balance   decimal.Decimal `db:"balance"`
	UpdatedAt time.Time       `db:"updated_at"`
}

// TokenBalanceData contains a token with its balance
type TokenBalanceData struct {
	Token     *types.Token
	Balance   decimal.Decimal
	UpdatedAt time.Time
}

// ScanWallet scans a database row into a Wallet struct
func ScanWallet(row interface {
	Scan(dest ...any) error
}) (*Wallet, error) {
	wallet := &Wallet{}
	var tagsJSON sql.NullString
	var balanceStr sql.NullString

	err := row.Scan(
		&wallet.ID,
		&wallet.KeyID,
		&wallet.ChainType,
		&wallet.Address,
		&wallet.Name,
		&tagsJSON,
		&balanceStr,
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

	// Parse the balance as decimal
	if balanceStr.Valid {
		dec, err := decimal.Parse(balanceStr.String)
		if err != nil {
			// If there's an error parsing, use zero
			wallet.Balance = decimal.Zero
		} else {
			wallet.Balance = dec
		}
	} else {
		wallet.Balance = decimal.Zero
	}

	return wallet, nil
}

// ScanTokenBalance scans a database row into a TokenBalance struct
func ScanTokenBalance(row interface {
	Scan(dest ...any) error
}) (*TokenBalance, error) {
	tokenBalance := &TokenBalance{}
	var balanceStr sql.NullString

	err := row.Scan(
		&tokenBalance.WalletID,
		&tokenBalance.TokenID,
		&balanceStr,
		&tokenBalance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse the balance as decimal
	if balanceStr.Valid {
		dec, err := decimal.Parse(balanceStr.String)
		if err != nil {
			// If there's an error parsing, use zero
			tokenBalance.Balance = decimal.Zero
		} else {
			tokenBalance.Balance = dec
		}
	} else {
		tokenBalance.Balance = decimal.Zero
	}

	return tokenBalance, nil
}
