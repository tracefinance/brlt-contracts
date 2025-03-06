package blockchain

import (
	"context"
	"errors"
	"math/big"
	"vault0/internal/wallet"
)

// Blockchain errors
var (
	ErrChainNotSupported   = errors.New("blockchain: chain not supported")
	ErrInvalidAddress      = errors.New("blockchain: invalid address")
	ErrTransactionFailed   = errors.New("blockchain: transaction failed")
	ErrInsufficientFunds   = errors.New("blockchain: insufficient funds")
	ErrInvalidAmount       = errors.New("blockchain: invalid amount")
	ErrInvalidNonce        = errors.New("blockchain: invalid nonce")
	ErrInvalidGasPrice     = errors.New("blockchain: invalid gas price")
	ErrInvalidGasLimit     = errors.New("blockchain: invalid gas limit")
	ErrInvalidContractCall = errors.New("blockchain: invalid contract call")
	ErrRPCConnectionFailed = errors.New("blockchain: RPC connection failed")
)

// ChainType represents a specific blockchain type
type ChainType string

// Supported chain types
const (
	Ethereum ChainType = "ethereum"
	Polygon  ChainType = "polygon"
	Base     ChainType = "base"
)

// Chain represents information about a specific blockchain
type Chain struct {
	ID          int64     // Chain ID
	Type        ChainType // Chain type
	Name        string    // Human-readable name
	Symbol      string    // Native currency symbol
	RPCUrl      string    // RPC URL for the chain
	ExplorerUrl string    // Block explorer URL
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

// Blockchain defines the interface for interacting with blockchains
type Blockchain interface {
	// GetChainID returns the chain ID
	GetChainID(ctx context.Context) (int64, error)

	// GetBalance gets the balance for an address
	// If blockNumber is nil, the latest block is used
	GetBalance(ctx context.Context, address string, blockNumber *big.Int) (*big.Int, error)

	// GetNonce gets the next nonce for an address
	GetNonce(ctx context.Context, address string) (uint64, error)

	// GetTransaction gets transaction information by hash
	GetTransaction(ctx context.Context, hash string) (*wallet.Transaction, error)

	// GetTransactionReceipt gets a transaction receipt by hash
	GetTransactionReceipt(ctx context.Context, hash string) (*TransactionReceipt, error)

	// EstimateGas estimates the gas required for a transaction
	EstimateGas(ctx context.Context, tx *wallet.Transaction) (uint64, error)

	// GetGasPrice gets the current gas price
	GetGasPrice(ctx context.Context) (*big.Int, error)

	// CallContract performs a contract call without creating a transaction
	// from is optional and can be empty
	CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error)

	// SendTransaction sends a transaction to the network
	SendTransaction(ctx context.Context, rawTx []byte) (string, error)

	// Close closes the client connection
	Close()
}
