package types

import "math/big"

type ERC20EventSignature string

// ERC20 event signatures
const (
	// ERC20TransferEventSignature is the standard ERC20 Transfer event signature
	ERC20TransferEventSignature ERC20EventSignature = "Transfer(address,address,uint256)"

	// ERC20ApprovalEventSignature is the standard ERC20 Approval event signature
	ERC20ApprovalEventSignature ERC20EventSignature = "Approval(address,address,uint256)"
)

type ERC20MethodSignature string

// ERC20 method signatures
const (
	// ERC20TransferMethod is the standard ERC20 Transfer method signature
	ERC20TransferMethod ERC20MethodSignature = "transfer(address,uint256)"

	// ERC20BalanceOfMethod is the standard ERC20 balanceOf method signature
	ERC20BalanceOfMethod ERC20MethodSignature = "balanceOf(address)"

	// ERC20ApproveMethod is the standard ERC20 approve method signature
	ERC20ApproveMethod ERC20MethodSignature = "approve(address,uint256)"

	// ERC20AllowanceMethod is the standard ERC20 allowance method signature
	ERC20AllowanceMethod ERC20MethodSignature = "allowance(address,address)"

	// ERC20TransferFromMethod is the standard ERC20 transferFrom method signature
	ERC20TransferFromMethod ERC20MethodSignature = "transferFrom(address,address,uint256)"

	// ERC20NameMethod is the standard ERC20 name method signature
	ERC20NameMethod ERC20MethodSignature = "name()"

	// ERC20SymbolMethod is the standard ERC20 symbol method signature
	ERC20SymbolMethod ERC20MethodSignature = "symbol()"

	// ERC20DecimalsMethod is the standard ERC20 decimals method signature
	ERC20DecimalsMethod ERC20MethodSignature = "decimals()"

	// ERC20TotalSupplyMethod is the standard ERC20 totalSupply method signature
	ERC20TotalSupplyMethod ERC20MethodSignature = "totalSupply()"
)

// ERC20 transaction type constants
const (
	// TransactionTypeERC20Transfer indicates a transaction is an ERC20 token transfer
	TransactionTypeERC20Transfer TransactionType = "erc20_transfer"
)

// ERC20Transfer represents an ERC20 token transfer transaction
// It embeds BaseTransaction for the core details and adds token-specific fields
type ERC20Transfer struct {
	// Embeds the core transaction details
	BaseTransaction
	// TokenAddress is the address of the ERC20 token contract
	TokenAddress string
	// TokenSymbol is the symbol of the ERC20 token
	TokenSymbol string
	// Recipient is the address receiving the tokens
	Recipient string
	// Amount is the amount of tokens transferred
	Amount *big.Int
}
