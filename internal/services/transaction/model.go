package transaction

import (
	"database/sql"
	"math/big"
	"time"

	"vault0/internal/errors"
	"vault0/internal/types"
)

// Transaction represents a transaction entity stored in the database.
type Transaction struct {
	// Service fields
	ID        int64         `db:"id"`
	WalletID  sql.NullInt64 `db:"wallet_id"`
	VaultID   sql.NullInt64 `db:"vault_id"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
	DeletedAt sql.NullTime  `db:"deleted_at"`

	// BaseTransaction
	Chain    types.ChainType       `db:"chain_type"`
	Hash     string                `db:"hash"`
	From     string                `db:"from_address"`
	To       string                `db:"to_address"`
	Value    *types.BigInt         `db:"value"`
	Data     []byte                `db:"data"`
	Nonce    uint64                `db:"nonce"`
	GasPrice *types.BigInt         `db:"gas_price"`
	GasLimit uint64                `db:"gas_limit"`
	Type     types.TransactionType `db:"type"`

	// Execution details
	ContractAddress sql.NullString          `db:"contract_address"`
	GasUsed         sql.NullInt64           `db:"gas_used"`
	Status          types.TransactionStatus `db:"status"`
	Timestamp       sql.NullInt64           `db:"timestamp"`
	BlockNumber     *types.BigInt           `db:"block_number"`
}

// ScanTransaction scans a database row into a Transaction struct.
// It handles the mapping between database types and the struct fields,
// including custom types like types.BigInt and sql.Null*.
func ScanTransaction(row interface {
	Scan(dest ...any) error
}) (*Transaction, error) {
	tx := &Transaction{}

	err := row.Scan(
		&tx.ID,
		&tx.WalletID,
		&tx.VaultID,
		&tx.CreatedAt,
		&tx.UpdatedAt,
		&tx.DeletedAt,
		&tx.Chain,
		&tx.Hash,
		&tx.From,
		&tx.To,
		&tx.Value,
		&tx.Data,
		&tx.Nonce,
		&tx.GasPrice,
		&tx.GasLimit,
		&tx.Type,
		&tx.ContractAddress,
		&tx.GasUsed,
		&tx.Status,
		&tx.Timestamp,
		&tx.BlockNumber,
	)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return tx, nil
}

// FromCoreTransaction converts a core transaction to a service transaction.
// It maps fields from types.Transaction (including embedded BaseTransaction)
// to the service layer's Transaction model.
// It attempts to extract WalletID and VaultID from the coreTx.Metadata if present.
func FromCoreTransaction(coreTx *types.Transaction) *Transaction {
	if coreTx == nil {
		return nil
	}

	// Initialize WalletID and VaultID as invalid
	var walletID sql.NullInt64
	var vaultID sql.NullInt64

	// Attempt to extract IDs from metadata if the map exists
	if coreTx.Metadata != nil {
		if idVal, ok := coreTx.Metadata.GetInt64(types.WalletIDMetadaKey); ok {
			walletID = sql.NullInt64{Int64: idVal, Valid: true}
		}
		if idVal, ok := coreTx.Metadata.GetInt64(types.VaultIDMetadaKey); ok {
			vaultID = sql.NullInt64{Int64: idVal, Valid: true}
		}
	}

	tx := &Transaction{
		WalletID: walletID,
		VaultID:  vaultID,
		Chain:    coreTx.BaseTransaction.ChainType,
		Hash:     coreTx.BaseTransaction.Hash,
		From:     coreTx.BaseTransaction.From,
		To:       coreTx.BaseTransaction.To,
		Data:     coreTx.BaseTransaction.Data,
		Nonce:    coreTx.BaseTransaction.Nonce,
		GasLimit: coreTx.BaseTransaction.GasLimit,
		Type:     coreTx.BaseTransaction.Type,
		Status:   coreTx.Status,
	}

	// Handle *big.Int to *types.BigInt conversion (checking for nil)
	if coreTx.BaseTransaction.Value != nil {
		value := types.NewBigInt(coreTx.BaseTransaction.Value)
		tx.Value = &value
	} else {
		zeroValue := types.NewBigInt(big.NewInt(0))
		tx.Value = &zeroValue
	}
	if coreTx.BaseTransaction.GasPrice != nil {
		gasPrice := types.NewBigInt(coreTx.BaseTransaction.GasPrice)
		tx.GasPrice = &gasPrice
	} else {
		zeroGasPrice := types.NewBigInt(big.NewInt(0))
		tx.GasPrice = &zeroGasPrice
	}
	if coreTx.BlockNumber != nil {
		blockNumber := types.NewBigInt(coreTx.BlockNumber)
		tx.BlockNumber = &blockNumber
	}

	// Handle nullable fields from coreTx
	if coreTx.Timestamp > 0 {
		tx.Timestamp = sql.NullInt64{Int64: coreTx.Timestamp, Valid: true}
	}
	if coreTx.GasUsed > 0 {
		tx.GasUsed = sql.NullInt64{Int64: int64(coreTx.GasUsed), Valid: true}
	}

	return tx
}

// ToCoreTransaction converts a service transaction back to a core transaction.
// It maps fields from the service layer's Transaction model to types.Transaction
// (including embedded BaseTransaction).
func (t *Transaction) ToCoreTransaction() *types.Transaction {
	if t == nil {
		return nil
	}

	coreTx := &types.Transaction{
		BaseTransaction: types.BaseTransaction{
			ChainType: t.Chain,
			Hash:      t.Hash,
			From:      t.From,
			To:        t.To,
			Data:      t.Data,
			Nonce:     t.Nonce,
			GasLimit:  t.GasLimit,
			Type:      t.Type,
		},
		Status: t.Status,
	}

	// Handle *types.BigInt to *big.Int conversion (checking for nil)
	if t.Value != nil {
		coreTx.BaseTransaction.Value = t.Value.ToBigInt()
	} else {
		coreTx.BaseTransaction.Value = big.NewInt(0)
	}
	if t.GasPrice != nil {
		coreTx.BaseTransaction.GasPrice = t.GasPrice.ToBigInt()
	} else {
		coreTx.BaseTransaction.GasPrice = big.NewInt(0)
	}
	if t.BlockNumber != nil {
		coreTx.BlockNumber = t.BlockNumber.ToBigInt()
	}

	// Handle nullable fields from service model
	if t.Timestamp.Valid {
		coreTx.Timestamp = t.Timestamp.Int64
	}
	if t.GasUsed.Valid {
		coreTx.GasUsed = uint64(t.GasUsed.Int64)
	}

	return coreTx
}

// IsContractCall checks if the transaction type is a contract call.
func (t *Transaction) IsContractCall() bool {
	return t != nil && t.Type == types.TransactionTypeContractCall
}

// IsERC20Transfer checks if the transaction type is an ERC20 transfer.
func (t *Transaction) IsERC20Transfer() bool {
	return t != nil && t.Type == types.TransactionTypeERC20Transfer
}

// Filter represents the criteria for filtering transactions in the service layer.
type Filter struct {
	Status       *types.TransactionStatus
	ChainType    *types.ChainType
	WalletID     *sql.NullInt64
	VaultID      *sql.NullInt64
	Address      *string
	Type         *types.TransactionType
	BlockNumber  *big.Int
	MinBlock     *big.Int
	MaxBlock     *big.Int
	TokenAddress *string // For filtering by contract address
}

// Helper method to check if any filter criteria are set
func (f *Filter) IsEmpty() bool {
	return f == nil || (f.Status == nil &&
		f.ChainType == nil &&
		f.WalletID == nil &&
		f.VaultID == nil &&
		f.Address == nil &&
		f.Type == nil &&
		f.BlockNumber == nil &&
		f.MinBlock == nil &&
		f.MaxBlock == nil &&
		f.TokenAddress == nil)
}
