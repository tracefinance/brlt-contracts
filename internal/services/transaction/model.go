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
	Value        types.BigInt    `db:"value"`
	Data         []byte          `db:"data"`
	Nonce        uint64          `db:"nonce"`
	GasPrice     types.BigInt    `db:"gas_price"`
	GasLimit     uint64          `db:"gas_limit"`
	Type         string          `db:"type"`
	TokenAddress string          `db:"token_address"`
	TokenSymbol  string          `db:"token_symbol"`
	Status       string          `db:"status"`
	Timestamp    int64           `db:"timestamp"`
	BlockNumber  *int64          `db:"block_number"`
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
		&tx.TokenSymbol,
		&tx.Status,
		&tx.Timestamp,
		&tx.BlockNumber,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Parse value from string
	if valueStr != "" {
		value, err := types.NewBigIntFromString(valueStr)
		if err != nil {
			return nil, err
		}
		tx.Value = value
	}

	// Parse gas price from string
	if gasPriceStr != "" {
		gasPrice, err := types.NewBigIntFromString(gasPriceStr)
		if err != nil {
			return nil, err
		}
		tx.GasPrice = gasPrice
	}

	return tx, nil
}

// FromCoreTransaction converts a core transaction to a service transaction
func FromCoreTransaction(coreTx *types.Transaction, walletID int64) *Transaction {
	var blockNumber *int64
	if coreTx.BlockNumber != nil {
		bn := coreTx.BlockNumber.Int64()
		blockNumber = &bn
	}

	return &Transaction{
		WalletID:     walletID,
		ChainType:    coreTx.Chain,
		Hash:         coreTx.Hash,
		FromAddress:  coreTx.From,
		ToAddress:    coreTx.To,
		Value:        types.NewBigInt(coreTx.Value),
		Data:         coreTx.Data,
		Nonce:        coreTx.Nonce,
		GasPrice:     types.NewBigInt(coreTx.GasPrice),
		GasLimit:     coreTx.GasLimit,
		Type:         string(coreTx.Type),
		TokenAddress: coreTx.TokenAddress,
		TokenSymbol:  coreTx.TokenSymbol,
		Status:       string(coreTx.Status),
		Timestamp:    coreTx.Timestamp,
		BlockNumber:  blockNumber,
	}
}

// ToCoreTransaction converts a service transaction to a core transaction
func (t *Transaction) ToCoreTransaction() *types.Transaction {
	var blockNumber *big.Int
	if t.BlockNumber != nil {
		blockNumber = big.NewInt(*t.BlockNumber)
	}

	return &types.Transaction{
		Chain:        t.ChainType,
		Hash:         t.Hash,
		From:         t.FromAddress,
		To:           t.ToAddress,
		Value:        t.Value.ToBigInt(),
		Data:         t.Data,
		Nonce:        t.Nonce,
		GasPrice:     t.GasPrice.ToBigInt(),
		GasLimit:     t.GasLimit,
		Type:         types.TransactionType(t.Type),
		TokenAddress: t.TokenAddress,
		TokenSymbol:  t.TokenSymbol,
		Status:       types.TransactionStatus(t.Status),
		Timestamp:    t.Timestamp,
		BlockNumber:  blockNumber,
	}
}

// Filter represents the criteria for filtering transactions
type Filter struct {
	Status       *string
	ChainType    *types.ChainType
	WalletID     *int64
	Address      *string
	TokenAddress *string
	Limit        int
	Offset       int
}

// NewFilter creates a new transaction filter with default pagination settings
func NewFilter() *Filter {
	return &Filter{
		Limit:  10, // Default limit
		Offset: 0,
	}
}

// WithStatus sets the status filter
func (f *Filter) WithStatus(status string) *Filter {
	f.Status = &status
	return f
}

// WithChainType sets the chain type filter
func (f *Filter) WithChainType(chainType types.ChainType) *Filter {
	f.ChainType = &chainType
	return f
}

// WithWalletID sets the wallet ID filter
func (f *Filter) WithWalletID(walletID int64) *Filter {
	f.WalletID = &walletID
	return f
}

// WithAddress sets the address filter (can be from or to address)
func (f *Filter) WithAddress(address string) *Filter {
	f.Address = &address
	return f
}

// WithTokenAddress sets the token address filter
func (f *Filter) WithTokenAddress(tokenAddress string) *Filter {
	f.TokenAddress = &tokenAddress
	return f
}

// WithPagination sets the pagination parameters
func (f *Filter) WithPagination(limit, offset int) *Filter {
	f.Limit = limit
	f.Offset = offset
	return f
}
