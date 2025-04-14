package types

import (
	"fmt"
	"math/big"
	"strings"

	"vault0/internal/errors"
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
	// TokenSymbol is the symbol of the token (e.g., "ETH", "USDC")
	TokenSymbol string
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
	Hash              string    // Transaction hash
	ChainType         ChainType // The chain this receipt belongs to
	BlockNumber       *big.Int  // Block number
	Status            uint64    // 1 for success, 0 for failure
	GasUsed           uint64    // Gas used by this transaction
	CumulativeGasUsed uint64    // Cumulative gas used in the block
	LogsBloom         []byte    // Bloom filter for logs
	Logs              []Log     // Logs emitted by the transaction
}

// Log represents a log entry from a transaction
type Log struct {
	Address         string    // Contract address that emitted the log
	ChainType       ChainType // The chain this log originated from
	Topics          []string  // Indexed log topics
	Data            []byte    // Log data
	BlockNumber     *big.Int  // Block number
	TransactionHash string    // Transaction hash
	LogIndex        uint      // Log index in the block
}

// ParseAddressFromTopic extracts and validates an address from the topic at the given index.
// Assumes the topic is a 32-byte hex string where the address is the last 20 bytes.
// Uses the Log's ChainType for validation.
// Returns an error on invalid index, topic format, or address.
func (l *Log) ParseAddressFromTopic(topicIndex int) (*Address, error) {
	if topicIndex < 0 || topicIndex >= len(l.Topics) {
		// Use specific error for index out of bounds
		return nil, errors.NewLogTopicIndexOutOfBoundsError(topicIndex, len(l.Topics))
	}

	topic := l.Topics[topicIndex]
	// Remove potential "0x" prefix for length calculation
	topicHex := strings.TrimPrefix(topic, "0x")

	// Expecting 64 hex characters for a 32-byte topic
	// We need at least the last 40 hex characters (20 bytes) for the address.
	if len(topicHex) < 40 {
		// Use specific error for invalid format (insufficient length)
		reason := fmt.Sprintf("insufficient length (%d), expected at least 40 hex characters", len(topicHex))
		return nil, errors.NewLogTopicInvalidFormatError(topicIndex, topic, reason)
	}

	// Extract the last 40 hex characters (20 bytes) and prepend "0x"
	addressHex := "0x" + topicHex[len(topicHex)-40:]

	// Use NewAddress for validation and creation, using the Log's ChainType
	addr, err := NewAddress(addressHex, l.ChainType)
	if err != nil {
		// Propagate the error from NewAddress directly (it should be a Vault0Error)
		return nil, err
	}

	return addr, nil
}
