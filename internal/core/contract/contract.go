package contract

import (
	"context"
	"math/big"
	"vault0/internal/types"
)

// Artifact represents a compiled smart contract artifact
type Artifact struct {
	// Name is the name of the contract
	Name string
	// ABI is the Application Binary Interface in JSON format
	ABI string
	// Bytecode is the compiled bytecode of the contract
	Bytecode []byte
	// DeployedBytecode is the bytecode of the deployed contract
	DeployedBytecode []byte
}

// DeploymentResult contains information about a deployed contract
type DeploymentResult struct {
	// ContractAddress is the address of the deployed contract
	ContractAddress string
	// TransactionHash is the hash of the deployment transaction
	TransactionHash string
	// BlockNumber is the block number where the contract was deployed
	BlockNumber uint64
	// DeploymentCost is the gas used for deployment * gas price
	DeploymentCost *big.Int
	// GasUsed is the amount of gas used for the deployment
	GasUsed uint64
}

// DeploymentOptions contains options for deploying a contract
type DeploymentOptions struct {
	// GasPrice is the gas price to use for the deployment (nil for auto)
	GasPrice *big.Int
	// GasLimit is the gas limit to use for the deployment (0 for auto)
	GasLimit uint64
	// Value is the amount of native currency to send with the deployment
	Value *big.Int
	// Nonce is the nonce to use for the deployment (nil for auto)
	Nonce *uint64
	// Constructor arguments for the contract deployment
	ConstructorArgs []any
}

// SmartContract defines methods for interacting with smart contracts on the blockchain.
// It provides functionality for loading contract artifacts, deploying contracts,
// and executing contract methods (both read-only and state-changing operations).
type SmartContract interface {
	// LoadArtifact loads a contract artifact from the filesystem.
	// 
	// Parameters:
	//   - ctx: The context for the operation, which can be used for cancellation.
	//   - contractName: Name of the contract to load (without file extension).
	//
	// Returns:
	//   - *Artifact: The loaded contract artifact containing ABI and bytecode.
	//   - error: Any error encountered during loading, such as file not found or parsing errors.
	LoadArtifact(ctx context.Context, contractName string) (*Artifact, error)

	// Deploy deploys a smart contract to the blockchain using the provided artifact and options.
	// 
	// Parameters:
	//   - ctx: The context for the operation, which can be used for cancellation.
	//   - artifact: The contract artifact containing the bytecode and ABI to be deployed.
	//   - options: Deployment configuration including gas price, gas limit, and constructor arguments.
	//
	// Returns:
	//   - *DeploymentResult: Information about the deployed contract including address and transaction hash.
	//   - error: Any error encountered during deployment, such as insufficient funds or network issues.
	Deploy(
		ctx context.Context,
		artifact *Artifact,
		options DeploymentOptions,
	) (*DeploymentResult, error)

	// WaitForDeployment waits for a contract deployment transaction to be mined and confirmed on the blockchain.
	// This method is useful for ensuring a contract is fully deployed before interacting with it.
	// 
	// Parameters:
	//   - ctx: The context for the operation, which can be used for cancellation or timeout.
	//   - transactionHash: The hash of the deployment transaction to wait for.
	//
	// Returns:
	//   - *DeploymentResult: Information about the completed deployment including contract address and block number.
	//   - error: Any error encountered while waiting, such as transaction failure or context cancellation.
	WaitForDeployment(
		ctx context.Context,
		transactionHash string,
	) (*DeploymentResult, error)

	// CallMethod calls a read-only (view/pure) method on a deployed contract.
	// These calls don't modify blockchain state and don't consume gas.
	// 
	// Parameters:
	//   - ctx: The context for the operation, which can be used for cancellation.
	//   - contractAddress: The address of the deployed contract to interact with.
	//   - artifact: The contract artifact containing the ABI needed to encode the call.
	//   - method: The name of the method to call on the contract.
	//   - args: Variable number of arguments to pass to the contract method.
	//
	// Returns:
	//   - []any: Array of return values from the contract method call.
	//   - error: Any error encountered during the call, such as method not found or execution revert.
	CallMethod(
		ctx context.Context,
		contractAddress string,
		artifact *Artifact,
		method string,
		args ...any,
	) ([]any, error)

	// ExecuteMethod executes a state-changing method on a deployed contract.
	// These calls modify blockchain state, require gas, and result in a new transaction.
	// 
	// Parameters:
	//   - ctx: The context for the operation, which can be used for cancellation.
	//   - contractAddress: The address of the deployed contract to interact with.
	//   - artifact: The contract artifact containing the ABI needed to encode the transaction.
	//   - method: The name of the method to execute on the contract.
	//   - options: Transaction options including gas price, gas limit, and value to send.
	//   - args: Variable number of arguments to pass to the contract method.
	//
	// Returns:
	//   - string: The transaction hash of the executed transaction.
	//   - error: Any error encountered during execution, such as insufficient funds or execution revert.
	ExecuteMethod(
		ctx context.Context,
		contractAddress string,
		artifact *Artifact,
		method string,
		options types.TransactionOptions,
		args ...any,
	) (string, error)
}
