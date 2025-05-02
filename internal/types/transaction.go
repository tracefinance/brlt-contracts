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
	TransactionTypeNative TransactionType = "native"
	TransactionTypeDeploy TransactionType = "deploy"

	// TransactionTypeContractCall indicates a transaction is a contract method call
	TransactionTypeContractCall TransactionType = "contract_call"
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

// CoreTransaction defines an interface for accessing the core properties of a blockchain transaction
type CoreTransaction interface {
	// GetChainType returns the blockchain type
	GetChainType() ChainType
	// GetHash returns the transaction hash
	GetHash() string
	// GetFrom returns the sender address
	GetFrom() string
	// GetTo returns the recipient address
	GetTo() string
	// GetValue returns the amount of native currency transferred
	GetValue() *big.Int
	// GetData returns the transaction input data
	GetData() []byte
	// GetNonce returns the transaction nonce
	GetNonce() uint64
	// GetGasPrice returns the price per unit of gas
	GetGasPrice() *big.Int
	// GetGasLimit returns the maximum gas units the transaction can consume
	GetGasLimit() uint64
	// GetType returns the general nature of the transaction
	GetType() TransactionType
	// GetTransaction returns the transaction
	GetTransaction() *Transaction
}

// BaseTransaction holds the core, immutable fields of a blockchain transaction.
// These fields are typically known before the transaction is executed.
type BaseTransaction struct {
	// Type indicates the general nature of the transaction (e.g., native transfer, deployment)
	Type TransactionType
	// ChainType is the blockchain type
	ChainType ChainType
	// Hash is the transaction hash
	Hash string
	// From is the sender address
	From string
	// To is the recipient address (can be nil for contract creation)
	To string
	// Value is the amount of native currency transferred
	Value *big.Int
	// Data is the transaction input data (for contract interactions or creation)
	Data []byte
	// Nonce is the transaction nonce provided by the sender
	Nonce uint64
	// GasPrice is the price per unit of gas (e.g., gwei for EVM)
	GasPrice *big.Int
	// GasLimit is the maximum gas units the transaction can consume
	GasLimit uint64
}

// Transaction represents a blockchain transaction including its execution outcome.
// It embeds BaseTransaction for the core details.
type Transaction struct {
	// Embeds the core, immutable transaction details
	BaseTransaction
	// GasUsed is the actual amount of gas consumed during execution
	GasUsed uint64
	// Status indicates the execution outcome (e.g., success, failed)
	Status TransactionStatus
	// Timestamp is the Unix timestamp when the transaction was included in a block
	Timestamp int64
	// BlockNumber is the number of the block containing the transaction
	BlockNumber *big.Int
	// Metadata is a flexible field for transformers to add contextual information.
	Metadata map[string]any
}

// TransactionOptions represents optional parameters for constructing a transaction
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
	ContractAddress   *string   // Address of deployed contract (nil if not contract creation)
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

// Methods to implement CoreTransaction interface for BaseTransaction

// GetChainType returns the blockchain type
func (tx *BaseTransaction) GetChainType() ChainType {
	return tx.ChainType
}

// GetHash returns the transaction hash
func (tx *BaseTransaction) GetHash() string {
	return tx.Hash
}

// GetFrom returns the sender address
func (tx *BaseTransaction) GetFrom() string {
	return tx.From
}

// GetTo returns the recipient address
func (tx *BaseTransaction) GetTo() string {
	return tx.To
}

// GetValue returns the amount of native currency transferred
func (tx *BaseTransaction) GetValue() *big.Int {
	return tx.Value
}

// GetData returns the transaction input data
func (tx *BaseTransaction) GetData() []byte {
	return tx.Data
}

// GetNonce returns the transaction nonce
func (tx *BaseTransaction) GetNonce() uint64 {
	return tx.Nonce
}

// GetGasPrice returns the price per unit of gas
func (tx *BaseTransaction) GetGasPrice() *big.Int {
	return tx.GasPrice
}

// GetGasLimit returns the maximum gas units the transaction can consume
func (tx *BaseTransaction) GetGasLimit() uint64 {
	return tx.GasLimit
}

// GetType returns the general nature of the transaction
func (tx *BaseTransaction) GetType() TransactionType {
	return tx.Type
}

// GetTransaction returns the transaction
func (tx *Transaction) GetTransaction() *Transaction {
	return tx
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
	addr, err := NewAddress(l.ChainType, addressHex)
	if err != nil {
		// Propagate the error from NewAddress directly (it should be a Vault0Error)
		return nil, err
	}

	return addr, nil
}
