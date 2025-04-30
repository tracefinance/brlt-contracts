package transaction

import (
	"vault0/internal/core/abiutils"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Factory creates and manages Mapper instances
type Factory interface {
	// NewMapper returns a Mapper instance for the specified chain type
	NewMapper(chainType types.ChainType) (Mapper, error)
}

// NewFactory creates a new transaction Mapper factory
func NewFactory(tokenStore tokenstore.TokenStore, log logger.Logger, abiUtilsFactory abiutils.Factory) Factory {
	return &factory{
		tokenStore:      tokenStore,
		log:             log,
		abiUtilsFactory: abiUtilsFactory,
		mappers:         make(map[types.ChainType]Mapper),
	}
}

type factory struct {
	tokenStore      tokenstore.TokenStore
	log             logger.Logger
	abiUtilsFactory abiutils.Factory
	mappers         map[types.ChainType]Mapper
}

// NewMapper returns a Mapper instance for the specified chain type
func (f *factory) NewMapper(chainType types.ChainType) (Mapper, error) {
	// Check if we already have an instance for this chain
	if mapper, ok := f.mappers[chainType]; ok {
		return mapper, nil
	}

	// Create a new ABI utils instance for the chain type
	abiUtils, err := f.abiUtilsFactory.NewABIUtils(chainType)
	if err != nil {
		return nil, err
	}

	// Create a new mapper instance based on chain type
	var mapper Mapper

	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		mapper = NewEvmMapper(f.tokenStore, f.log, abiUtils)
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}

	// Store the mapper instance
	f.mappers[chainType] = mapper
	return mapper, nil
}
