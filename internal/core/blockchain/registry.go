package blockchain

import (
	"fmt"
	"sync"
	"vault0/internal/config"
	"vault0/internal/types"
)

type Registry interface {
	GetBlockchain(chainType types.ChainType) (Blockchain, error)
}

// Factory creates blockchain implementations
type registry struct {
	cfg        *config.Config
	chains     types.Chains
	clients    map[types.ChainType]Blockchain
	clientsMux sync.RWMutex
}

// NewRegistry creates a new blockchain registry with the given configuration
func NewRegistry(chains types.Chains, cfg *config.Config) Registry {
	return &registry{
		cfg:     cfg,
		chains:  chains,
		clients: make(map[types.ChainType]Blockchain),
	}
}

// NewBlockchain creates a new blockchain client for the specified chain type
func (r *registry) GetBlockchain(chainType types.ChainType) (Blockchain, error) {
	r.clientsMux.Lock()
	defer r.clientsMux.Unlock()

	// Check if we already have a client for this chain type
	if client, exists := r.clients[chainType]; exists {
		return client, nil
	}

	chain := r.chains[chainType]

	// Create a new client
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		client, err := NewEVMBlockchain(chain)
		if err != nil {
			return nil, err
		}
		r.clients[chainType] = client
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported chain type: %s", chain.Type)
	}
}
