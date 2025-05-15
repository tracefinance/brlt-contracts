package abi

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// SupportedABIType defines the types of ABIs that can be explicitly loaded by name.
// Currently limited to ERC20 and MultiSig.
type SupportedABIType string

const (
	ABITypeERC20    SupportedABIType = "erc20"
	ABITypeMultiSig SupportedABIType = "multisig"
)

// ABILoader is responsible for loading ABI from files and addresses.
type ABILoader interface {
	// LoadABIByType loads a known ABI type using the configured mapping to find the artifact file.
	LoadABIByType(ctx context.Context, abiType SupportedABIType) (string, error)

	// LoadABIByAddress loads ABI from the configured block explorer or local cache.
	// It attempts to resolve proxy contracts by looking for an 'implementation()' method.
	LoadABIByAddress(ctx context.Context, contractAddress types.Address) (string, error)
}

// abiLoader is responsible for loading ABI from files and addresses.
type abiLoader struct {
	config           *config.Config
	explorer         blockexplorer.BlockExplorer
	blockchainClient blockchain.BlockchainClient
	log              logger.Logger
	chainType        types.ChainType
	abiCache         *sync.Map // Cache: string (address or name) -> string (ABI JSON)
	abiUtils         ABIUtils
}

// NewABILoader creates a new ABILoader instance.
func NewABILoader(
	chainType types.ChainType,
	cfg *config.Config,
	explorer blockexplorer.BlockExplorer,
	blockchainClient blockchain.BlockchainClient,
	abiUtils ABIUtils,
	log logger.Logger,
) ABILoader {
	return &abiLoader{
		config:           cfg,
		explorer:         explorer,
		blockchainClient: blockchainClient,
		log:              log,
		chainType:        chainType,
		abiCache:         &sync.Map{},
		abiUtils:         abiUtils,
	}
}

// LoadABIFromFile has been removed and its functionality incorporated into LoadABIByName

// LoadABIByType loads a known ABI type using the configured mapping to find the artifact file.
func (l *abiLoader) LoadABIByType(ctx context.Context, abiType SupportedABIType) (string, error) {
	cacheKey := string(abiType)

	// Check cache
	if cached, found := l.abiCache.Load(cacheKey); found {
		if abiStr, ok := cached.(string); ok {
			return abiStr, nil
		}
	}

	// Get artifact path from config mapping
	artifactPath, err := l.config.GetArtifactPathForType(string(abiType))
	if err != nil {
		return "", err
	}

	// Check if the file exists at the configured path
	if _, err := os.Stat(artifactPath); os.IsNotExist(err) {
		return "", errors.NewInvalidContractError(string(abiType), fmt.Errorf("artifact file not found at path: %s", artifactPath))
	}

	// Read the artifact file
	data, err := os.ReadFile(artifactPath)
	if err != nil {
		return "", errors.NewInvalidContractError(string(abiType), fmt.Errorf("failed to read artifact file %s: %w", artifactPath, err))
	}

	// Parse the artifact JSON - supporting both Hardhat and Truffle formats
	var artifactData map[string]any
	if err := json.Unmarshal(data, &artifactData); err != nil {
		return "", errors.NewInvalidContractError(string(abiType), fmt.Errorf("failed to parse artifact JSON %s: %w", artifactPath, err))
	}

	// Extract the ABI
	abiData, ok := artifactData["abi"]
	if !ok {
		return "", errors.NewInvalidContractError(string(abiType), fmt.Errorf("abi not found in artifact file %s", artifactPath))
	}

	abiJSON, err := json.Marshal(abiData)
	if err != nil {
		return "", errors.NewInvalidContractError(string(abiType), fmt.Errorf("failed to marshal ABI from artifact file %s: %w", artifactPath, err))
	}

	// Store in cache
	abiString := string(abiJSON)
	l.abiCache.Store(cacheKey, abiString)

	return abiString, nil
}

