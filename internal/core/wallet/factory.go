package wallet

import (
	"context"
	"vault0/internal/config"
	"vault0/internal/core/abi"
	"vault0/internal/core/keystore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Factory is the interface for creating wallet instances
type Factory interface {
	NewManager(ctx context.Context, chainType types.ChainType, keyID string) (WalletManager, error)
}

// Factory creates wallet instances for different chains
type factory struct {
	keystore   keystore.KeyStore
	chains     *types.Chains
	abiFactory abi.Factory
	config     *config.Config
	log        logger.Logger
}

// NewFactory creates a new factory instance
func NewFactory(keystore keystore.KeyStore, chains *types.Chains, abiFactory abi.Factory, config *config.Config, log logger.Logger) Factory {
	return &factory{
		keystore:   keystore,
		chains:     chains,
		abiFactory: abiFactory,
		config:     config,
		log:        log,
	}
}

// NewManager creates a new wallet for the given chain type and key ID
func (f *factory) NewManager(ctx context.Context, chainType types.ChainType, keyID string) (WalletManager, error) {
	chain, err := f.chains.Get(chainType)
	if err != nil {
		return nil, err
	}
	abiUtils, err := f.abiFactory.NewABIUtils(chainType)
	if err != nil {
		return nil, err
	}
	abiLoader, err := f.abiFactory.NewABILoader(chainType)
	if err != nil {
		return nil, err
	}

	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		return NewEVMWallet(keyID, chain, f.keystore, abiUtils, abiLoader, f.log)
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
