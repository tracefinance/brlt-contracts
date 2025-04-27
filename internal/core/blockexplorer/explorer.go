package blockexplorer

import (
	"context"
	"vault0/internal/types"
)

// ContractInfo represents detailed information about a smart contract
// retrieved from a blockchain explorer.
type ContractInfo struct {
	// ABI is the contract's Application Binary Interface in JSON format
	// It defines the methods and events available in the contract
	ABI string
	// ContractName is the name of the contract as defined in its source code
	ContractName string
	// SourceCode contains the verified source code of the contract
	// This will be empty if the contract is not verified
	SourceCode string
	// IsVerified indicates whether the contract's source code has been
	// verified and published on the block explorer
	IsVerified bool
}

// TransactionType represents different categories of blockchain transactions
// that can be queried from the explorer.
type TransactionType string

const (
	// TxTypeNormal represents standard blockchain transactions
	// These include native currency transfers and contract interactions
	TxTypeNormal TransactionType = "normal"
	// TxTypeInternal represents internal transactions
	// These are transfers of native currency triggered by smart contract execution
	TxTypeInternal TransactionType = "internal"
	// TxTypeERC20 represents ERC20 token transfer transactions
	// These track the movement of ERC20 tokens between addresses
	TxTypeERC20 TransactionType = "erc20"
	// TxTypeERC721 represents NFT transfer transactions
	// These track the movement of NFTs (ERC721 tokens) between addresses
	TxTypeERC721 TransactionType = "erc721"
)

// TransactionHistoryOptions contains parameters for filtering and paginating
// transaction history queries. It provides fine-grained control over which
// transactions are returned.
type TransactionHistoryOptions struct {
	// StartBlock specifies the earliest block to include in the query
	// Use 0 to start from the genesis block
	StartBlock int64
	// EndBlock specifies the latest block to include in the query
	// Use 0 to include up to the latest block
	EndBlock int64
	// TransactionType specifies which type of transactions to include
	// If not specified, normal transactions will be used
	TransactionType TransactionType
	// SortAscending determines the order of returned transactions
	// If true, transactions are sorted from oldest to newest
	// If false, transactions are sorted from newest to oldest
	SortAscending bool
	// Limit specifies the maximum number of transactions to return per page
	Limit int
}

// BlockExplorer defines the interface for interacting with blockchain explorers
// such as Etherscan, Polygonscan, etc.
//
// This interface is blockchain-agnostic and allows for different implementations
// for various blockchains while providing a consistent API. It supports querying
// transaction history, balances, and token information for any address on the
// supported blockchain.
type BlockExplorer interface {
	// GetTransactionHistory retrieves transaction history for an address with pagination.
	// It supports filtering by transaction type and sorting by timestamp.
	//
	// The method returns a Page containing the requested transactions and pagination info.
	// If TransactionType is not specified in options, TxTypeNormal will be used by default.
	// If nextToken is provided, it will be used for pagination.
	//
	// Returns:
	//   - ErrInvalidAddress if the address is invalid for the chain
	//   - ErrRateLimitExceeded if the explorer's rate limit is hit
	//   - Other explorer-specific errors for API/network issues
	GetTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions, nextToken string) (*types.Page[*types.Transaction], error)

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
