package types

import (
	"errors"
	"math/big"
)

// Common errors
var (
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// Common addresses
const (
	// ZeroAddress represents the Ethereum zero address (0x0)
	ZeroAddress = "0x0000000000000000000000000000000000000000"
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

// TransactionReceipt contains information about a transaction's execution
type TransactionReceipt struct {
	Hash              string   // Transaction hash
	BlockNumber       *big.Int // Block number
	Status            uint64   // 1 for success, 0 for failure
	GasUsed           uint64   // Gas used by this transaction
	CumulativeGasUsed uint64   // Cumulative gas used in the block
	LogsBloom         []byte   // Bloom filter for logs
	Logs              []Log    // Logs emitted by the transaction
}

// Log represents a log entry from a transaction
type Log struct {
	Address         string   // Contract address that emitted the log
	Topics          []string // Indexed log topics
	Data            []byte   // Log data
	BlockNumber     *big.Int // Block number
	TransactionHash string   // Transaction hash
	LogIndex        uint     // Log index in the block
}
