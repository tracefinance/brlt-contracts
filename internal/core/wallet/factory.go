package wallet

import (
	"context"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/keystore"
	"vault0/internal/types"
)

type Factory interface {
	NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (Wallet, error)
}

// Factory provides methods to create different wallet types.
type factory struct {
	keyStore     keystore.KeyStore
	appConfig    *config.Config
	chainFactory types.ChainFactory
}

// NewFactory creates a new wallet factory.
func NewFactory(keyStore keystore.KeyStore, appConfig *config.Config) Factory {
	if appConfig == nil {
		panic("appConfig must not be nil")
	}
	chainFactory := types.NewChainFactory(appConfig)
	return &factory{
		keyStore:     keyStore,
		appConfig:    appConfig,
		chainFactory: chainFactory,
	}
}

// NewWallet creates a new wallet instance for the specified chain type and key ID.
func (f *factory) NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (Wallet, error) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// Get chain struct from blockchain package
		chain, err := f.chainFactory.NewChain(chainType)
		if err != nil {
			return nil, fmt.Errorf("failed to create chain: %w", err)
		}

		// All EVM-compatible chains use the same implementation
		return NewEVMWallet(f.keyStore, chain, keyID)
	default:
		return nil, fmt.Errorf("%w: %s", types.ErrUnsupportedChain, chainType)
	}
}
