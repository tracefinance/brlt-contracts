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
	// ERC20TransferMethodSignature is the standard ERC20 Transfer method signature
	ERC20TransferMethodSignature ERC20MethodSignature = "transfer(address,uint256)"

	// ERC20BalanceOfMethodSignature is the standard ERC20 balanceOf method signature
	ERC20BalanceOfMethodSignature ERC20MethodSignature = "balanceOf(address)"

	// ERC20ApproveMethodSignature is the standard ERC20 approve method signature
	ERC20ApproveMethodSignature ERC20MethodSignature = "approve(address,uint256)"

	// ERC20AllowanceMethodSignature is the standard ERC20 allowance method signature
	ERC20AllowanceMethodSignature ERC20MethodSignature = "allowance(address,address)"

	// ERC20TransferFromMethodSignature is the standard ERC20 transferFrom method signature
	ERC20TransferFromMethodSignature ERC20MethodSignature = "transferFrom(address,address,uint256)"

	// ERC20NameMethodSignature is the standard ERC20 name method signature
	ERC20NameMethodSignature ERC20MethodSignature = "name()"

	// ERC20SymbolMethodSignature is the standard ERC20 symbol method signature
	ERC20SymbolMethodSignature ERC20MethodSignature = "symbol()"

	// ERC20DecimalsMethodSignature is the standard ERC20 decimals method signature
	ERC20DecimalsMethodSignature ERC20MethodSignature = "decimals()"

	// ERC20TotalSupplyMethodSignature is the standard ERC20 totalSupply method signature
	ERC20TotalSupplyMethodSignature ERC20MethodSignature = "totalSupply()"
)
