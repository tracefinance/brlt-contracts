package blockchain

import (
	"context"
	"crypto/elliptic"
	"errors"
	"math/big"
	"vault0/internal/keygen"
	"vault0/internal/types"
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

// Chain represents information about a specific blockchain
type Chain struct {
	ID          int64           // Chain ID
	Type        types.ChainType // Chain type
	Name        string          // Human-readable name
	Symbol      string          // Native currency symbol
	RPCUrl      string          // RPC URL for the chain
	ExplorerUrl string          // Block explorer URL
	Curve       elliptic.Curve  // Elliptic curve for key generation
	KeyType     keygen.KeyType  // Key type for the blockchain
}

// Blockchain defines methods for interacting with a blockchain
type Blockchain interface {
	// GetTransaction retrieves transaction details by hash
	GetTransaction(ctx context.Context, hash string) (*types.Transaction, error)

	// GetTransactionReceipt retrieves transaction receipt by hash
	GetTransactionReceipt(ctx context.Context, hash string) (*types.TransactionReceipt, error)

	// EstimateGas estimates the gas needed to execute a transaction
	EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error)

	// BroadcastTransaction broadcasts a signed transaction to the network
	BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error)

	// GetBalance retrieves the balance of an address
	GetBalance(ctx context.Context, address string) (*big.Int, error)

	// GetNonce retrieves the next nonce for an address
	GetNonce(ctx context.Context, address string) (uint64, error)

	// GetGasPrice retrieves the current gas price
	GetGasPrice(ctx context.Context) (*big.Int, error)

	// CallContract executes a read-only call to a smart contract
	CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error)

	// Chain returns the chain information
	Chain() Chain

	// Close closes any open connections
	Close()
}
