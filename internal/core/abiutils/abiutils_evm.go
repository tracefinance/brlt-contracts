package abiutils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/contract"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// evmAbiUtils implements the ABIUtils interface for EVM chains.
type evmAbiUtils struct {
	config           *config.Config
	explorer         blockexplorer.BlockExplorer // Store the specific explorer instance
	blockchainClient blockchain.BlockchainClient
	log              logger.Logger
	chainType        types.ChainType // Store chain type for context
	abiCache         *sync.Map       // Cache: string (address or name) -> string (ABI JSON)
}

// NewEvmAbiUtils creates a new EVM ABI utility instance for a specific chain.
func NewEvmAbiUtils(chainType types.ChainType, cfg *config.Config, log logger.Logger, explorer blockexplorer.BlockExplorer, blockchainClient blockchain.BlockchainClient) (ABIUtils, error) {
	return &evmAbiUtils{
		config:           cfg,
		explorer:         explorer, // Store the created explorer
		log:              log,
		blockchainClient: blockchainClient,
		chainType:        chainType, // Store the chain type
		abiCache:         &sync.Map{},
	}, nil
}

// LoadABIByName loads a known ABI type using the configured mapping to find the artifact file.
func (u *evmAbiUtils) LoadABIByName(ctx context.Context, abiType SupportedABIType) (string, error) {
	cacheKey := string(abiType)

	// 1. Check cache
	if cached, found := u.abiCache.Load(cacheKey); found {
		if abiStr, ok := cached.(string); ok {
			return abiStr, nil
		}
	}

	// 2. Get contract name from config mapping
	contractName, err := u.config.GetABIContractNameForType(string(abiType))
	if err != nil {
		// Wrap the config error for clarity
		return "", err
	}

	// 3. Load artifact using the contract name
	artifact, err := u.loadArtifact(contractName)
	if err != nil {
		return "", err // Propagate error from artifact loading
	}

	// 4. Store ABI string in cache and return
	u.abiCache.Store(cacheKey, artifact.ABI)
	return artifact.ABI, nil
}

// loadArtifact is a helper adapted from evm_manager.go
// Uses contract.Artifact from the contract package.
func (u *evmAbiUtils) loadArtifact(contractName string) (*contract.Artifact, error) {
	contractsPath := u.config.GetSmartContractsPath()

	// Get the base name by removing any extension
	ext := filepath.Ext(contractName)
	baseName := strings.TrimSuffix(contractName, ext)

	// Construct the expected artifact path: {contractsPath}/{contractName}/{baseName}.json
	artifactPath := filepath.Join(contractsPath, contractName, baseName+".json")

	// Check if the file exists at the primary expected path
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		// If not found, return an error indicating the expected path
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("artifact file not found at expected path: %s", artifactPath))
	}

	// Read the artifact file
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to read artifact file %s: %w", artifactPath, err))
	}

	var artifactData map[string]any
	if err := json.Unmarshal(data, &artifactData); err != nil {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to parse artifact JSON %s: %w", artifactPath, err))
	}

	abiData, ok := artifactData["abi"]
	if !ok {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("abi not found in artifact file %s", artifactPath))
	}
	abiJSON, err := json.Marshal(abiData)
	if err != nil {
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to marshal ABI from artifact file %s: %w", artifactPath, err))
	}

	return &contract.Artifact{
		Name: baseName, // Use baseName for the Artifact struct as well
		ABI:  string(abiJSON),
	}, nil
}

