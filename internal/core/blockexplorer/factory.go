package blockexplorer

import (
	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Factory creates and manages BlockExplorer instances
type Factory interface {
	// GetExplorer returns a BlockExplorer instance for the specified chain type
	GetExplorer(chainType types.ChainType) (BlockExplorer, error)
}

// NewFactory creates a new BlockExplorer factory
func NewFactory(chains *types.Chains, cfg *config.Config, logger logger.Logger) Factory {
	return &factory{
		chains:    chains,
		cfg:       cfg,
		logger:    logger,
		explorers: make(map[types.ChainType]BlockExplorer),
	}
}

type factory struct {
	chains    *types.Chains
	cfg       *config.Config
	logger    logger.Logger
	explorers map[types.ChainType]BlockExplorer
}

// GetExplorer returns a BlockExplorer instance for the specified chain type
func (f *factory) GetExplorer(chainType types.ChainType) (BlockExplorer, error) {
	// Check if we already have an instance for this chain
	if explorer, ok := f.explorers[chainType]; ok {
		return explorer, nil
	}

	// Get chain information
	chain, err := f.chains.Get(chainType)
	if err != nil {
		return nil, err
	}

	// Create a new explorer instance based on chain type
	var explorer BlockExplorer

	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// Create EVM-compatible explorer
		explorer = NewEtherscanExplorer(chain, chain.ExplorerAPIUrl, chain.ExplorerUrl, chain.ExplorerAPIKey, f.logger)
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}

	// Store the explorer instance
	f.explorers[chainType] = explorer
	return explorer, nil
}
