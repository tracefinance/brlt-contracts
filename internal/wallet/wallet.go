package wallet

import (
	"context"
	"errors"
	"math/big"
)

// Common errors
var (
	ErrUnsupportedChain    = errors.New("unsupported blockchain")
	ErrInvalidAddress      = errors.New("invalid address")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// ChainType represents the blockchain type
type ChainType string

// Supported blockchain types
const (
	ChainTypeEthereum ChainType = "ethereum"
	ChainTypePolygon  ChainType = "polygon"
	ChainTypeBase     ChainType = "base"
)

// TransactionType represents the type of transaction
type TransactionType string

// Supported transaction types
const (
	TransactionTypeNative   TransactionType = "native"
	TransactionTypeERC20    TransactionType = "erc20"
	TransactionTypeContract TransactionType = "contract"
)

// Transaction represents a blockchain transaction
type Transaction struct {
	// Chain is the blockchain type
	Chain ChainType
	// Hash is the transaction hash
	Hash string
	// From is the sender address
	From string
	// To is the recipient address
	To string
	// Value is the amount of native currency
	Value *big.Int
	// Data is the transaction data (for smart contract interactions)
	Data []byte
	// Nonce is the transaction nonce
	Nonce uint64
	// GasPrice is the gas price (for EVM chains)
	GasPrice *big.Int
	// GasLimit is the gas limit (for EVM chains)
	GasLimit uint64
	// Type is the transaction type
	Type TransactionType
	// TokenAddress is the token contract address (for ERC20 transactions)
	TokenAddress string
	// Status is the transaction status
	Status string
	// Timestamp is the transaction timestamp
	Timestamp int64
}

// TransactionOptions represents optional parameters for a transaction
type TransactionOptions struct {
	// GasPrice is the gas price (for EVM chains)
	GasPrice *big.Int
	// GasLimit is the gas limit (for EVM chains)
	GasLimit uint64
	// Nonce is the transaction nonce
	Nonce uint64
	// Data is additional transaction data
	Data []byte
}

// Wallet defines the interface for wallet operations
type Wallet interface {
	// ChainType returns the blockchain type
	ChainType() ChainType

	// DeriveAddress derives a wallet address from a public key
	DeriveAddress(ctx context.Context, publicKey []byte) (string, error)

	// CreateNativeTransaction creates a native currency transaction without broadcasting
	CreateNativeTransaction(ctx context.Context, fromAddress, toAddress string, amount *big.Int, options TransactionOptions) (*Transaction, error)

	// CreateTokenTransaction creates an ERC20 token transaction without broadcasting
	CreateTokenTransaction(ctx context.Context, fromAddress, tokenAddress, toAddress string, amount *big.Int, options TransactionOptions) (*Transaction, error)

	// SignTransaction signs a transaction
	SignTransaction(ctx context.Context, keyID string, tx *Transaction) ([]byte, error)
}