// LoadABIByAddress loads ABI from the configured block explorer or local cache.
// It attempts to resolve proxy contracts by looking for an 'implementation()' method.
// If a proxy is detected and its implementation ABI can be fetched, that ABI is
// cached under the proxy's address and returned.
func (u *evmAbiUtils) LoadABIByAddress(ctx context.Context, contractAddress types.Address) (string, error) {
	cacheKey := contractAddress.String()

	// 1. Check cache first (this covers direct hits and previously resolved proxy implementations)
	if cached, found := u.abiCache.Load(cacheKey); found {
		if abiStr, ok := cached.(string); ok {
			return abiStr, nil
		}
	}

	// 2. Fetch the ABI for the given contractAddress. This might be a proxy or a direct contract.
	// fetchAndCacheABIForAddress will also cache this ABI under contractAddress.String().
	proxyOrDirectABIString, err := u.fetchAndCacheABIForAddress(ctx, contractAddress)
	if err != nil {
		return "", err // Propagate errors from fetching the initial ABI (e.g., contract not found)
	}

	// 3. Try to get the implementation address from the potentially proxy contract.
	implementationAddress, proxyErr := u.getImplementationAddressFromProxy(ctx, contractAddress, proxyOrDirectABIString)

	// 4. Handle outcomes of proxy resolution
	if proxyErr != nil {
		// Check if the error is because the 'implementation' method was not found.
		// This indicates it's likely not a proxy of the type we're handling, or it's a direct contract.
		if errors.IsError(proxyErr, errors.ErrCodeABIProxyMethodNotFound) || errors.IsError(proxyErr, errors.ErrCodeABIProxyMethodSignatureInvalid) {
			// Not an 'implementation()' proxy, or method signature is wrong.
			// The proxyOrDirectABIString is the correct one.
			// It's already cached by fetchAndCacheABIForAddress, so just return it.
			return proxyOrDirectABIString, nil
		}

		// For other errors during proxy resolution (e.g., blockchain call failure to get implementation address),
		// log the error and fall back to the proxy's direct ABI.
		u.log.Warn("Error resolving proxy implementation, falling back to direct ABI",
			logger.String("contract_address", contractAddress.String()),
			logger.String("chain_type", string(u.chainType)),
			logger.Error(proxyErr),
		)
		// The proxyOrDirectABIString is already cached, so we return it.
		return proxyOrDirectABIString, nil
	}

	// 5. If implementationAddress was successfully retrieved (proxyErr is nil and implementationAddress is not nil)
	if implementationAddress == nil { // Should not happen if proxyErr is nil, but good for robustness
		// This case implies an unexpected nil address with no error from getImplementationAddressFromProxy.
		// Fallback to direct ABI
		u.log.Warn("Proxy resolution resulted in nil implementation address without error, falling back to direct ABI",
			logger.String("contract_address", contractAddress.String()),
		)
		return proxyOrDirectABIString, nil
	}

	// Log successful proxy identification
	u.log.Info("Proxy contract identified, proceeding to fetch implementation ABI",
		logger.String("proxy_address", contractAddress.String()),
		logger.String("implementation_address", implementationAddress.String()),
		logger.String("chain_type", string(u.chainType)),
	)

	// Now, fetch the ABI for the implementation contract.
	implementationABIString, implErr := u.fetchAndCacheABIForAddress(ctx, *implementationAddress)
	if implErr != nil {
		// Failed to fetch the ABI for the implementation address.
		// Log the error and fall back to the proxy's direct ABI.
		u.log.Warn("Error fetching implementation ABI, falling back to proxy's direct ABI",
			logger.String("proxy_address", contractAddress.String()),
			logger.String("implementation_address", implementationAddress.String()),
			logger.String("chain_type", string(u.chainType)),
			logger.Error(implErr),
		)
		// The proxyOrDirectABIString is already cached by the first call to fetchAndCacheABIForAddress.
		return proxyOrDirectABIString, nil
	}

	// 6. Successfully fetched the implementation ABI.
	// Cache the implementation ABI under the *proxy's address* for subsequent calls.
	// This fulfills task 4 requirement.
	u.abiCache.Store(cacheKey, implementationABIString) // cacheKey is contractAddress.String()
	return implementationABIString, nil
}

