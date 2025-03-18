package blockchain

import (
	"sync"
	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Registry is an interface for managing blockchain clients
type Registry interface {
	// GetBlockchain returns a blockchain client for the specified chain type.
	// Returns ErrCodeChainNotSupported if the chain type is not supported.
	GetBlockchain(chainType types.ChainType) (Blockchain, error)
}

// Factory creates blockchain implementations
type registry struct {
	chains     *types.Chains
	cfg        *config.Config
	log        logger.Logger
	clients    map[types.ChainType]Blockchain
	clientsMux sync.RWMutex
}

// NewRegistry creates a new blockchain registry with the given configuration
func NewRegistry(chains *types.Chains, cfg *config.Config, log logger.Logger) Registry {
	return &registry{
		cfg:     cfg,
		log:     log,
		chains:  chains,
		clients: make(map[types.ChainType]Blockchain),
	}
}

// GetBlockchain returns a blockchain client for the specified chain type
func (r *registry) GetBlockchain(chainType types.ChainType) (Blockchain, error) {
	r.clientsMux.Lock()
	defer r.clientsMux.Unlock()

	// Check if we already have a client for this chain
	if client, exists := r.clients[chainType]; exists {
		return client, nil
	}

	// Create a new client based on the chain type
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		chain, err := r.chains.Get(chainType)
		if err != nil {
			return nil, err
		}
		client, err := NewEVMBlockchain(chain, r.log)
		if err != nil {
			return nil, err
		}
		r.clients[chainType] = client
		return client, nil
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
