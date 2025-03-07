package smartcontract

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

// Contract defines methods for interacting with smart contracts
type Contract interface {
	// LoadArtifact loads a contract artifact from the filesystem
	LoadArtifact(ctx context.Context, contractName string) (*Artifact, error)

	// DeployContract deploys a smart contract to the blockchain
	DeployContract(
		ctx context.Context,
		keyID string,
		artifact *Artifact,
		options DeploymentOptions,
	) (*DeploymentResult, error)

	// WaitForDeployment waits for a contract deployment to complete
	WaitForDeployment(
		ctx context.Context,
		transactionHash string,
	) (*DeploymentResult, error)

	// CallMethod calls a read-only method on a deployed contract
	CallMethod(
		ctx context.Context,
		contractAddress string,
		artifact *Artifact,
		method string,
		args ...any,
	) ([]any, error)

	// ExecuteMethod executes a state-changing method on a deployed contract
	ExecuteMethod(
		ctx context.Context,
		keyID string,
		contractAddress string,
		artifact *Artifact,
		method string,
		options types.TransactionOptions,
		args ...any,
	) (string, error)
}
