package contract

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/wallet"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// evmContractManager implements the SmartContract interface for EVM compatible chains
type evmContractManager struct {
	// chainType is the type of blockchain
	chainType types.ChainType
	// blockchain is the blockchain client
	blockchain blockchain.BlockchainClient
	// wallet is the wallet client
	wallet wallet.WalletManager
	// explorer is the block explorer client
	explorer blockexplorer.BlockExplorer
	// config is the app configuration
	config *config.Config
}

// NewEVMSmartContract creates a new EVM contract manager
func NewEVMSmartContract(
	blockchain blockchain.BlockchainClient,
	wallet wallet.WalletManager,
	config *config.Config,
) (*evmContractManager, error) {
	chain := wallet.Chain()
	if chain.Type != types.ChainTypeEthereum &&
		chain.Type != types.ChainTypePolygon &&
		chain.Type != types.ChainTypeBase {
		return nil, errors.NewChainNotSupportedError(string(chain.Type))
	}

	return &evmContractManager{
		chainType:  chain.Type,
		blockchain: blockchain,
		wallet:     wallet,
		config:     config,
	}, nil
}

// ChainType returns the blockchain type
func (c *evmContractManager) ChainType() types.ChainType {
	return c.wallet.Chain().Type
}

// LoadArtifact loads a contract artifact from the filesystem
func (c *evmContractManager) LoadArtifact(ctx context.Context, contractName string) (*Artifact, error) {
	// Get the path to the smart contracts directory
	contractsPath := c.config.GetSmartContractsPath()

	// Get the base name by removing any extension
	ext := filepath.Ext(contractName)
	baseName := strings.TrimSuffix(contractName, ext)

	// Construct the expected artifact path: {contractsPath}/{baseName}/{baseName}.json
	artifactPath := filepath.Join(contractsPath, baseName, baseName+".json")

	// Check if the file exists at the expected path
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		// If not found, return an error indicating the expected path
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("artifact file not found at expected path: %s", artifactPath))
	}

	// Read the artifact file
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to read artifact file %s: %w", artifactPath, err))
	}

	// Parse the artifact JSON - supporting both Hardhat and Truffle formats
	var artifactData map[string]any
	if err := json.Unmarshal(data, &artifactData); err != nil {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to parse artifact JSON %s: %w", artifactPath, err))
	}

	// Extract the ABI
	abiData, ok := artifactData["abi"]
	if !ok {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("abi not found in artifact file %s", artifactPath))
	}

	abiJSON, err := json.Marshal(abiData)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to marshal ABI from artifact file %s: %w", artifactPath, err))
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
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("bytecode not found in artifact file %s", artifactPath))
	}

	// Remove 0x prefix if present
	bytecodeHex = strings.TrimPrefix(bytecodeHex, "0x")

	// Convert hex to bytes
	bytecode, err := hex.DecodeString(bytecodeHex)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to decode bytecode from artifact file %s: %w", artifactPath, err))
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
			return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to decode deployed bytecode from artifact file %s: %w", artifactPath, err))
		}
	}

	return &Artifact{
		Name:             baseName,
		ABI:              string(abiJSON),
		Bytecode:         bytecode,
		DeployedBytecode: deployedBytecode,
	}, nil
}

