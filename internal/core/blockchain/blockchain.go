package blockchain

import (
	"context"
	"math/big"
	"vault0/internal/types"
)

// Blockchain defines a blockchain client interface that abstracts interactions with
// various blockchain networks. It provides methods for querying blockchain state,
// submitting transactions, interacting with smart contracts, and subscribing to events.
//
// This interface is designed to be blockchain-agnostic, allowing implementations
// for different blockchain networks (Ethereum, Polygon, etc.) while providing a
// consistent API for application code. All blockchain-specific details should be
// handled by the implementation.
//
// Most methods accept a context.Context parameter to allow for cancellation and
// timeouts when interacting with the blockchain network.
type Blockchain interface {
	// GetTransaction retrieves transaction details by hash.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - hash: Transaction hash in hexadecimal format
	//
	// Returns:
	//   - Transaction details if found
	//   - Error if transaction cannot be retrieved or doesn't exist
	GetTransaction(ctx context.Context, hash string) (*types.Transaction, error)

	// GetBlock retrieves block details by number or hash.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - identifier: Block number as a string or hash in hexadecimal format
	//                 Special values: "latest", "earliest", "pending"
	//
	// Returns:
	//   - Block details if found
	//   - Error if block cannot be retrieved or doesn't exist
	GetBlock(ctx context.Context, identifier string) (*types.Block, error)

	// GetTransactionReceipt retrieves transaction receipt by hash.
	// The receipt contains execution results including status, logs, and gas used.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - hash: Transaction hash in hexadecimal format
	//
	// Returns:
	//   - Transaction receipt if the transaction has been mined
	//   - Error if receipt cannot be retrieved or transaction hasn't been mined
	GetTransactionReceipt(ctx context.Context, hash string) (*types.TransactionReceipt, error)

	// EstimateGas estimates the gas needed to execute a transaction.
	// This performs a simulation of the transaction execution without committing to the blockchain.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - tx: Transaction object containing the operation details
	//
	// Returns:
	//   - Estimated gas amount required for the transaction
	//   - Error if estimation fails (e.g., transaction would revert)
	EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error)

	// BroadcastTransaction broadcasts a signed transaction to the network.
	// The transaction must be properly signed before broadcasting.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - signedTx: RLP-encoded signed transaction bytes
	//
	// Returns:
	//   - Transaction hash if successfully broadcasted
	//   - Error if broadcasting fails or transaction is invalid
	BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error)

	// GetBalance retrieves the balance of an address.
	// The balance is returned in the smallest denomination (e.g., wei for Ethereum).
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - address: Account address in the blockchain's format
	//
	// Returns:
	//   - Account balance as a big integer
	//   - Error if balance cannot be retrieved
	GetBalance(ctx context.Context, address string) (*big.Int, error)

	// GetNonce retrieves the next nonce for an address.
	// The nonce is used to prevent transaction replay and must be included in transactions.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - address: Account address in the blockchain's format
	//
	// Returns:
	//   - Next nonce to use for transactions from this address
	//   - Error if nonce cannot be retrieved
	GetNonce(ctx context.Context, address string) (uint64, error)

	// GetGasPrice retrieves the current gas price.
	// This is the recommended price per gas unit for timely transaction processing.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//
	// Returns:
	//   - Current gas price as a big integer
	//   - Error if gas price cannot be retrieved
	GetGasPrice(ctx context.Context) (*big.Int, error)

	// CallContract executes a read-only call to a smart contract.
	// This simulates the contract execution without creating a transaction.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - from: Address to execute the call from (can affect execution context)
	//   - to: Contract address to call
	//   - data: ABI-encoded function call data
	//
	// Returns:
	//   - Result data from the contract call
	//   - Error if the call fails or reverts
	CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error)

	// FilterContractLogs retrieves historical logs matching the filter criteria.
	// Logs are events emitted by smart contracts during transaction execution.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used for cancellation
	//   - addresses: List of contract addresses to filter logs from (empty for all)
	//   - eventSignature: The signature of the event (e.g., "Transfer(address indexed from, address indexed to, uint256 value)")
	//   - eventArgs: The arguments to filter by (nil or empty for no filtering)
	//   - fromBlock: Starting block number for the filter (negative for latest)
	//   - toBlock: Ending block number for the filter (negative for latest)
	//
	// Returns:
	//   - Array of logs matching the filter criteria
	//   - Error if logs cannot be retrieved
	FilterContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock, toBlock int64) ([]types.Log, error)

	// SubscribeContractLogs subscribes to live events matching the filter criteria.
	// This creates a real-time subscription to contract events as they occur.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used to cancel the subscription
	//   - addresses: List of contract addresses to filter events from (empty for all)
	//   - eventSignature: The signature of the event (e.g., "Transfer(address indexed from, address indexed to, uint256 value)")
	//   - eventArgs: The arguments to filter by (nil or empty for no filtering)
	//   - fromBlock: The block number to start the subscription from:
	//     * Positive value: Start from the specified block number
	//     * Zero or negative value: Implementation will default to (current block - 50,000)
	//       to prevent exceeding maximum block range limitations
	//
	// Returns:
	//   - Channel that receives matching log events in real-time
	//   - Channel that receives subscription errors
	//   - Error if subscription cannot be created
	SubscribeContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock int64) (<-chan types.Log, <-chan error, error)

	// SubscribeNewHead subscribes to new block headers as they are mined.
	// This creates a real-time subscription to receive new blocks as they are added to the chain.
	//
	// Parameters:
	//   - ctx: Context for the operation, can be used to cancel the subscription
	//
	// Returns:
	//   - Channel that receives new block headers in real-time
	//   - Channel that receives subscription errors
	//   - Error if subscription cannot be created
	SubscribeNewHead(ctx context.Context) (<-chan types.Block, <-chan error, error)

	// Chain returns the chain information.
	// This includes details like chain ID, network name, and other chain-specific data.
	//
	// Returns:
	//   - Chain information object
	Chain() types.Chain

	// Close closes any open connections.
	// This should be called when the blockchain client is no longer needed.
	Close()
}
