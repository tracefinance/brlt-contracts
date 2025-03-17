package transaction

import (
	"math/big"
	"time"

	"vault0/internal/types"
)

// Transaction represents a transaction entity stored in the database
type Transaction struct {
	ID           int64           `db:"id"`
	WalletID     int64           `db:"wallet_id"`
	ChainType    types.ChainType `db:"chain_type"`
	Hash         string          `db:"hash"`
	FromAddress  string          `db:"from_address"`
	ToAddress    string          `db:"to_address"`
	Value        *big.Int        `db:"value"`
	Data         []byte          `db:"data"`
	Nonce        uint64          `db:"nonce"`
	GasPrice     *big.Int        `db:"gas_price"`
	GasLimit     uint64          `db:"gas_limit"`
	Type         string          `db:"type"`
	TokenAddress string          `db:"token_address"`
	Status       string          `db:"status"`
	Timestamp    int64           `db:"timestamp"`
	CreatedAt    time.Time       `db:"created_at"`
	UpdatedAt    time.Time       `db:"updated_at"`
}

// ScanTransaction scans a database row into a Transaction struct
func ScanTransaction(row interface {
	Scan(dest ...any) error
}) (*Transaction, error) {
	tx := &Transaction{}
	var valueStr, gasPriceStr string

	err := row.Scan(
		&tx.ID,
		&tx.WalletID,
		&tx.ChainType,
		&tx.Hash,
		&tx.FromAddress,
		&tx.ToAddress,
		&valueStr,
		&tx.Data,
		&tx.Nonce,
		&gasPriceStr,
		&tx.GasLimit,
		&tx.Type,
		&tx.TokenAddress,
		&tx.Status,
		&tx.Timestamp,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse value from string
	tx.Value = new(big.Int)
	if valueStr != "" {
		tx.Value.SetString(valueStr, 10)
	}

	// Parse gas price from string
	tx.GasPrice = new(big.Int)
	if gasPriceStr != "" {
		tx.GasPrice.SetString(gasPriceStr, 10)
	}

	return tx, nil
}

// FromCoreTransaction converts a core transaction to a service transaction
func FromCoreTransaction(coreTx *types.Transaction, walletID int64) *Transaction {
	return &Transaction{
		WalletID:     walletID,
		ChainType:    coreTx.Chain,
		Hash:         coreTx.Hash,
		FromAddress:  coreTx.From,
		ToAddress:    coreTx.To,
		Value:        coreTx.Value,
		Data:         coreTx.Data,
		Nonce:        coreTx.Nonce,
		GasPrice:     coreTx.GasPrice,
		GasLimit:     coreTx.GasLimit,
		Type:         string(coreTx.Type),
		TokenAddress: coreTx.TokenAddress,
		Status:       coreTx.Status,
		Timestamp:    coreTx.Timestamp,
	}
}

// ToCoreTransaction converts a service transaction to a core transaction
func (t *Transaction) ToCoreTransaction() *types.Transaction {
	return &types.Transaction{
		Chain:        t.ChainType,
		Hash:         t.Hash,
		From:         t.FromAddress,
		To:           t.ToAddress,
		Value:        t.Value,
		Data:         t.Data,
		Nonce:        t.Nonce,
		GasPrice:     t.GasPrice,
		GasLimit:     t.GasLimit,
		Type:         types.TransactionType(t.Type),
		TokenAddress: t.TokenAddress,
		Status:       t.Status,
		Timestamp:    t.Timestamp,
	}
}
