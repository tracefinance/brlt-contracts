package blockexplorer

import (
	"context"
	"vault0/internal/types"
)

// BlockExplorer defines the interface for interacting with blockchain explorers
// such as Etherscan, Polygonscan, etc.
//
// This interface is blockchain-agnostic and allows for different implementations
// for various blockchains while providing a consistent API. It supports querying
// transaction history, balances, and token information for any address on the
// supported blockchain.
type BlockExplorer interface {
	// GetTransactionHistory retrieves the transaction history for a given address with pagination.
	// It supports filtering by transaction type (normal, internal, ERC20, ERC721) and block range.
	GetTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions, nextToken string) (*types.Page[any], error)

	// GetTransactionByHash retrieves detailed information about a specific transaction
	// given its hash. This is useful for getting the current state of a transaction
	// or verifying its execution status.
	//
	// Returns:
	//   - ErrExplorerRequestFailed for API/network issues
	//   - ErrInvalidExplorerResponse if the response cannot be parsed
	//   - ErrTransactionNotFound if the transaction cannot be found
	GetTransactionByHash(ctx context.Context, hash string) (*types.Transaction, error)

	// GetTransactionReceiptByHash retrieves the receipt of a transaction by its hash.
	// A transaction receipt contains information about the execution of a transaction,
	// including gas used, logs generated, and execution status.
	//
	// Returns:
	//   - ErrExplorerRequestFailed for API/network issues
	//   - ErrInvalidExplorerResponse if the response cannot be parsed
	//   - ErrTransactionNotFound if the transaction receipt cannot be found
	GetTransactionReceiptByHash(ctx context.Context, hash string) (*types.TransactionReceipt, error)

	// GetContract retrieves detailed information about a smart contract.
	// This includes the contract's ABI, source code (if verified), and other metadata.
	//
	// Returns:
	//   - ErrContractNotFound if no contract exists at the address
	//   - ErrExplorerRequestFailed for API/network issues
	GetContract(ctx context.Context, address string) (*ContractInfo, error)

	// GetTokenURL returns the block explorer's URL for viewing a token's details.
	// This URL can be used to direct users to the block explorer's web interface.
	GetTokenURL(address string) string

	// Chain returns the configuration for the blockchain this explorer is connected to.
	// This includes information like the chain type, network ID, and other chain-specific details.
	Chain() types.Chain
}
