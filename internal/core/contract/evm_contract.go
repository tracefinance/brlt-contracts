package contract

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/wallet"
	coreTypes "vault0/internal/types"
)

// EVMSmartContract implements the SmartContract interface for EVM compatible chains
type EVMSmartContract struct {
	// chainType is the type of blockchain
	chainType coreTypes.ChainType
	// blockchain is the blockchain client
	blockchain blockchain.Blockchain
	// wallet is the wallet client
	wallet wallet.Wallet
	// config is the app configuration
	config config.Config
}

// NewEVMSmartContract creates a new EVM contract manager
func NewEVMSmartContract(
	blockchain blockchain.Blockchain,
	wallet wallet.Wallet,
	config config.Config,
) (*EVMSmartContract, error) {
	// Get chain information from wallet
	chain := wallet.Chain()

	// Validate chain type
	if chain.Type != coreTypes.ChainTypeEthereum &&
		chain.Type != coreTypes.ChainTypePolygon &&
		chain.Type != coreTypes.ChainTypeBase {
		return nil, fmt.Errorf("unsupported chain type: %s", chain.Type)
	}

	return &EVMSmartContract{
		chainType:  chain.Type,
		blockchain: blockchain,
		wallet:     wallet,
		config:     config,
	}, nil
}

// ChainType returns the blockchain type
func (c *EVMSmartContract) ChainType() coreTypes.ChainType {
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
			return nil, fmt.Errorf("contract artifact not found: %s", contractName)
		}
	}

	// Read the artifact file
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract artifact: %w", err)
	}

	// Parse the artifact JSON - supporting both Hardhat and Truffle formats
	var artifactData map[string]any
	if err := json.Unmarshal(data, &artifactData); err != nil {
		return nil, fmt.Errorf("failed to parse contract artifact: %w", err)
	}

	// Extract the ABI
	abiData, ok := artifactData["abi"]
	if !ok {
		return nil, errors.New("ABI not found in contract artifact")
	}

	abiJSON, err := json.Marshal(abiData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ABI: %w", err)
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
		return nil, errors.New("bytecode not found in contract artifact")
	}

	// Remove 0x prefix if present
	bytecodeHex = strings.TrimPrefix(bytecodeHex, "0x")

	// Convert hex to bytes
	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bytecode: %w", err)
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
			return nil, fmt.Errorf("failed to decode deployed bytecode: %w", err)
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
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Encode constructor parameters
	constructorInput := []byte{}
	if len(options.ConstructorArgs) > 0 {
		var err error
		constructorInput, err = parsedABI.Pack("", options.ConstructorArgs...)
		if err != nil {
			return nil, fmt.Errorf("failed to encode constructor arguments: %w", err)
		}
	}

	// Prepare deployment data (bytecode + constructor args)
	deployData := append(artifact.Bytecode, constructorInput...)

	// Prepare transaction options
	txOptions := coreTypes.TransactionOptions{
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
		coreTypes.ZeroAddress, // Use zero address for contract deployment
		options.Value,
		txOptions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment transaction: %w", err)
	}

	// Sign the transaction
	signedTx, err := c.wallet.SignTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign deployment transaction: %w", err)
	}

	// Broadcast the transaction
	txHash, err := c.blockchain.BroadcastTransaction(ctx, signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to broadcast deployment transaction: %w", err)
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
		return nil, fmt.Errorf("failed to get transaction receipt: %w", err)
	}

	// Check if the transaction was successful
	if receipt.Status == 0 {
		return nil, errors.New("contract deployment failed")
	}

	// Get transaction details
	tx, err := c.blockchain.GetTransaction(ctx, transactionHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction details: %w", err)
	}

	// Calculate deployment cost
	gasPrice := tx.GasPrice
	gasUsed := receipt.GasUsed
	deploymentCost := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasUsed)))

	// Convert block number to uint64
	blockNumber := uint64(0)
	if receipt.BlockNumber != nil {
		blockNumber = receipt.BlockNumber.Uint64()
	}

	// Return deployment result
	return &DeploymentResult{
		ContractAddress: receipt.Logs[0].Address, // Contract address is usually in the first log
		TransactionHash: transactionHash,
		BlockNumber:     blockNumber,
		DeploymentCost:  deploymentCost,
		GasUsed:         gasUsed,
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
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the method call data
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to encode method call data: %w", err)
	}

	// Call the contract
	result, err := c.blockchain.CallContract(ctx, "", contractAddress, data)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract method: %w", err)
	}

	// Unpack the result
	methodABI, exists := parsedABI.Methods[method]
	if !exists {
		return nil, fmt.Errorf("method %s not found in ABI", method)
	}

	// Prepare output variables array
	outputs := make([]any, 0, len(methodABI.Outputs))

	// Create an appropriate map to unpack outputs into
	outputMap := make(map[string]any)

	// Unpack the result into the output map
	if err := parsedABI.UnpackIntoMap(outputMap, method, result); err != nil {
		return nil, fmt.Errorf("failed to decode method return data: %w", err)
	}

	// Convert map values to array in the correct order
	for _, output := range methodABI.Outputs {
		if val, ok := outputMap[output.Name]; ok {
			outputs = append(outputs, val)
		} else {
			// For unnamed outputs, we might need to reconstruct from a different approach
			outputs = append(outputs, nil) // Placeholder for unnamed outputs
		}
	}

	return outputs, nil
}

// ExecuteMethod executes a state-changing method on a deployed contract
func (c *EVMSmartContract) ExecuteMethod(
	ctx context.Context,
	contractAddress string,
	artifact *Artifact,
	method string,
	options coreTypes.TransactionOptions,
	args ...any,
) (string, error) {
	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(artifact.ABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the method call data
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		return "", fmt.Errorf("failed to encode method call data: %w", err)
	}

	// Set the transaction data
	options.Data = data

	// Create the transaction
	tx, err := c.wallet.CreateNativeTransaction(
		ctx,
		contractAddress,
		big.NewInt(0), // No value to send for normal method calls
		options,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create method transaction: %w", err)
	}

	// Sign the transaction
	signedTx, err := c.wallet.SignTransaction(ctx, tx)
	if err != nil {
		return "", fmt.Errorf("failed to sign method transaction: %w", err)
	}

	// Broadcast the transaction
	txHash, err := c.blockchain.BroadcastTransaction(ctx, signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to broadcast method transaction: %w", err)
	}

	return txHash, nil
}
