package blockchain

import (
	"sync"
	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// Factory is an interface for managing blockchain clients
type Factory interface {
	// NewClient returns a blockchain client for the specified chain type.
	// Returns ErrCodeChainNotSupported if the chain type is not supported.
	NewClient(chainType types.ChainType) (BlockchainClient, error)

	// NewMonitor returns a blockchain monitor for the specified chain type.
	// Returns ErrCodeChainNotSupported if the chain type is not supported.
	NewMonitor(chainType types.ChainType) (BLockchainEventMonitor, error)
}

// Factory creates blockchain implementations
type factory struct {
	chains      *types.Chains
	cfg         *config.Config
	log         logger.Logger
	clients     map[types.ChainType]BlockchainClient
	clientsMux  sync.RWMutex
	monitors    map[types.ChainType]BLockchainEventMonitor
	monitorsMux sync.RWMutex
}

// NewFactory creates a new blockchain factory with the given configuration
func NewFactory(chains *types.Chains, cfg *config.Config, log logger.Logger) Factory {
	return &factory{
		cfg:      cfg,
		log:      log,
		chains:   chains,
		clients:  make(map[types.ChainType]BlockchainClient),
		monitors: make(map[types.ChainType]BLockchainEventMonitor),
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
		if chain.RPCUrl == "" {
			return nil, errors.NewInvalidBlockchainConfigError(string(chain.Type), "rpc_url")
		}

		// Create a new Ethereum RPC client
		rpcClient, err := rpc.Dial(chain.RPCUrl)
		if err != nil {
			return nil, errors.NewRPCError(err)
		}

		// Create an Ethereum client from the RPC client
		ethClient := ethclient.NewClient(rpcClient)
		client, err := NewEVMBlockchainClient(chain, rpcClient, ethClient, f.log)
		if err != nil {
			return nil, err
		}

		f.clients[chainType] = client
		return client, nil
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}

// NewMonitor returns a blockchain monitor for the specified chain type.
func (f *factory) NewMonitor(chainType types.ChainType) (BLockchainEventMonitor, error) {
	f.monitorsMux.RLock()
	if monitor, exists := f.monitors[chainType]; exists {
		f.monitorsMux.RUnlock()
		return monitor, nil
	}
	f.monitorsMux.RUnlock()

	// If not cached, need exclusive lock to create and store
	f.monitorsMux.Lock()
	defer f.monitorsMux.Unlock()

	// Double-check if another goroutine created it while we waited for the lock
	if monitor, exists := f.monitors[chainType]; exists {
		return monitor, nil
	}

	// Get the client for the chain, utilizing the client cache
	client, err := f.NewClient(chainType)
	if err != nil {
		return nil, err
	}

	// Create and cache the new monitor
	monitor := NewMonitor(f.log, client)
	f.monitors[chainType] = monitor

	return monitor, nil
}
