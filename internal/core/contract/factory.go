// Package contract provides functionality for interacting with smart contracts
// across different blockchain networks.

package contract

import (
	"context"
	"fmt"
	"sync"
	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/wallet"
	"vault0/internal/types"
)

// Factory creates and manages SmartContract instances.
// It provides a centralized way to create and cache contract instances
// for different blockchain networks.
type Factory interface {
	// NewContract creates or returns a cached SmartContract instance for the specified chain type.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type (e.g., Ethereum, Polygon)
	//
	// Returns:
	//   - SmartContract: The contract instance for the specified chain
	//   - error: Any error during creation, such as unsupported chain type
	NewContract(ctx context.Context, chainType types.ChainType) (SmartContract, error)
}

// NewFactory creates a new SmartContract factory.
//
// Parameters:
//   - blockchainRegistry: Registry for creating blockchain clients
//   - walletFactory: Factory for creating wallet instances
//   - cfg: Application configuration
//
// Returns:
//   - Factory: The configured contract factory instance
func NewFactory(blockchainRegistry blockchain.Registry, walletFactory wallet.Factory, cfg *config.Config) Factory {
	return &factory{
		blockchainRegistry: blockchainRegistry,
		walletFactory:      walletFactory,
		cfg:                cfg,
		contracts:          make(map[types.ChainType]SmartContract),
		mu:                 sync.RWMutex{},
	}
}

type factory struct {
	blockchainRegistry blockchain.Registry
	walletFactory      wallet.Factory
	cfg                *config.Config
	contracts          map[types.ChainType]SmartContract
	mu                 sync.RWMutex
}

// NewContract implements the Factory interface.
func (f *factory) NewContract(ctx context.Context, chainType types.ChainType) (SmartContract, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Check if we already have an instance for this chain
	if contract, ok := f.contracts[chainType]; ok {
		return contract, nil
	}

	// Get blockchain client for the chain type
	blockchainClient, err := f.blockchainRegistry.GetBlockchain(chainType)
	if err != nil {
		return nil, fmt.Errorf("failed to get blockchain client: %w", err)
	}

	// Create a wallet for the chain type
	// Note: We need to get the keyID from somewhere - this might need to be passed in context
	// or we might need to modify the interface to accept keyID
	walletClient, err := f.walletFactory.NewWallet(ctx, chainType, "default")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create a new contract instance based on chain type
	var contract SmartContract
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// Create EVM-compatible contract manager
		contract, err = NewEVMSmartContract(blockchainClient, walletClient, f.cfg)
	default:
		return nil, &types.UnsupportedChainError{ChainType: chainType}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create contract for chain %s: %w", chainType, err)
	}

	// Store the contract instance
	f.contracts[chainType] = contract
	return contract, nil
}