// LoadABIByAddress loads ABI from the configured block explorer or local cache.
// It attempts to resolve proxy contracts by looking for an 'implementation()' method.
func (l *abiLoader) LoadABIByAddress(ctx context.Context, contractAddress types.Address) (string, error) {
	cacheKey := contractAddress.String()

	// 1. Check cache first
	if cached, found := l.abiCache.Load(cacheKey); found {
		if abiStr, ok := cached.(string); ok {
			return abiStr, nil
		}
	}

	// 2. Fetch the ABI for the given contractAddress
	proxyOrDirectABIString, err := l.fetchAndCacheABIForAddress(ctx, contractAddress)
	if err != nil {
		return "", err
	}

	// 3. Try to get the implementation address from the potentially proxy contract
	implementationAddress, proxyErr := l.getImplementationAddressFromProxy(ctx, contractAddress, proxyOrDirectABIString)

	// 4. Handle outcomes of proxy resolution
	if proxyErr != nil {
		// Check if the error is because the 'implementation' method was not found
		if errors.IsError(proxyErr, errors.ErrCodeABIProxyMethodNotFound) ||
			errors.IsError(proxyErr, errors.ErrCodeABIProxyMethodSignatureInvalid) {
			// Not a proxy or method signature is wrong - use direct ABI
			return proxyOrDirectABIString, nil
		}

		// For other errors, log and fall back to direct ABI
		l.log.Warn("Error resolving proxy implementation, falling back to direct ABI",
			logger.String("contract_address", contractAddress.String()),
			logger.String("chain_type", string(l.chainType)),
			logger.Error(proxyErr),
		)
		return proxyOrDirectABIString, nil
	}

	// 5. If implementationAddress was successfully retrieved
	if implementationAddress == nil {
		l.log.Warn("Proxy resolution resulted in nil implementation address without error, falling back to direct ABI",
			logger.String("contract_address", contractAddress.String()),
		)
		return proxyOrDirectABIString, nil
	}

	// Log successful proxy identification
	l.log.Info("Proxy contract identified, proceeding to fetch implementation ABI",
		logger.String("proxy_address", contractAddress.String()),
		logger.String("implementation_address", implementationAddress.String()),
		logger.String("chain_type", string(l.chainType)),
	)

	// Fetch the ABI for the implementation contract
	implementationABIString, implErr := l.fetchAndCacheABIForAddress(ctx, *implementationAddress)
	if implErr != nil {
		// Failed to fetch implementation ABI, fall back to proxy's direct ABI
		l.log.Warn("Error fetching implementation ABI, falling back to proxy's direct ABI",
			logger.String("proxy_address", contractAddress.String()),
			logger.String("implementation_address", implementationAddress.String()),
			logger.String("chain_type", string(l.chainType)),
			logger.Error(implErr),
		)
		return proxyOrDirectABIString, nil
	}

	// 6. Cache the implementation ABI under the proxy's address for subsequent calls
	l.abiCache.Store(cacheKey, implementationABIString)
	return implementationABIString, nil
}

// fetchAndCacheABIForAddress retrieves ABI for a given address, utilizing cache
func (l *abiLoader) fetchAndCacheABIForAddress(ctx context.Context, addressToFetch types.Address) (string, error) {
	cacheKey := addressToFetch.String()

	// 1. Check cache
	if cached, found := l.abiCache.Load(cacheKey); found {
		if abiStr, ok := cached.(string); ok {
			return abiStr, nil
		}
	}

	// 2. Fetch from block explorer
	contractInfo, err := l.explorer.GetContract(ctx, addressToFetch.String())
	if err != nil {
		if errors.IsError(err, errors.ErrCodeContractNotFound) {
			return "", errors.NewContractNotFoundError(addressToFetch.String(), string(l.chainType))
		}
		return "", errors.NewExplorerRequestFailedError(
			fmt.Errorf("failed to get contract info for ABI fetch for %s on chain %s: %w",
				addressToFetch.String(), l.chainType, err),
		)
	}

	// 3. Check if ABI is available
	if contractInfo.ABI == "" {
		return "", errors.NewABIUnavailableOrUnverifiedError(addressToFetch.String(), string(l.chainType))
	}

	// 4. Store ABI in cache and return
	l.abiCache.Store(cacheKey, contractInfo.ABI)
	return contractInfo.ABI, nil
}

// getImplementationAddressFromProxy attempts to read the implementation address
// from a proxy contract that follows the "implementation() view returns (address)" pattern
func (l *abiLoader) getImplementationAddressFromProxy(
	ctx context.Context,
	proxyAddress types.Address,
	proxyABIString string,
) (*types.Address, error) {
	// 1. Pack the calldata for the implementation() method
	packedCallData, err := l.abiUtils.Pack(proxyABIString, "implementation")
	if err != nil {
		if errors.IsError(err, errors.ErrCodeABIMethodNotFound) {
			return nil, errors.NewABIProxyMethodNotFoundError(proxyAddress.String(), "implementation")
		}
		return nil, errors.NewABIProxyPackError(err, proxyAddress.String(), "implementation")
	}

	// 2. Call blockchain
	resultData, err := l.blockchainClient.CallContract(ctx, proxyAddress.String(), proxyAddress.String(), packedCallData)
	if err != nil {
		return nil, errors.NewABIProxyCallError(err, proxyAddress.String(), "implementation", string(l.chainType))
	}

	// 3. Check result data
	if len(resultData) == 0 {
		return nil, errors.NewABIProxyEmptyResultError(proxyAddress.String(), "implementation")
	}

	// 4. Unpack result
	unpackedArgs, err := l.abiUtils.Unpack(proxyABIString, "implementation", resultData)
	if err != nil {
		return nil, errors.NewABIProxyUnpackError(err, proxyAddress.String(), "implementation", "unpacking failed")
	}

	// 5. Extract the implementation address from the result
	implAddress, err := l.abiUtils.GetAddressFromArgs(unpackedArgs, "0")
	if err != nil {
		return nil, errors.NewABIProxyAddressConversionError(err, "implementation return value", string(l.chainType))
	}

	return &implAddress, nil
}
