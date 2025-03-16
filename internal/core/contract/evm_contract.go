package contract

import (
	"context"
	"encoding/hex"
	"encoding/json"
	stderrors "errors"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/wallet"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// EVMSmartContract implements the SmartContract interface for EVM compatible chains
type EVMSmartContract struct {
	// chainType is the type of blockchain
	chainType types.ChainType
	// blockchain is the blockchain client
	blockchain blockchain.Blockchain
	// wallet is the wallet client
	wallet wallet.Wallet
	// config is the app configuration
	config *config.Config
}

// NewEVMSmartContract creates a new EVM contract manager
func NewEVMSmartContract(
	blockchain blockchain.Blockchain,
	wallet wallet.Wallet,
	config *config.Config,
) (*EVMSmartContract, error) {
	// Get chain information from wallet
	chain := wallet.Chain()

	// Validate chain type
	if chain.Type != types.ChainTypeEthereum &&
		chain.Type != types.ChainTypePolygon &&
		chain.Type != types.ChainTypeBase {
		return nil, errors.NewChainNotSupportedError(string(chain.Type))
	}

	return &EVMSmartContract{
		chainType:  chain.Type,
		blockchain: blockchain,
		wallet:     wallet,
		config:     config,
	}, nil
}

// ChainType returns the blockchain type
func (c *EVMSmartContract) ChainType() types.ChainType {
	return c.wallet.Chain().Type
}

// LoadArtifact loads a contract artifact from the filesystem
func (c *EVMSmartContract) LoadArtifact(ctx context.Context, contractName string) (*Artifact, error) {
	// Get the path to the smart contracts directory
	contractsPath := c.config.GetSmartContractsPath()

	// Construct the artifact path - supporting both Hardhat and Truffle formats
	// Try Hardhat format first
	artifactPath := filepath.Join(contractsPath, contractName+".json")

	// If file doesn't exist, try looking in a subdirectory (Truffle format)
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		artifactPath = filepath.Join(contractsPath, contractName, contractName+".json")

		// If still doesn't exist, return error
		if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
			return nil, errors.NewInvalidContractError(contractName, err)
		}
	}

	// Read the artifact file
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, err)
	}

	// Parse the artifact JSON - supporting both Hardhat and Truffle formats
	var artifactData map[string]any
	if err := json.Unmarshal(data, &artifactData); err != nil {
		return nil, errors.NewInvalidContractError(contractName, err)
	}

	// Extract the ABI
	abiData, ok := artifactData["abi"]
	if !ok {
		return nil, errors.NewInvalidContractError(contractName, stderrors.New("abi not found"))
	}

	abiJSON, err := json.Marshal(abiData)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, err)
	}

	// Extract the bytecode - checking multiple possible fields based on compiler
	var bytecodeHex string

	// Try Hardhat format first
	if bytecode, ok := artifactData["bytecode"]; ok {
		bytecodeHex, _ = bytecode.(string)
	}

	// Try Truffle format if not found
	if bytecodeHex == "" {
		if bytecode, ok := artifactData["unlinked_binary"]; ok {
			bytecodeHex, _ = bytecode.(string)
		}
	}

	// Handle old Truffle format
	if bytecodeHex == "" {
		if bytecode, ok := artifactData["code"]; ok {
			bytecodeHex, _ = bytecode.(string)
		}
	}

	if bytecodeHex == "" {
		return nil, errors.NewInvalidContractError(contractName, err)
	}

	// Remove 0x prefix if present
	bytecodeHex = strings.TrimPrefix(bytecodeHex, "0x")

	// Convert hex to bytes
	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, err)
	}

	// Extract deployed bytecode if available
	var deployedBytecodeHex string

	// Try Hardhat format
	if deployedBytecode, ok := artifactData["deployedBytecode"]; ok {
		deployedBytecodeHex, _ = deployedBytecode.(string)
	}

	// Try Truffle format
	if deployedBytecodeHex == "" {
		if deployedBytecode, ok := artifactData["deployed_bytecode"]; ok {
			deployedBytecodeHex, _ = deployedBytecode.(string)
		}
	}

	// Initialize deployed bytecode
	var deployedBytecode []byte
	if deployedBytecodeHex != "" {
		// Remove 0x prefix if present
		deployedBytecodeHex = strings.TrimPrefix(deployedBytecodeHex, "0x")

		// Convert hex to bytes
		deployedBytecode, err = hex.DecodeString(deployedBytecodeHex)
		if err != nil {
			return nil, errors.NewInvalidContractError(contractName, err)
		}
	}

	return &Artifact{
		Name:             contractName,
		ABI:              string(abiJSON),
		Bytecode:         bytecode,
		DeployedBytecode: deployedBytecode,
	}, nil
}

