package types

import (
	"math/big"
)

// TransactionType represents the type of transaction
type TransactionType string

// Supported transaction types
const (
	TransactionTypeNative   TransactionType = "native"
	TransactionTypeERC20    TransactionType = "erc20"
	TransactionTypeContract TransactionType = "contract"
)

// TransactionStatus represents the status of a transaction in the blockchain
type TransactionStatus string

const (
	// TransactionStatusPending indicates a transaction is waiting to be mined
	TransactionStatusPending TransactionStatus = "pending"

	// TransactionStatusMined indicates a transaction is included in a block
	// but execution status is not yet known
	TransactionStatusMined TransactionStatus = "mined"

	// TransactionStatusSuccess indicates a transaction was successfully executed
	TransactionStatusSuccess TransactionStatus = "success"

	// TransactionStatusFailed indicates a transaction execution failed
	TransactionStatusFailed TransactionStatus = "failed"

	// TransactionStatusDropped indicates a transaction was dropped from mempool
	TransactionStatusDropped TransactionStatus = "dropped"

	// TransactionStatusUnknown indicates a transaction with unknown status
	TransactionStatusUnknown TransactionStatus = "unknown"
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
	// GasUsed is the actual amount of gas used (only available after mining)
	GasUsed uint64
	// Type is the transaction type
	Type TransactionType
	// TokenAddress is the token contract address (for ERC20 transactions)
	TokenAddress string
	// Status is the transaction status
	Status TransactionStatus
	// Timestamp is the transaction timestamp
	Timestamp int64
	// BlockNumber is the block number
	BlockNumber *big.Int
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
