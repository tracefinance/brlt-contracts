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
	ChainTypeTron     ChainType = "tron"
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
	// Nonce is the transaction nonce (optional, will be fetched if not provided)
	Nonce *uint64
	// Data is additional transaction data
	Data []byte
}

// Wallet defines the interface for wallet operations
type Wallet interface {
	// ChainType returns the blockchain type
	ChainType() ChainType

	// DeriveAddress derives a wallet address from a public key
	DeriveAddress(ctx context.Context, publicKey []byte) (string, error)

	// GetBalance returns the native currency balance of an address
	GetBalance(ctx context.Context, address string) (*big.Int, error)

	// GetTokenBalance returns the token balance of an address
	GetTokenBalance(ctx context.Context, address, tokenAddress string) (*big.Int, error)

	// SendNative sends native currency
	SendNative(ctx context.Context, keyID, toAddress string, amount *big.Int, options *TransactionOptions) (*Transaction, error)

	// SendToken sends ERC20 tokens
	SendToken(ctx context.Context, keyID, tokenAddress, toAddress string, amount *big.Int, options *TransactionOptions) (*Transaction, error)

	// SignTransaction signs a transaction without broadcasting
	SignTransaction(ctx context.Context, keyID string, tx *Transaction) ([]byte, error)

	// BroadcastTransaction broadcasts a signed transaction
	BroadcastTransaction(ctx context.Context, signedTx []byte) (*Transaction, error)

	// GetTransaction retrieves a transaction by hash
	GetTransaction(ctx context.Context, hash string) (*Transaction, error)
}
