package wallet

import (
	"context"
	"vault0/internal/config"
	"vault0/internal/core/keystore"
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
}

// NewFactory creates a new factory instance
func NewFactory(keystore keystore.KeyStore, chains *types.Chains, config *config.Config) Factory {
	return &factory{
		keystore: keystore,
		chains:   chains,
		config:   config,
	}
}

// NewWallet creates a new wallet for the given chain type and key ID
func (f *factory) NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (Wallet, error) {
	return NewWallet(ctx, f.keystore, f.chains, f.config, chainType, keyID)
}