// ParseContractInput parses the input data of a transaction using the provided ABI and method name.
// If methodName is provided, it will be used to find the method in the ABI.
func (u *evmAbiUtils) ParseContractInput(contractABI string, methodName string, inputData []byte) (map[string]interface{}, error) {
	// Parse the ABI string
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, errors.NewABIParseError(err)
	}

	// Prepare variables
	var method abi.Method
	var dataToUnpack []byte
	var methodID []byte

	// Extract method selector if available
	// Assumes inputData ALWAYS includes the selector
	if len(inputData) < 4 {
		return nil, errors.NewABIInputDataTooShortError(len(inputData), 4)
	}
	methodID = inputData[:4]
	dataToUnpack = inputData[4:]

	// Case 1: Find method by name
	if methodName != "" {
		var ok bool
		method, ok = parsedABI.Methods[methodName]
		if !ok {
			return nil, errors.NewABIMethodNotFoundError(methodName, true)
		}

		// Verify the selector matches the method name
		if !bytes.Equal(method.ID, methodID) {
			return nil, errors.NewABIMethodSelectorMismatchError(methodName, method.ID, methodID)
		}

		// Found method by name successfully
	} else {
		// Case 2: Find method by selector (no name provided)
		found := false
		for _, m := range parsedABI.Methods {
			if bytes.Equal(m.ID, methodID) {
				method = m
				found = true
				break
			}
		}

		if !found {
			return nil, errors.NewABIMethodNotFoundError(fmt.Sprintf("%x", methodID), false)
		}

		// Found method by selector successfully
	}

	// Validate input data
	if len(method.Inputs) > 0 && len(dataToUnpack) == 0 {
		return nil, errors.NewABIInputDataEmptyError(method.Name)
	}

	if len(dataToUnpack) > 0 && len(dataToUnpack)%32 != 0 {
		return nil, errors.NewABIInputDataInvalidLengthError(method.Name, len(dataToUnpack))
	}

	// Unpack the arguments
	unpackedArgs := make(map[string]any)
	err = method.Inputs.UnpackIntoMap(unpackedArgs, dataToUnpack)
	if err != nil {
		return nil, errors.NewABIUnpackFailedError(err, method.Name)
	}

	return unpackedArgs, nil
}

// ExtractMethodID extracts the 4-byte method selector from transaction data.
// Returns nil if data is shorter than 4 bytes.
func (u *evmAbiUtils) ExtractMethodID(data []byte) []byte {
	if len(data) < 4 {
		return nil
	}
	return data[:4]
}

// Pack packs arguments for a method call.
func (u *evmAbiUtils) Pack(contractABI string, name string, args ...interface{}) ([]byte, error) {
	if contractABI == "" {
		return nil, errors.NewInvalidParameterError("contract ABI is empty", "contractABI")
	}

	// Parse the ABI string
	parsedABI, err := abi.JSON(strings.NewReader(contractABI))
	if err != nil {
		return nil, errors.NewABIParseError(err)
	}

	data, err := parsedABI.Pack(name, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no method with id") || strings.Contains(err.Error(), "method '"+name+"' not found") {
			return nil, errors.NewABIMethodNotFoundError(name, true)
		}
		return nil, errors.NewABIPackFailedError(err, name)
	}
	return data, nil
}

// GetAddressFromArgs helper implementation
func (u *evmAbiUtils) GetAddressFromArgs(args map[string]any, key string) (types.Address, error) {
	val, ok := args[key]
	if !ok {
		return types.Address{}, errors.NewABIArgumentNotFoundError(key)
	}

	// Check if it's already a types.Address
	if addr, ok := val.(types.Address); ok {
		return addr, nil
	}

	// Check if it's a common.Address (expected case with go-ethereum)
	if ethAddr, ok := val.(common.Address); ok {
		// Convert common.Address to types.Address
		addr, err := types.NewAddress(u.chainType, ethAddr.Hex())
		if err != nil {
			return types.Address{}, errors.NewABIArgumentConversionError(err, key, "types.Address from common.Address", ethAddr.Hex())
		}
		return *addr, nil
	}

	// As a fallback, try to handle it as a string address
	if addrStr, ok := val.(string); ok {
		addr, err := types.NewAddress(u.chainType, addrStr)
		if err != nil {
			return types.Address{}, errors.NewABIArgumentConversionError(err, key, "types.Address from string", addrStr)
		}
		return *addr, nil
	}

	return types.Address{}, errors.NewABIArgumentInvalidTypeError(key, "types.Address, common.Address or string", fmt.Sprintf("%T", val))
}

