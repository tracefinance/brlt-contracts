package abi

import (
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Factory is an interface for managing ABI utilities and loaders
type Factory interface {
	// NewABIUtils returns an ABIUtils instance for the specified chain type.
	// Returns ErrCodeChainNotSupported if the chain type is not supported.
	NewABIUtils(chainType types.ChainType) (ABIUtils, error)

	// NewABILoader returns an ABILoader instance for the specified chain type.
	// Returns ErrCodeChainNotSupported if the chain type is not supported.
	NewABILoader(chainType types.ChainType) (ABILoader, error)
}

type factory struct {
	cfg               *config.Config
	log               logger.Logger
	blockchainFactory blockchain.Factory
	explorerFactory   blockexplorer.Factory
	abiUtils          map[types.ChainType]ABIUtils
	abiLoaders        map[types.ChainType]ABILoader
	abiUtilsMux       sync.RWMutex
	abiLoadersMux     sync.RWMutex
}

// NewFactory creates a new ABI factory
func NewFactory(
	cfg *config.Config,
	log logger.Logger,
	blockchainFactory blockchain.Factory,
	explorerFactory blockexplorer.Factory,
) Factory {
	return &factory{
		cfg:               cfg,
		log:               log,
		blockchainFactory: blockchainFactory,
		explorerFactory:   explorerFactory,
		abiUtils:          make(map[types.ChainType]ABIUtils),
		abiLoaders:        make(map[types.ChainType]ABILoader),
	}
}

// NewABIUtils returns an ABIUtils instance for the specified chain type.
func (f *factory) NewABIUtils(chainType types.ChainType) (ABIUtils, error) {
	// Try to get from cache
	f.abiUtilsMux.RLock()
	if utils, exists := f.abiUtils[chainType]; exists {
		f.abiUtilsMux.RUnlock()
		return utils, nil
	}
	f.abiUtilsMux.RUnlock()

	// Not in cache, create new
	f.abiUtilsMux.Lock()
	defer f.abiUtilsMux.Unlock()

	// Double check if another goroutine created it
	if utils, exists := f.abiUtils[chainType]; exists {
		return utils, nil
	}

	// Create and cache new instance
	utils, err := NewABIUtils(chainType, f.log)
	if err != nil {
		return nil, err
	}

	f.abiUtils[chainType] = utils
	return utils, nil
}

// NewABILoader returns an ABILoader instance for the specified chain type.
func (f *factory) NewABILoader(chainType types.ChainType) (ABILoader, error) {
	// Try to get from cache
	f.abiLoadersMux.RLock()
	if loader, exists := f.abiLoaders[chainType]; exists {
		f.abiLoadersMux.RUnlock()
		return loader, nil
	}
	f.abiLoadersMux.RUnlock()

	// Not in cache, create new
	f.abiLoadersMux.Lock()
	defer f.abiLoadersMux.Unlock()

	// Double check if another goroutine created it
	if loader, exists := f.abiLoaders[chainType]; exists {
		return loader, nil
	}

	// Get blockchain client for the chain
	blockchainClient, err := f.blockchainFactory.NewClient(chainType)
	if err != nil {
		return nil, err
	}

	// Get block explorer for the chain
	explorer, err := f.explorerFactory.NewExplorer(chainType)
	if err != nil {
		return nil, err
	}

	// Get ABI utils for the chain
	abiUtils, err := f.NewABIUtils(chainType)
	if err != nil {
		return nil, err
	}

	// Create and cache new loader
	loader := NewABILoader(
		chainType,
		f.cfg,
		explorer,
		blockchainClient,
		abiUtils,
		f.log,
	)

	f.abiLoaders[chainType] = loader
	return loader, nil
}
