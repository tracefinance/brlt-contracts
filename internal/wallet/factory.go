package wallet

import (
	"context"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/keystore"
	"vault0/internal/types"
)

// Factory provides methods to create different wallet types.
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

// NewWallet creates a new wallet instance for the specified chain type and key ID.
func (f *Factory) NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (Wallet, error) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// All EVM-compatible chains use the same implementation
		return NewEVMWallet(f.keyStore, chainType, keyID, f.appConfig)
	default:
		return nil, fmt.Errorf("%w: %s", types.ErrUnsupportedChain, chainType)
	}
}
