package blockchain

import (
	"sync"
	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Factory is an interface for managing blockchain clients
type Factory interface {
	// NewClient returns a blockchain client for the specified chain type.
	// Returns ErrCodeChainNotSupported if the chain type is not supported.
	NewClient(chainType types.ChainType) (BlockchainClient, error)
}

// Factory creates blockchain implementations
type factory struct {
	chains     *types.Chains
	cfg        *config.Config
	log        logger.Logger
	clients    map[types.ChainType]BlockchainClient
	clientsMux sync.RWMutex
}

// NewFactory creates a new blockchain factory with the given configuration
func NewFactory(chains *types.Chains, cfg *config.Config, log logger.Logger) Factory {
	return &factory{
		cfg:     cfg,
		log:     log,
		chains:  chains,
		clients: make(map[types.ChainType]BlockchainClient),
	}
}

// NewClient returns a blockchain client for the specified chain type
func (f *factory) NewClient(chainType types.ChainType) (BlockchainClient, error) {
	f.clientsMux.Lock()
	defer f.clientsMux.Unlock()

	// Check if we already have a client for this chain
	if client, exists := f.clients[chainType]; exists {
		return client, nil
	}

	// Create a new client based on the chain type
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		chain, err := f.chains.Get(chainType)
		if err != nil {
			return nil, err
		}
		client, err := NewEVMBlockchainClient(chain, f.log)
		if err != nil {
			return nil, err
		}
		f.clients[chainType] = client
		return client, nil
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
