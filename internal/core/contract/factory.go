// Package contract provides functionality for interacting with smart contracts
// across different blockchain networks.

package contract

import (
	"context"
	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/wallet"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Factory creates SmartContract instances based on a provided wallet.
// It provides a centralized way to create contract instances
// for the specific blockchain network associated with the wallet.
type Factory interface {
	// NewManager creates a ContractManager instance tailored for the provided wallet's chain.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - wallet: The wallet instance determining the target blockchain and credentials.
	//
	// Returns:
	//   - ContractManager: The contract instance configured for the wallet's chain.
	//   - error: Any error during creation, such as unsupported chain type or blockchain client issues.
	NewManager(ctx context.Context, wallet wallet.Wallet) (ContractManager, error)
}

// NewFactory creates a new ContractManager factory.
//
// Parameters:
//   - blockchainRegistry: Registry for creating blockchain clients
//   - explorerFactory: Factory for creating BlockExplorer instances
//   - cfg: Application configuration
//
// Returns:
//   - Factory: The configured contract factory instance
func NewFactory(blockchainRegistry blockchain.Factory, explorerFactory blockexplorer.Factory, cfg *config.Config) Factory {
	return &factory{
		blockchainRegistry: blockchainRegistry,
		explorerFactory:    explorerFactory,
		cfg:                cfg,
	}
}

type factory struct {
	blockchainRegistry blockchain.Factory
	explorerFactory    blockexplorer.Factory
	cfg                *config.Config
}

// NewManager implements the Factory interface. It creates a new SmartContract instance
// for the chain associated with the provided wallet.
func (f *factory) NewManager(ctx context.Context, wallet wallet.Wallet) (ContractManager, error) {
	// Get chain type from the provided wallet
	chainType := wallet.Chain().Type

	// Get blockchain client for the derived chain type
	blockchainClient, err := f.blockchainRegistry.NewClient(chainType)
	if err != nil {
		return nil, err
	}

	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// These are all EVM-compatible chains, so use EVMSmartContract
		return NewEVMSmartContract(blockchainClient, wallet, f.cfg)
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
