package types

import "math/big"

// ERC721EventSignature defines the type for ERC721 event signatures.
type ERC721EventSignature string

// ERC721 event signatures
const (
	// ERC721TransferEvent is the standard ERC721 Transfer event signature: Transfer(address,address,uint256)
	ERC721TransferEvent ERC721EventSignature = "Transfer(address,address,uint256)"
	// ERC721ApprovalEvent is the standard ERC721 Approval event signature: Approval(address,address,uint256)
	ERC721ApprovalEvent ERC721EventSignature = "Approval(address,address,uint256)"
	// ERC721ApprovalForAllEvent is the standard ERC721 ApprovalForAll event signature: ApprovalForAll(address,address,bool)
	ERC721ApprovalForAllEvent ERC721EventSignature = "ApprovalForAll(address,address,bool)"
)

// ERC721MethodSignature defines the type for ERC721 method signatures.
type ERC721MethodSignature string

// ERC721 method signatures
const (
	ERC721BalanceOfMethod            ERC721MethodSignature = "balanceOf(address)"
	ERC721OwnerOfMethod              ERC721MethodSignature = "ownerOf(uint256)"
	ERC721SafeTransferFromMethod     ERC721MethodSignature = "safeTransferFrom(address,address,uint256)"
	ERC721SafeTransferFromDataMethod ERC721MethodSignature = "safeTransferFrom(address,address,uint256,bytes)"
	ERC721TransferFromMethod         ERC721MethodSignature = "transferFrom(address,address,uint256)"
	ERC721ApproveMethod              ERC721MethodSignature = "approve(address,uint256)"
	ERC721SetApprovalForAllMethod    ERC721MethodSignature = "setApprovalForAll(address,bool)"
	ERC721GetApprovedMethod          ERC721MethodSignature = "getApproved(uint256)"
	ERC721IsApprovedForAllMethod     ERC721MethodSignature = "isApprovedForAll(address,address)"
	ERC721NameMethod                 ERC721MethodSignature = "name()"
	ERC721SymbolMethod               ERC721MethodSignature = "symbol()"
	ERC721TokenURIMethod             ERC721MethodSignature = "tokenURI(uint256)"
)

// ERC721 transaction type constants
const (
	// TransactionTypeERC721Transfer indicates a transaction is an ERC721 token transfer
	TransactionTypeERC721Transfer TransactionType = "erc721_transfer"
)

// ERC721Transfer represents an ERC721 token transfer transaction.
// It embeds Transaction for the core details and adds token-specific fields.
type ERC721Transfer struct {
	// Embeds the core transaction details
	Transaction
	// TokenAddress is the address of the ERC721 token contract
	TokenAddress string
	// TokenSymbol is the symbol of the ERC721 token (optional)
	TokenSymbol string
	// TokenName is the name of the ERC721 token (optional)
	TokenName string
	// TokenID is the unique identifier of the transferred token
	TokenID *big.Int
	// Recipient is the address receiving the token
	Recipient string
}