// Deploy deploys a smart contract to the blockchain
func (c *evmContractManager) Deploy(
	ctx context.Context,
	artifact *Artifact,
	options DeploymentOptions,
) (*DeploymentResult, error) {
	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(artifact.ABI))
	if err != nil {
		return nil, errors.NewInvalidContractError(artifact.Name, fmt.Errorf("failed to parse ABI for deployment: %w", err))
	}

	// Encode constructor parameters
	constructorInput := []byte{}
	if len(options.ConstructorArgs) > 0 {
		var err error
		constructorInput, err = parsedABI.Pack("", options.ConstructorArgs...)
		if err != nil {
			return nil, errors.NewInvalidContractError(artifact.Name, fmt.Errorf("failed to pack constructor arguments: %w", err))
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

	// Create transaction
	tx, err := c.wallet.CreateNativeTransaction(
		ctx,
		"",
		options.Value,
		txOptions,
	)
	if err != nil {
		return nil, errors.NewTransactionCreationError("contract deployment", err)
	}

	// Sign the transaction
	signedTx, err := c.wallet.SignTransaction(ctx, tx)
	if err != nil {
		return nil, errors.NewTransactionSigningError(err)
	}

	// Broadcast the transaction
	txHash, err := c.blockchain.BroadcastTransaction(ctx, signedTx)
	if err != nil {
		return nil, errors.NewTransactionBroadcastError(err)
	}

	return &DeploymentResult{
		TransactionHash: txHash,
	}, nil
}

// GetDeployment waits for a contract deployment to complete
func (c *evmContractManager) GetDeployment(
	ctx context.Context,
	transactionHash string,
) (*DeploymentResult, error) {
	// Get transaction receipt
	receipt, err := c.blockchain.GetTransactionReceipt(ctx, transactionHash)
	if err != nil {
		if errors.IsError(err, errors.ErrCodeTransactionNotFound) {
			return nil, errors.NewTransactionNotFoundError(transactionHash)
		}
		return nil, errors.NewBlockchainError(fmt.Errorf("failed to get transaction receipt %s: %w", transactionHash, err))
	}

	// Check if transaction was successful
	if receipt.Status == 0 {
		return nil, errors.NewTransactionFailedError(fmt.Errorf("deployment transaction %s reverted", transactionHash))
	}

	// Use ContractAddress from receipt (if present)
	contractAddress := ""
	if receipt.ContractAddress != nil {
		contractAddress = *receipt.ContractAddress
	} else {
		return nil, errors.NewBlockchainError(fmt.Errorf("successful deployment transaction %s missing contract address in receipt", transactionHash))
	}

	// Get transaction to get gas price
	tx, err := c.blockchain.GetTransaction(ctx, transactionHash)
	if err != nil {
		return nil, errors.NewBlockchainError(fmt.Errorf("failed to get transaction %s for gas price: %w", transactionHash, err))
	}

	// Calculate deployment cost
	deploymentCost := new(big.Int).Mul(tx.GasPrice, big.NewInt(int64(receipt.GasUsed)))

	// Return deployment result
	return &DeploymentResult{
		ContractAddress: contractAddress,
		TransactionHash: transactionHash,
		BlockNumber:     receipt.BlockNumber.Uint64(),
		DeploymentCost:  deploymentCost,
		GasUsed:         receipt.GasUsed,
	}, nil
}

// CallMethod calls a read-only method on a deployed contract
func (c *evmContractManager) CallMethod(
	ctx context.Context,
	contractAddress string,
	contractABI string,
	method string,
	args ...any,
) ([]any, error) {
	var parsedABI *abi.ABI
	var err error

	// Use provided ABI if available, otherwise fetch/cache
	if contractABI != "" {
		// Parse the provided ABI string
		tmpABI, parseErr := abi.JSON(strings.NewReader(contractABI))
		if parseErr != nil {
			return nil, errors.NewInvalidContractError(contractAddress, fmt.Errorf("failed to parse provided ABI: %w", parseErr))
		}
		parsedABI = &tmpABI
	} else {
		// Get parsed ABI from cache or fetch it
		parsedABI, err = c.getContractABI(ctx, contractAddress)
		if err != nil {
			return nil, err // Propagate errors from ABI fetching/parsing
		}
		if parsedABI == nil { // Should not happen if getOrFetchABI returns nil err, but defensive check
			return nil, errors.NewInvalidContractError(contractAddress, fmt.Errorf("failed to obtain ABI for %s", contractAddress))
		}
	}

	// Pack the method call data using the obtained ABI
	data, err := parsedABI.Pack(method, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no method with id") || strings.Contains(err.Error(), "method '"+method+"' not found") {
			return nil, errors.NewMethodNotFoundError(method, contractAddress)
		}
		return nil, errors.NewInvalidContractCallError(contractAddress, fmt.Errorf("failed to pack method '%s' call data: %w", method, err))
	}

	// Call the contract
	result, err := c.blockchain.CallContract(ctx, "", contractAddress, data)
	if err != nil {
		return nil, errors.NewBlockchainError(fmt.Errorf("blockchain call failed for method '%s' on %s: %w", method, contractAddress, err))
	}

	// Unpack the result
	outputs, err := parsedABI.Unpack(method, result)
	if err != nil {
		return nil, errors.NewInvalidContractCallError(contractAddress, fmt.Errorf("failed to unpack result for method '%s': %w", method, err))
	}

	return outputs, nil
}

