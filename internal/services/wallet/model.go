package wallet

import (
	"database/sql"
	"math/big"
	"time"

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
	Balance         *big.Int          `db:"balance"`
	LastBlockNumber int64             `db:"last_block_number"`
	CreatedAt       time.Time         `db:"created_at"`
	UpdatedAt       time.Time         `db:"updated_at"`
	DeletedAt       sql.NullTime      `db:"deleted_at"`
}

// TokenBalance represents a token balance for a wallet
type TokenBalance struct {
	WalletID     int64     `db:"wallet_id"`
	TokenAddress string    `db:"token_address"`
	Balance      *big.Int  `db:"balance"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// TokenBalanceData contains a token with its balance
type TokenBalanceData struct {
	Token     *types.Token
	Balance   *big.Int
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

	// Parse the balance as big.Int
	if balanceStr.Valid {
		balance, success := new(big.Int).SetString(balanceStr.String, 10)
		if !success {
			// If there's an error parsing, use zero
			wallet.Balance = new(big.Int).SetInt64(0)
		} else {
			wallet.Balance = balance
		}
	} else {
		wallet.Balance = new(big.Int).SetInt64(0)
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
		&tokenBalance.TokenAddress,
		&balanceStr,
		&tokenBalance.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse the balance as big.Int
	if balanceStr.Valid {
		balance, success := new(big.Int).SetString(balanceStr.String, 10)
		if !success {
			// If there's an error parsing, use zero
			tokenBalance.Balance = new(big.Int).SetInt64(0)
		} else {
			tokenBalance.Balance = balance
		}
	} else {
		tokenBalance.Balance = new(big.Int).SetInt64(0)
	}

	return tokenBalance, nil
}

// GetToken returns the native token for the wallet's blockchain
func (w *Wallet) GetToken() (*types.Token, error) {
	return types.NewNativeToken(w.ChainType)
}