// GetBytes32FromArgs helper implementation
func (u *evmAbiUtils) GetBytes32FromArgs(args map[string]any, key string) ([32]byte, error) {
	val, ok := args[key]
	if !ok {
		return [32]byte{}, errors.NewABIArgumentNotFoundError(key)
	}
	bytes32Val, ok := val.([32]byte)
	if !ok {
		byteSlice, sliceOk := val.([]byte)
		if sliceOk && len(byteSlice) == 32 {
			copy(bytes32Val[:], byteSlice)
			ok = true
		} else {
			return [32]byte{}, errors.NewABIArgumentInvalidTypeError(key, "[32]byte or []byte with length 32", fmt.Sprintf("%T (length %d)", val, len(byteSlice)))
		}
	}
	return bytes32Val, nil
}

// GetBigIntFromArgs helper implementation
func (u *evmAbiUtils) GetBigIntFromArgs(args map[string]any, key string) (*types.BigInt, error) {
	val, ok := args[key]
	if !ok {
		return nil, errors.NewABIArgumentNotFoundError(key)
	}
	typesBigIntValue, ok := val.(types.BigInt)
	if ok {
		newVal := types.NewBigInt(typesBigIntValue.Int)
		return &newVal, nil
	}
	typesBigIntPtrValue, ok := val.(*types.BigInt)
	if ok {
		if typesBigIntPtrValue == nil || typesBigIntPtrValue.Int == nil {
			return nil, errors.NewABIArgumentNilValueError(key)
		}
		newVal := types.NewBigInt(typesBigIntPtrValue.Int)
		return &newVal, nil
	}
	goBigInt, goOk := val.(*big.Int)
	if goOk {
		if goBigInt == nil {
			return nil, errors.NewABIArgumentNilValueError(key)
		}
		newVal := types.NewBigInt(goBigInt)
		return &newVal, nil
	}
	return nil, errors.NewABIArgumentInvalidTypeError(key, "types.BigInt, *types.BigInt or *big.Int", fmt.Sprintf("%T", val))
}

// GetUint64FromArgs helper implementation
func (u *evmAbiUtils) GetUint64FromArgs(args map[string]any, key string) (uint64, error) {
	val, ok := args[key]
	if !ok {
		return 0, errors.NewABIArgumentNotFoundError(key)
	}
	uintVal, ok := val.(uint64)
	if !ok {
		bigIntVal, bigOk := val.(*big.Int)
		if bigOk && bigIntVal.IsUint64() {
			uintVal = bigIntVal.Uint64()
			ok = true
		} else {
			return 0, errors.NewABIArgumentInvalidTypeError(key, "uint64 or *big.Int representable as uint64", fmt.Sprintf("%T", val))
		}
	}
	return uintVal, nil
}

// fetchAndCacheABIForAddress retrieves ABI for a given address, utilizing a cache.
// It first checks the cache. If not found, it fetches from the block explorer,
// then stores it in the cache before returning.
func (u *evmAbiUtils) fetchAndCacheABIForAddress(ctx context.Context, addressToFetch types.Address) (string, error) {
	cacheKey := addressToFetch.String()

	// 1. Check cache
	if cached, found := u.abiCache.Load(cacheKey); found {
		if abiStr, ok := cached.(string); ok {
			return abiStr, nil
		}
	}

	// 2. Fetch from the stored explorer instance
	contractInfo, err := u.explorer.GetContract(ctx, addressToFetch.String())
	if err != nil {
		if errors.IsError(err, errors.ErrCodeContractNotFound) {
			return "", errors.NewContractNotFoundError(addressToFetch.String(), string(u.chainType))
		}
		// Wrap the explorer request error, including chain type
		return "", errors.NewExplorerRequestFailedError(fmt.Errorf("failed to get contract info for ABI fetch for %s on chain %s: %w", addressToFetch.String(), u.chainType, err))
	}

	// 3. Check if ABI is available
	if contractInfo.ABI == "" {
		return "", errors.NewABIUnavailableOrUnverifiedError(addressToFetch.String(), string(u.chainType))
	}

	// 4. Store ABI string in cache and return
	u.abiCache.Store(cacheKey, contractInfo.ABI)
	return contractInfo.ABI, nil
}