// ExecuteMethod executes a state-changing method on a deployed contract
func (c *evmContractManager) ExecuteMethod(
	ctx context.Context,
	contractAddress string,
	contractABI string,
	method string,
	options ExecutionOptions,
	args ...any,
) (string, error) {
	var parsedABI *abi.ABI
	var err error

	// Use provided ABI if available, otherwise fetch/cache
	if contractABI != "" {
		// Parse the provided ABI string
		tmpABI, parseErr := abi.JSON(strings.NewReader(contractABI))
		if parseErr != nil {
			return "", errors.NewInvalidContractError(contractAddress, fmt.Errorf("failed to parse provided ABI: %w", parseErr))
		}
		parsedABI = &tmpABI
	} else {
		// Get parsed ABI from cache or fetch it
		parsedABI, err = c.getContractABI(ctx, contractAddress)
		if err != nil {
			return "", err
		}
		if parsedABI == nil { // Should not happen if getOrFetchABI returns nil err, but defensive check
			return "", errors.NewInvalidContractError(contractAddress, fmt.Errorf("failed to obtain ABI for %s", contractAddress))
		}
	}

	// Pack the method call data using the obtained ABI
	// NOTE: We pack here primarily for validation, but the actual packing for the
	// transaction data will happen inside CreateContractCallTransaction.
	_, err = parsedABI.Pack(method, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no method with id") || strings.Contains(err.Error(), "method '"+method+"' not found") {
			return "", errors.NewMethodNotFoundError(method, contractAddress)
		}
		return "", errors.NewInvalidContractCallError(contractAddress, fmt.Errorf("failed to pack method '%s' call data: %w", method, err))
	}

	// Translate ExecuteOptions to types.TransactionOptions
	// We don't set Data here, as CreateContractCallTransaction will handle encoding.
	txOptions := types.TransactionOptions{
		GasPrice: options.GasPrice,
		GasLimit: options.GasLimit,
		Nonce:    options.Nonce,
		// Data field is intentionally omitted here
	}

	// Ensure value from ExecuteOptions is not nil (use 0 if it is)
	// This value will now be passed to CreateContractCallTransaction.
	callValue := options.Value
	if callValue == nil {
		callValue = big.NewInt(0)
	}

	// Create transaction using CreateContractCallTransaction
	// It requires the ABI string to perform the encoding itself.
	finalAbiString := contractABI // Use provided if available
	if finalAbiString == "" {
		// Marshal the parsed ABI back to string if it wasn't provided
		abiBytes, marshalErr := json.Marshal(parsedABI.Methods)
		if marshalErr != nil {
			return "", errors.NewInvalidContractError(contractAddress, fmt.Errorf("failed to marshal fetched ABI: %w", marshalErr))
		}
		finalAbiString = string(abiBytes)
	}

	tx, err := c.wallet.CreateContractCallTransaction(
		ctx,
		contractAddress,
		callValue,      // Pass the value from options
		finalAbiString, // Pass the ABI string
		method,         // Pass the method name
		args,           // Pass the arguments
		txOptions,
	)
	if err != nil {
		return "", errors.NewTransactionCreationError(fmt.Sprintf("method %s on %s", method, contractAddress), err)
	}

	// Sign the transaction
	signedTx, err := c.wallet.SignTransaction(ctx, tx)
	if err != nil {
		return "", errors.NewTransactionSigningError(err)
	}

	// Broadcast the transaction
	txHash, err := c.blockchain.BroadcastTransaction(ctx, signedTx)
	if err != nil {
		return "", errors.NewTransactionBroadcastError(err)
	}

	return txHash, nil
}

// getContractABI retrieves and parses the ABI for a given contract address.
// It prioritizes fetching from the block explorer if available.
func (c *evmContractManager) getContractABI(ctx context.Context, contractAddress string) (*abi.ABI, error) {
	// Check if block explorer is configured and available
	if c.explorer == nil {
		// If no explorer, we cannot fetch the ABI dynamically
		// Assuming NewConfigurationError exists and takes context string
		return nil, errors.NewConfigurationError(fmt.Sprintf("BlockExplorer not configured for chain %s, cannot fetch ABI for %s", c.chainType, contractAddress))
	}

	// Fetch contract info using the block explorer
	contractInfo, err := c.explorer.GetContract(ctx, contractAddress)
	if err != nil {
		// Propagate errors from the explorer (e.g., network error, contract not found)
		return nil, err
	}

	// Check if ABI was actually returned
	if contractInfo == nil || contractInfo.ABI == "" {
		// Use correct NewABIError signature: NewABIError(err error, context string)
		// Since there's no underlying technical error here, just missing data, we pass nil for err.
		return nil, errors.NewABIError(nil, fmt.Sprintf("ABI not found for contract %s via block explorer", contractAddress))
	}

	// Parse the fetched ABI string
	parsedABI, err := abi.JSON(strings.NewReader(contractInfo.ABI))
	if err != nil {
		// Use correct NewABIError signature
		return nil, errors.NewABIError(err, fmt.Sprintf("failed to parse ABI fetched from explorer for %s", contractAddress))
	}

	return &parsedABI, nil
}
