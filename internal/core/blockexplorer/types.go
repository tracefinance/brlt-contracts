package blockexplorer

import (
	"math/big"

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

// TransactionHistoryEntry represents a transaction entry returned by explorer history APIs.
// It embeds the core BaseTransaction and adds common execution details available from history.
type TransactionHistoryEntry struct {
	types.BaseTransaction
	Status      types.TransactionStatus
	Timestamp   int64
	BlockNumber *big.Int
	GasUsed     uint64
}

// NormalTxHistoryEntry represents a normal transaction from history.
type NormalTxHistoryEntry struct {
	TransactionHistoryEntry
	// ContractAddress is set if this normal transaction resulted in a contract deployment.
	ContractAddress string
}

// InternalTxHistoryEntry represents an internal transaction from history.
type InternalTxHistoryEntry struct {
	TransactionHistoryEntry
	// Note: Internal transactions via Etherscan lack Nonce, GasPrice, GasLimit, Data.
	// The embedded BaseTransaction fields for these will be zero/nil.
}

// ERC20TxHistoryEntry represents an ERC20 token transfer event from history.
type ERC20TxHistoryEntry struct {
	TransactionHistoryEntry
	TokenAddress   string
	TokenSymbol    string
	TokenRecipient string   // The actual recipient of the token transfer
	TokenAmount    *big.Int // The amount of tokens transferred
}
