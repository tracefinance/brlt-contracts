package blockexplorer

import (
	"context"
	"math/big"
	"vault0/internal/types"
)

// TokenBalance represents a token balance for any blockchain.
// It includes all necessary information about a token and its balance
// for a specific address.
type TokenBalance struct {
	// TokenAddress is the contract address of the token
	TokenAddress string
	// TokenName is the human-readable name of the token
	TokenName string
	// TokenSymbol is the trading symbol of the token (e.g., "USDC", "DAI")
	TokenSymbol string
	// TokenDecimal specifies the number of decimal places for the token
	// For example, most ERC20 tokens use 18 decimals
	TokenDecimal uint8
	// Balance represents the token balance in its smallest unit
	// To get the actual balance, divide by 10^TokenDecimal
	Balance *big.Int
}

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
	// Page specifies which page of results to return (1-based)
	// Must be greater than 0
	Page int
	// PageSize specifies how many transactions to return per page
	// The actual number of returned items may be less than this value
	PageSize int
	// TransactionTypes filters which types of transactions to include
	// If empty, all transaction types will be included
	TransactionTypes []TransactionType
	// SortAscending determines the order of returned transactions
	// If true, transactions are sorted from oldest to newest
	// If false, transactions are sorted from newest to oldest
	SortAscending bool
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
	// If no TransactionTypes are specified in options, all types will be included.
	//
	// Returns:
	//   - ErrInvalidAddress if the address is invalid for the chain
	//   - ErrRateLimitExceeded if the explorer's rate limit is hit
	//   - Other explorer-specific errors for API/network issues
	GetTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions) (*types.Page[*types.Transaction], error)

	// GetTransactionsByHash retrieves detailed information about specific transactions
	// given their hashes. This is useful for getting the current state of transactions
	// or verifying their execution status.
	//
	// The returned transactions will be in the same order as the input hashes.
	// If a transaction is not found, it will be omitted from the results.
	//
	// Returns:
	//   - ErrExplorerRequestFailed for API/network issues
	//   - ErrInvalidExplorerResponse if the response cannot be parsed
	GetTransactionsByHash(ctx context.Context, hashes []string) ([]*types.Transaction, error)

	// GetAddressBalance retrieves the native token balance for an address
	// (e.g., ETH for Ethereum, MATIC for Polygon).
	//
	// The balance is returned in the smallest unit of the native currency
	// (e.g., Wei for Ethereum). To get the actual balance, divide by 10^18.
	//
	// Returns:
	//   - ErrInvalidAddress if the address is invalid
	//   - ErrExplorerRequestFailed for API/network issues
	GetAddressBalance(ctx context.Context, address string) (*big.Int, error)

	// GetTokenBalances retrieves all token balances for an address.
	// This includes both ERC20 and ERC721 tokens that the address has interacted with.
	//
	// The balances are returned in the smallest unit of each token.
	// To get the actual balance, divide by 10^TokenDecimal.
	//
	// Returns:
	//   - ErrInvalidAddress if the address is invalid
	//   - ErrExplorerRequestFailed for API/network issues
	GetTokenBalances(ctx context.Context, address string) ([]*TokenBalance, error)

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
