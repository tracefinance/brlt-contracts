package blockchain

import (
	"sync"
	"vault0/internal/config"
	"vault0/internal/types"
)

// Factory creates blockchain implementations
type Factory struct {
	cfg        *config.Config
	clients    map[types.ChainType]Blockchain
	clientsMux sync.RWMutex
}

// NewFactory creates a new blockchain factory with the given configuration
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		cfg:     cfg,
		clients: make(map[types.ChainType]Blockchain),
	}
}

// NewBlockchain creates a new blockchain client for the specified chain type
func (f *Factory) NewBlockchain(chainType types.ChainType) (Blockchain, error) {
	f.clientsMux.Lock()
	defer f.clientsMux.Unlock()

	// Check if we already have a client for this chain type
	if client, exists := f.clients[chainType]; exists {
		return client, nil
	}

	// Create a new client
	chain, err := NewChain(chainType, f.cfg)
	if err != nil {
		return nil, err
	}

	client, err := NewEVMBlockchain(chain)
	if err != nil {
		return nil, err
	}

	// Store the client in the map for future use
	f.clients[chainType] = client
	return client, nil
}
