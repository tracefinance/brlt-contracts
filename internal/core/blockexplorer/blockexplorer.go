package blockexplorer

import (
	"context"
	"errors"
	"math/big"
	"vault0/internal/types"
)

// Common errors
var (
	ErrExplorerNotSupported = errors.New("blockexplorer: explorer not supported")
	ErrRateLimitExceeded    = errors.New("blockexplorer: rate limit exceeded")
	ErrInvalidAPIKey        = errors.New("blockexplorer: invalid API key")
	ErrInvalidAddress       = errors.New("blockexplorer: invalid address")
	ErrInvalidResponse      = errors.New("blockexplorer: invalid response from explorer")
	ErrRequestFailed        = errors.New("blockexplorer: request failed")
	ErrMissingAPIKey        = errors.New("blockexplorer: missing API key")
)

// TransactionType represents different transaction categories for filtering
type TransactionType string

const (
	// TxTypeNormal represents standard transactions
	TxTypeNormal TransactionType = "normal"
	// TxTypeInternal represents internal transactions
	TxTypeInternal TransactionType = "internal"
	// TxTypeERC20 represents ERC20 token transfers
	TxTypeERC20 TransactionType = "erc20"
	// TxTypeERC721 represents ERC721 token transfers (NFTs)
	TxTypeERC721 TransactionType = "erc721"
)

// TransactionHistoryOptions contains parameters for fetching transaction history
type TransactionHistoryOptions struct {
	// StartBlock is the starting block for the history query
	StartBlock int64
	// EndBlock is the ending block for the history query
	EndBlock int64
	// Page is the page number for pagination
	Page int
	// PageSize is the number of items per page
	PageSize int
	// TransactionTypes filters specific transaction types
	TransactionTypes []TransactionType
	// SortAscending sorts by ascending timestamp if true
	SortAscending bool
}

// BlockExplorer defines the interface for interacting with blockchain explorers
// such as Etherscan, Polygonscan, etc.
//
// This interface is blockchain-agnostic and allows for different implementations
// for various blockchains while providing a consistent API. It supports querying
// transaction history, balances, and token information for any address on the
// supported blockchain.
//
// Example usage:
//
//	factory := blockexplorer.NewFactory(chains, cfg)
//	explorer, err := factory.GetExplorer(types.ChainTypeEthereum)
//	if err != nil {
//	    return err
//	}
//	defer explorer.Close()
//
//	// Fetch transaction history
//	options := TransactionHistoryOptions{
//	    StartBlock: 0,
//	    EndBlock: 0, // Latest block
//	    Page: 1,
//	    PageSize: 10,
//	}
//	txs, err := explorer.GetTransactionHistory(ctx, address, options)
type BlockExplorer interface {
	// GetTransactionHistory retrieves transaction history for an address.
	//
	// The method supports pagination and filtering through TransactionHistoryOptions.
	// It returns a slice of Transaction objects containing details such as hash,
	// from/to addresses, value, gas information, and status.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - address: The blockchain address to query (must be valid for the chain)
	//   - options: Configuration for filtering and pagination
	//
	// Returns:
	//   - []*types.Transaction: Slice of transactions matching the query
	//   - error: ErrInvalidAddress if address is invalid, ErrRateLimitExceeded if
	//     rate limit is hit, or other errors for API/network issues
	GetTransactionHistory(ctx context.Context, address string, options TransactionHistoryOptions) ([]*types.Transaction, error)

	// GetTransactionsByHash retrieves transaction details for multiple transaction hashes.
	//
	// This method is useful for getting detailed information about specific transactions
	// when you have their hashes. It fetches complete transaction data including status,
	// gas usage, and confirmations.
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - hashes: Slice of transaction hashes to look up
	//
	// Returns:
	//   - []*types.Transaction: Slice of transactions in the same order as the input hashes
	//   - error: ErrRequestFailed for API issues, ErrInvalidResponse for parsing errors
	GetTransactionsByHash(ctx context.Context, hashes []string) ([]*types.Transaction, error)

	// GetAddressBalance retrieves the native token balance for an address.
	//
	// The balance is returned in the smallest unit of the native currency
	// (e.g., Wei for Ethereum). For human-readable values, divide by the appropriate
	// number of decimals (e.g., 1e18 for ETH).
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - address: The blockchain address to query
	//
	// Returns:
	//   - *big.Int: The balance in the smallest unit of the native currency
	//   - error: ErrInvalidAddress if address is invalid, or API/network errors
	GetAddressBalance(ctx context.Context, address string) (*big.Int, error)

	// GetTokenBalances retrieves ERC20 token balances for an address.
	//
	// Returns a map of token contract addresses to their respective balances.
	// The balances are in the smallest unit of each token (need to be divided
	// by the token's decimals for human-readable values).
	//
	// Parameters:
	//   - ctx: Context for timeout and cancellation
	//   - address: The blockchain address to query
	//
	// Returns:
	//   - map[string]*big.Int: Map of token addresses to balances
	//   - error: ErrInvalidAddress if address is invalid, or API/network errors
	GetTokenBalances(ctx context.Context, address string) (map[string]*big.Int, error)

	// Close releases any resources used by the explorer.
	//
	// This should be called when the explorer is no longer needed to clean up
	// resources like rate limiters and HTTP clients.
	Close() error

	// Chain returns information about the blockchain this explorer is connected to.
	//
	// The returned Chain object contains details about the blockchain network,
	// including its type (e.g., Ethereum, Polygon), network ID, and other
	// chain-specific information.
	//
	// Returns:
	//   - types.Chain: Information about the blockchain
	Chain() types.Chain
}
