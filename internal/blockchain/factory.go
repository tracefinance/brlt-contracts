package blockchain

import (
	"fmt"
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

// NewBlockchain creates a new blockchain implementation for the given chain type
// or returns an existing instance if one has already been created (singleton pattern)
func (f *Factory) NewBlockchain(chainType types.ChainType) (Blockchain, error) {
	// Check if we already have a client for this chain type
	f.clientsMux.RLock()
	client, exists := f.clients[chainType]
	f.clientsMux.RUnlock()

	if exists {
		return client, nil
	}

	// If no client exists, create a new one with write lock
	f.clientsMux.Lock()
	defer f.clientsMux.Unlock()

	// Double-check in case another goroutine created the client while we were waiting
	if client, exists := f.clients[chainType]; exists {
		return client, nil
	}

	var chainCfg *config.BlockchainConfig
	var chainName string

	switch chainType {
	case types.ChainTypeEthereum:
		chainCfg = &f.cfg.Blockchains.Ethereum
		chainName = "Ethereum"
	case types.ChainTypePolygon:
		chainCfg = &f.cfg.Blockchains.Polygon
		chainName = "Polygon"
	case types.ChainTypeBase:
		chainCfg = &f.cfg.Blockchains.Base
		chainName = "Base"
	default:
		return nil, fmt.Errorf("unsupported chain type %s: %w", chainType, ErrChainNotSupported)
	}

	if chainCfg.RPCURL == "" {
		return nil, fmt.Errorf("missing RPC URL for %s: %w", chainName, ErrRPCConnectionFailed)
	}

	chain := Chain{
		ID:          chainCfg.ChainID,
		Type:        chainType,
		Name:        chainName,
		Symbol:      getChainSymbol(chainType),
		RPCUrl:      chainCfg.RPCURL,
		ExplorerUrl: chainCfg.ExplorerURL,
	}

	client, err := NewEVMClient(chain)
	if err != nil {
		return nil, err
	}

	// Store the client in the map for future use
	f.clients[chainType] = client
	return client, nil
}

// CloseAll closes all blockchain client connections
func (f *Factory) CloseAll() {
	f.clientsMux.Lock()
	defer f.clientsMux.Unlock()

	for _, client := range f.clients {
		client.Close()
	}
	f.clients = make(map[types.ChainType]Blockchain)
}

// getChainSymbol returns the symbol for a given chain type
func getChainSymbol(chainType types.ChainType) string {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypeBase:
		return "ETH"
	case types.ChainTypePolygon:
		return "MATIC"
	default:
		return ""
	}
}
