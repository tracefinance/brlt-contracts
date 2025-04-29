package transaction

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
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/contract"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// evmAbiUtils implements the ABIUtils interface for EVM chains.
type evmAbiUtils struct {
	config    *config.Config
	explorer  blockexplorer.BlockExplorer // Store the specific explorer instance
	chainType types.ChainType             // Store chain type for context
	abiCache  *sync.Map                   // Cache: string (address or name) -> *abi.ABI
}

// NewEvmAbiUtils creates a new EVM ABI utility instance for a specific chain.
func NewEvmAbiUtils(chainType types.ChainType, cfg *config.Config, explorerFactory blockexplorer.Factory) (ABIUtils, error) {
	// Use the factory to get the explorer for this specific chain
	explorer, err := explorerFactory.NewExplorer(chainType)
	if err != nil {
		// Wrap the error for clarity
		return nil, fmt.Errorf("failed to create block explorer for chain %s in ABIUtils: %w", chainType, err)
	}

	return &evmAbiUtils{
		config:    cfg,
		explorer:  explorer,  // Store the created explorer
		chainType: chainType, // Store the chain type
		abiCache:  &sync.Map{},
	}, nil
}

// LoadABIByName loads a known ABI type using the configured mapping to find the artifact file.
func (u *evmAbiUtils) LoadABIByName(ctx context.Context, abiType SupportedABIType) (*abi.ABI, error) {
	cacheKey := string(abiType)

	// 1. Check cache
	if cached, found := u.abiCache.Load(cacheKey); found {
		if abiPtr, ok := cached.(*abi.ABI); ok {
			return abiPtr, nil
		}
	}

	// 2. Get contract name from config mapping
	contractName, err := u.config.GetABIContractNameForType(string(abiType))
	if err != nil {
		// Wrap the config error for clarity
		return nil, err
	}

	// 3. Load artifact using the contract name
	artifact, err := u.loadArtifact(ctx, contractName)
	if err != nil {
		return nil, err // Propagate error from artifact loading
	}

	// 4. Parse ABI from artifact
	parsedABI, err := abi.JSON(strings.NewReader(artifact.ABI))
	if err != nil {
		// Include contract name from mapping in the error
		return nil, errors.NewInvalidContractError(contractName, fmt.Errorf("failed to parse ABI from artifact for type '%s' (contract %s): %w", abiType, contractName, err))
	}

	// 5. Store pointer in cache and return
	u.abiCache.Store(cacheKey, &parsedABI)
	return &parsedABI, nil
}

