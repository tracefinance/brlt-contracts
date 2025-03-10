package wallet

import (
	"context"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/keystore"
	"vault0/internal/types"
)

// Factory provides methods to create different wallet types.
type Factory struct {
	keyStore          keystore.KeyStore
	appConfig         *config.Config
	blockchainFactory *blockchain.Factory
}

// NewFactory creates a new wallet factory.
func NewFactory(keyStore keystore.KeyStore, appConfig *config.Config) *Factory {
	if appConfig == nil {
		panic("appConfig must not be nil")
	}

	return &Factory{
		keyStore:          keyStore,
		appConfig:         appConfig,
		blockchainFactory: blockchain.NewFactory(appConfig),
	}
}

// NewWallet creates a new wallet instance for the specified chain type and key ID.
func (f *Factory) NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (Wallet, error) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// Get chain struct from blockchain package
		chain, err := blockchain.NewChain(chainType, f.appConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create chain: %w", err)
		}

		// All EVM-compatible chains use the same implementation
		return NewEVMWallet(f.keyStore, chain, keyID, f.appConfig)
	default:
		return nil, fmt.Errorf("%w: %s", types.ErrUnsupportedChain, chainType)
	}
}