// Deploy deploys a smart contract to the blockchain
func (c *EVMSmartContract) Deploy(
	ctx context.Context,
	artifact *Artifact,
	options DeploymentOptions,
) (*DeploymentResult, error) {
	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(artifact.ABI))
	if err != nil {
		return nil, errors.NewInvalidContractError(artifact.Name, err)
	}

	// Encode constructor parameters
	constructorInput := []byte{}
	if len(options.ConstructorArgs) > 0 {
		var err error
		constructorInput, err = parsedABI.Pack("", options.ConstructorArgs...)
		if err != nil {
			return nil, errors.NewInvalidContractError(artifact.Name, err)
		}
	}

	// Prepare deployment data (bytecode + constructor args)
	deployData := append(artifact.Bytecode, constructorInput...)

	// Prepare transaction options
	txOptions := types.TransactionOptions{
		GasPrice: options.GasPrice,
		GasLimit: options.GasLimit,
		Data:     deployData,
	}

	// Set nonce if provided
	if options.Nonce != nil {
		txOptions.Nonce = *options.Nonce
	}

	// Create transaction - using a zero address as "to" for contract creation
	tx, err := c.wallet.CreateNativeTransaction(
		ctx,
		types.ZeroAddress, // Use zero address for contract deployment
		options.Value,
		txOptions,
	)
	if err != nil {
		return nil, err
	}

	// Sign the transaction
	signedTx, err := c.wallet.SignTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	// Broadcast the transaction
	txHash, err := c.blockchain.BroadcastTransaction(ctx, signedTx)
	if err != nil {
		return nil, err
	}

	// Return initial result with transaction hash, other fields will be populated by WaitForDeployment
	return &DeploymentResult{
		TransactionHash: txHash,
	}, nil
}

// WaitForDeployment waits for a contract deployment to complete
func (c *EVMSmartContract) WaitForDeployment(
	ctx context.Context,
	transactionHash string,
) (*DeploymentResult, error) {
	// Get transaction receipt
	receipt, err := c.blockchain.GetTransactionReceipt(ctx, transactionHash)
	if err != nil {
		return nil, errors.NewTransactionNotFoundError(transactionHash)
	}

	// Check if transaction was successful
	if receipt.Status == 0 {
		return nil, errors.NewTransactionFailedError(err)
	}

	// Get transaction to get gas price
	tx, err := c.blockchain.GetTransaction(ctx, transactionHash)
	if err != nil {
		return nil, err
	}

	// Calculate deployment cost
	deploymentCost := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(receipt.GasUsed)))

	// Return deployment result
	return &DeploymentResult{
		ContractAddress: receipt.Logs[0].Address, // Contract address is usually in the first log
		TransactionHash: transactionHash,
		BlockNumber:     receipt.BlockNumber.Uint64(),
		DeploymentCost:  deploymentCost,
		GasUsed:         receipt.GasUsed,
	}, nil
}

// CallMethod calls a read-only method on a deployed contract
func (c *EVMSmartContract) CallMethod(
	ctx context.Context,
	contractAddress string,
	artifact *Artifact,
	method string,
	args ...any,
) ([]any, error) {
	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(artifact.ABI))
	if err != nil {
		return nil, errors.NewInvalidContractError(artifact.Name, err)
	}

	// Pack the method call data
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, errors.NewInvalidContractCallError(artifact.Name, err)
	}

	// Call the contract
	result, err := c.blockchain.CallContract(ctx, "", contractAddress, data)
	if err != nil {
		return nil, err
	}

	// Unpack the result
	outputs, err := parsedABI.Unpack(method, result)
	if err != nil {
		return nil, errors.NewInvalidContractCallError(artifact.Name, err)
	}

	return outputs, nil
}

// ExecuteMethod executes a state-changing method on a deployed contract
func (c *EVMSmartContract) ExecuteMethod(
	ctx context.Context,
	contractAddress string,
	artifact *Artifact,
	method string,
	options types.TransactionOptions,
	args ...any,
) (string, error) {
	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(artifact.ABI))
	if err != nil {
		return "", errors.NewInvalidContractError(artifact.Name, err)
	}

	// Pack the method call data
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return "", errors.NewInvalidContractCallError(artifact.Name, err)
	}

	// Add method data to transaction options
	options.Data = data

	// Create transaction
	tx, err := c.wallet.CreateNativeTransaction(
		ctx,
		contractAddress,
		big.NewInt(0), // No value to send for normal method calls
		options,
	)
	if err != nil {
		return "", err
	}

	// Sign the transaction
	signedTx, err := c.wallet.SignTransaction(ctx, tx)
	if err != nil {
		return "", err
	}

	// Broadcast the transaction
	txHash, err := c.blockchain.BroadcastTransaction(ctx, signedTx)
	if err != nil {
		return "", err
	}

	return txHash, nil
}
