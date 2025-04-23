package types

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