// getImplementationAddressFromProxy attempts to read the implementation address
// from a proxy contract that follows the "implementation() view returns (address)" pattern.
func (u *evmAbiUtils) getImplementationAddressFromProxy(ctx context.Context, proxyAddress types.Address, proxyABIString string) (*types.Address, error) {
	// 1. Parse Proxy ABI
	parsedProxyABI, err := abi.JSON(strings.NewReader(proxyABIString))
	if err != nil {
		return nil, errors.NewABIProxyParseError(err, proxyAddress.String())
	}

	// 2. Define Method Name
	methodName := "implementation"

	// 3. Retrieve Method
	method, ok := parsedProxyABI.Methods[methodName]

	// 4. Validate method signature
	if !ok {
		// Method not found, this might be an EIP-1967 proxy or not a proxy at all.
		// Return a specific error code that LoadABIByAddress can check for.
		return nil, errors.NewABIProxyMethodNotFoundError(proxyAddress.String(), methodName)
	}

	// Further validation for "implementation() view returns (address)"
	// Expected: 0 inputs, 1 output of type address.
	// Solidity `view` functions are `Constant` in go-ethereum ABI.
	if !(len(method.Inputs) == 0 && len(method.Outputs) == 1 && method.Outputs[0].Type.T == abi.AddressTy) {
		return nil, errors.NewABIProxyMethodSignatureInvalidError(proxyAddress.String(), methodName, "implementation() view returns (address)", method.Sig)
	}

	// 5. Pack Call Data
	packedCallData, err := parsedProxyABI.Pack(methodName) // No arguments for implementation()
	if err != nil {
		// This should be unlikely if method was found and ABI parsed correctly.
		return nil, errors.NewABIProxyPackError(err, proxyAddress.String(), methodName)
	}

	// 6. Call Blockchain
	// For a view call, 'from' can be the proxyAddress itself or a zero address.
	resultData, err := u.blockchainClient.CallContract(ctx, proxyAddress.String(), proxyAddress.String(), packedCallData)
	if err != nil {
		return nil, errors.NewABIProxyCallError(err, proxyAddress.String(), methodName, string(u.chainType))
	}

	// 7. Check Result Data
	if len(resultData) == 0 {
		return nil, errors.NewABIProxyEmptyResultError(proxyAddress.String(), methodName)
	}

	// 8. Unpack Result
	var outputs []any // A slice to hold all output arguments
	outputs, err = method.Outputs.Unpack(resultData)
	if err != nil {
		return nil, errors.NewABIProxyUnpackError(err, proxyAddress.String(), methodName, "unpacking failed")
	}

	// Check if we got at least one output
	if len(outputs) == 0 || outputs[0] == nil {
		return nil, errors.NewABIProxyUnpackError(nil, proxyAddress.String(), methodName, "unpacked result is empty or nil")
	}

	// The first output should be the address
	implEthAddr, ok := outputs[0].(common.Address)
	if !ok {
		return nil, errors.NewABIProxyUnpackError(nil, proxyAddress.String(), methodName, fmt.Sprintf("unpacked result is not common.Address, got %T", outputs[0]))
	}

	// 9. Convert common.Address to types.Address
	implementationAddress, err := types.NewAddress(u.chainType, implEthAddr.Hex())
	if err != nil {
		return nil, errors.NewABIProxyAddressConversionError(err, implEthAddr.Hex(), string(u.chainType))
	}

	// 10. Return
	return implementationAddress, nil
}