// loadArtifact is a helper adapted from evm_manager.go
// Uses contract.Artifact from the contract package.
func (u *evmAbiUtils) loadArtifact(ctx context.Context, contractName string) (*contract.Artifact, error) {
	contractsPath := u.config.GetSmartContractsPath()

	// Get the base name by removing any extension
	ext := filepath.Ext(contractName)
	baseName := strings.TrimSuffix(contractName, ext)

	// Construct the expected artifact path: {contractsPath}/{baseName}/{baseName}.json
	artifactPath := filepath.Join(contractsPath, baseName, baseName+".json")

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

// LoadABIByAddress loads ABI from the configured block explorer, caching the result.
func (u *evmAbiUtils) LoadABIByAddress(ctx context.Context, contractAddress common.Address) (*abi.ABI, error) {
	cacheKey := contractAddress.Hex()

	// 1. Check cache
	if cached, found := u.abiCache.Load(cacheKey); found {
		if abiPtr, ok := cached.(*abi.ABI); ok {
			return abiPtr, nil
		}
	}

	// 2. Fetch from the stored explorer instance
	if u.explorer == nil {
		// This should not happen if constructor succeeded
		return nil, fmt.Errorf("ABIUtils explorer is not initialized for chain %s", u.chainType)
	}
	contractInfo, err := u.explorer.GetContract(ctx, contractAddress.Hex())
	if err != nil {
		if errors.IsError(err, errors.ErrCodeContractNotFound) {
			// Use the stored chainType in the error message
			return nil, errors.NewContractNotFoundError(contractAddress.Hex(), string(u.chainType))
		}
		// Wrap the explorer request error, including chain type
		return nil, errors.NewExplorerRequestFailedError(fmt.Errorf("failed to get contract info for ABI fetch %s on chain %s: %w", contractAddress.Hex(), u.chainType, err))
	}

	// 3. Check if ABI is available
	if contractInfo.ABI == "" {
		return nil, errors.NewInvalidContractError(contractAddress.Hex(), fmt.Errorf("ABI not found via explorer or contract not verified for %s on chain %s", contractAddress.Hex(), u.chainType))
	}

	// 4. Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(contractInfo.ABI))
	if err != nil {
		return nil, errors.NewInvalidContractError(contractAddress.Hex(), fmt.Errorf("failed to parse ABI fetched from explorer for %s on chain %s: %w", contractAddress.Hex(), u.chainType, err))
	}

	// 5. Store pointer in cache and return
	u.abiCache.Store(cacheKey, &parsedABI)
	return &parsedABI, nil
}

// GetMethodFromABI finds the ABI method corresponding to the given ID.
func (u *evmAbiUtils) GetMethodFromABI(contractABI *abi.ABI, methodID []byte) (*abi.Method, error) {
	if contractABI == nil {
		return nil, errors.NewInvalidParameterError("contract ABI is nil", "contractABI")
	}
	if len(methodID) != 4 {
		return nil, errors.NewInvalidParameterError("method ID must be 4 bytes", "methodID")
	}

	for _, m := range contractABI.Methods {
		if bytes.Equal(m.ID, methodID) {
			return &m, nil
		}
	}
	return nil, errors.NewABIError(fmt.Errorf("method with ID %x not found in ABI", methodID), "method_lookup")
}

// ParseContractInput parses the input data of a transaction using the provided ABI method.
// inputData should NOT include the 4-byte selector.
func (u *evmAbiUtils) ParseContractInput(method *abi.Method, inputData []byte) (map[string]interface{}, error) {
	if method == nil {
		return nil, errors.NewInvalidParameterError("ABI method is nil", "method")
	}

	// Validate input length against expected parameters
	if len(method.Inputs) > 0 && len(inputData) == 0 {
		return nil, errors.NewABIError(fmt.Errorf("input data is empty for method %s requiring arguments", method.Name), "empty_input")
	}
	if len(inputData)%32 != 0 {
		return nil, errors.NewABIError(fmt.Errorf("input data length %d is not a multiple of 32 bytes for method %s", len(inputData), method.Name), "invalid_length")
	}

	unpackedArgs := make(map[string]any)
	err := method.Inputs.UnpackIntoMap(unpackedArgs, inputData)
	if err != nil {
		return nil, errors.NewABIError(fmt.Errorf("failed to unpack input for method %s: %w", method.Name, err), "unpacking_failed")
	}

	return unpackedArgs, nil
}

// ExtractMethodID extracts the 4-byte method selector from transaction data.
func (u *evmAbiUtils) ExtractMethodID(data []byte) []byte {
	if len(data) < 4 {
		return nil
	}
	return data[:4]
}

// Pack packs arguments for a method call.
func (u *evmAbiUtils) Pack(contractABI *abi.ABI, name string, args ...interface{}) ([]byte, error) {
	if contractABI == nil {
		return nil, errors.NewInvalidParameterError("contract ABI is nil", "contractABI")
	}
	data, err := contractABI.Pack(name, args...)
	if err != nil {
		if strings.Contains(err.Error(), "no method with id") || strings.Contains(err.Error(), "method '"+name+"' not found") {
			return nil, errors.NewABIError(fmt.Errorf("method '%s' not found in provided ABI for packing", name), "pack_method_not_found")
		}
		return nil, errors.NewABIError(fmt.Errorf("failed to pack arguments for method '%s': %w", name, err), "packing_failed")
	}
	return data, nil
}

// GetAddressFromArgs helper implementation
func (u *evmAbiUtils) GetAddressFromArgs(args map[string]interface{}, key string) (common.Address, error) {
	val, ok := args[key]
	if !ok {
		return common.Address{}, errors.NewABIError(fmt.Errorf("argument '%s' not found", key), "missing_arg")
	}
	addr, ok := val.(common.Address)
	if !ok {
		return common.Address{}, errors.NewABIError(fmt.Errorf("argument '%s' is not a valid address type (%T)", key, val), "invalid_arg_type")
	}
	return addr, nil
}

// GetBytes32FromArgs helper implementation
func (u *evmAbiUtils) GetBytes32FromArgs(args map[string]interface{}, key string) ([32]byte, error) {
	val, ok := args[key]
	if !ok {
		return [32]byte{}, errors.NewABIError(fmt.Errorf("argument '%s' not found", key), "missing_arg")
	}
	bytes32Val, ok := val.([32]byte)
	if !ok {
		byteSlice, sliceOk := val.([]byte)
		if sliceOk && len(byteSlice) == 32 {
			copy(bytes32Val[:], byteSlice)
			ok = true
		} else {
			return [32]byte{}, errors.NewABIError(fmt.Errorf("argument '%s' is not a valid [32]byte type (%T)", key, val), "invalid_arg_type")
		}
	}
	return bytes32Val, nil
}

// GetBigIntFromArgs helper implementation
func (u *evmAbiUtils) GetBigIntFromArgs(args map[string]interface{}, key string) (*types.BigInt, error) {
	val, ok := args[key]
	if !ok {
		return nil, errors.NewABIError(fmt.Errorf("argument '%s' not found", key), "missing_arg")
	}
	typesBigIntValue, ok := val.(types.BigInt)
	if ok {
		newVal := types.NewBigInt(typesBigIntValue.Int)
		return &newVal, nil
	}
	typesBigIntPtrValue, ok := val.(*types.BigInt)
	if ok {
		if typesBigIntPtrValue == nil || typesBigIntPtrValue.Int == nil {
			return nil, errors.NewABIError(fmt.Errorf("argument '%s' resolved to a nil *types.BigInt or embedded nil *big.Int", key), "nil_value")
		}
		newVal := types.NewBigInt(typesBigIntPtrValue.Int)
		return &newVal, nil
	}
	goBigInt, goOk := val.(*big.Int)
	if goOk {
		if goBigInt == nil {
			return nil, errors.NewABIError(fmt.Errorf("argument '%s' resolved to a nil *big.Int", key), "nil_value")
		}
		newVal := types.NewBigInt(goBigInt)
		return &newVal, nil
	}
	return nil, errors.NewABIError(fmt.Errorf("argument '%s' is not a valid big integer type (%T)", key, val), "invalid_arg_type")
}

// GetUint64FromArgs helper implementation
func (u *evmAbiUtils) GetUint64FromArgs(args map[string]interface{}, key string) (uint64, error) {
	val, ok := args[key]
	if !ok {
		return 0, errors.NewABIError(fmt.Errorf("argument '%s' not found", key), "missing_arg")
	}
	uintVal, ok := val.(uint64)
	if !ok {
		bigIntVal, bigOk := val.(*big.Int)
		if bigOk && bigIntVal.IsUint64() {
			uintVal = bigIntVal.Uint64()
			ok = true
		} else {
			return 0, errors.NewABIError(fmt.Errorf("argument '%s' is not a valid uint64 type (%T)", key, val), "invalid_arg_type")
		}
	}
	return uintVal, nil
}
