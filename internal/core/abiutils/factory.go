package abiutils

import (
	"fmt"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Factory is an interface for creating ABIUtils instances for different chains.
type Factory interface {
	// NewABIUtils returns an ABI utility implementation for the specified chain type.
	// Returns ErrCodeChainNotSupported if the chain type is not supported.
	NewABIUtils(chainType types.ChainType) (ABIUtils, error)
}

// factory implements the Factory interface.
type factory struct {
	cfg             *config.Config
	log             logger.Logger
	explorerFactory blockexplorer.Factory
	utilsCache      map[types.ChainType]ABIUtils
	utilsMux        sync.RWMutex
}

// NewFactory creates a new ABIUtils factory.
func NewFactory(cfg *config.Config, log logger.Logger, explorerFactory blockexplorer.Factory) Factory {
	return &factory{
		cfg:             cfg,
		log:             log,
		explorerFactory: explorerFactory,
		utilsCache:      make(map[types.ChainType]ABIUtils),
	}
}

// NewABIUtils returns an ABI utility implementation for the specified chain type.
func (f *factory) NewABIUtils(chainType types.ChainType) (ABIUtils, error) {
	f.utilsMux.RLock()
	cachedUtil, exists := f.utilsCache[chainType]
	f.utilsMux.RUnlock()

	if exists {
		return cachedUtil, nil
	}

	// Acquire write lock for creation
	f.utilsMux.Lock()
	defer f.utilsMux.Unlock()

	// Double-check if another goroutine created it while waiting for the lock
	if cachedUtil, exists = f.utilsCache[chainType]; exists {
		return cachedUtil, nil
	}

	// Create a new ABIUtils instance based on the chain type
	var newUtil ABIUtils
	var err error

	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		newUtil, err = NewEvmAbiUtils(chainType, f.cfg, f.explorerFactory)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.NewChainNotSupportedError(fmt.Sprintf("ABIUtils for chain type '%s'", chainType))
	}

	f.utilsCache[chainType] = newUtil
	f.log.Debug("Created and cached ABIUtils", logger.String("chain_type", string(chainType)))

	return newUtil, nil
}
