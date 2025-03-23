package wallet

import (
	"database/sql"
	"math/big"
	"time"

	"vault0/internal/types"
)

// Wallet represents a wallet entity stored in the database
type Wallet struct {
	ID              int64           `db:"id"`
	KeyID           string          `db:"key_id"`
	ChainType       types.ChainType `db:"chain_type"`
	Address         string          `db:"address"`
	Name            string          `db:"name"`
	Tags            types.JSONMap   `db:"tags"`
	Balance         types.BigInt    `db:"balance"`
	LastBlockNumber int64           `db:"last_block_number"`
	CreatedAt       time.Time       `db:"created_at"`
	UpdatedAt       time.Time       `db:"updated_at"`
	DeletedAt       sql.NullTime    `db:"deleted_at"`
}

// TokenBalance represents a token balance for a wallet
type TokenBalance struct {
	WalletID     int64        `db:"wallet_id"`
	TokenAddress string       `db:"token_address"`
	Balance      types.BigInt `db:"balance"`
	UpdatedAt    time.Time    `db:"updated_at"`
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

	// Use the JSONMap to handle tags
	if tagsJSON.Valid {
		err = wallet.Tags.Scan(tagsJSON.String)
		if err != nil {
			return nil, err
		}
	}

	// Parse the balance as BigInt
	if balanceStr.Valid {
		balance, err := types.NewBigIntFromString(balanceStr.String)
		if err != nil {
			// If there's an error parsing, use zero
			wallet.Balance = types.NewBigInt(big.NewInt(0))
		} else {
			wallet.Balance = balance
		}
	} else {
		wallet.Balance = types.NewBigInt(big.NewInt(0))
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

	// Parse the balance as BigInt
	if balanceStr.Valid {
		balance, err := types.NewBigIntFromString(balanceStr.String)
		if err != nil {
			// If there's an error parsing, use zero
			tokenBalance.Balance = types.NewBigInt(big.NewInt(0))
		} else {
			tokenBalance.Balance = balance
		}
	} else {
		tokenBalance.Balance = types.NewBigInt(big.NewInt(0))
	}

	return tokenBalance, nil
}

// GetToken returns the native token for the wallet's blockchain
func (w *Wallet) GetToken() (*types.Token, error) {
	return types.NewNativeToken(w.ChainType)
}

// ToBigInt converts the wallet's Balance to a standard *big.Int
// This is for backward compatibility with existing code
func (w *Wallet) ToBigInt() *big.Int {
	return w.Balance.ToBigInt()
}

// GetTagsMap returns the wallet's Tags as a map[string]string
// This is for backward compatibility with existing code
func (w *Wallet) GetTagsMap() map[string]string {
	return w.Tags.Map()
}
