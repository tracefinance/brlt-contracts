package wallet

import (
	"context"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/keystore"
)

// Factory provides methods to create different wallet types.
// Unlike the previous implementation, this factory is stateless and
// only focused on creating wallet instances with the proper dependencies.
type Factory struct {
	keyStore  keystore.KeyStore
	appConfig *config.Config
}

// NewFactory creates a new wallet factory.
func NewFactory(keyStore keystore.KeyStore, appConfig *config.Config) *Factory {
	if appConfig == nil {
		panic("appConfig must not be nil")
	}

	return &Factory{
		keyStore:  keyStore,
		appConfig: appConfig,
	}
}

// CreateWallet creates a new wallet instance for the specified chain type.
// The consumer is responsible for caching and lifecycle management.
func (f *Factory) CreateWallet(ctx context.Context, chainType ChainType) (Wallet, error) {
	switch chainType {
	case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
		// All EVM-compatible chains use the same implementation
		return NewEVMWallet(f.keyStore, chainType, f.appConfig)
	default:
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedChain, chainType)
	}
}
