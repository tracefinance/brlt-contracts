package wallet

import (
	"context"
	"vault0/internal/config"
	"vault0/internal/core/keystore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Factory is the interface for creating wallet instances
type Factory interface {
	NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (Wallet, error)
}

// Factory creates wallet instances for different chains
type factory struct {
	keystore keystore.KeyStore
	chains   *types.Chains
	config   *config.Config
	log      logger.Logger
}

// NewFactory creates a new factory instance
func NewFactory(keystore keystore.KeyStore, chains *types.Chains, config *config.Config, log logger.Logger) Factory {
	return &factory{
		keystore: keystore,
		chains:   chains,
		config:   config,
		log:      log,
	}
}

// NewWallet creates a new wallet for the given chain type and key ID
func (f *factory) NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (Wallet, error) {
	chain, err := f.chains.Get(chainType)
	if err != nil {
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}

	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		return NewEVMWallet(keyID, chain, f.keystore, f.log)
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
