package wallet

import (
	"context"
	"fmt"
	"sync"

	"vault0/internal/config"
	"vault0/internal/keystore"
)

// Factory creates and manages wallet instances
type Factory struct {
	keyStore  keystore.KeyStore
	appConfig *config.Config
	config    map[ChainType]interface{}
	wallets   map[ChainType]Wallet
	mu        sync.RWMutex
}

// NewFactory creates a new wallet factory
func NewFactory(keyStore keystore.KeyStore, appConfig *config.Config) *Factory {
	// Ensure appConfig is never nil
	if appConfig == nil {
		panic("appConfig must not be nil")
	}

	return &Factory{
		keyStore:  keyStore,
		appConfig: appConfig,
		config:    make(map[ChainType]interface{}),
		wallets:   make(map[ChainType]Wallet),
	}
}

// SetConfig sets configuration for a specific chain type
func (f *Factory) SetConfig(chainType ChainType, config interface{}) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.config[chainType] = config
}

// GetWallet returns a wallet for the specified chain type
func (f *Factory) GetWallet(ctx context.Context, chainType ChainType) (Wallet, error) {
	f.mu.RLock()
	wallet, exists := f.wallets[chainType]
	f.mu.RUnlock()

	if exists {
		return wallet, nil
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Check again in case another goroutine created it while we were waiting for the write lock
	wallet, exists = f.wallets[chainType]
	if exists {
		return wallet, nil
	}

	// Get config for this chain type
	config, exists := f.config[chainType]
	if !exists {
		config = nil // Use default config
	}

	// Create a new wallet based on chain type
	var err error

	// Helper function to create wallet by type
	createWallet := func() (Wallet, error) {
		switch chainType {
		case ChainTypeEthereum, ChainTypePolygon, ChainTypeBase:
			// All EVM-compatible chains use the same implementation
			return NewEVMWallet(f.keyStore, chainType, config, f.appConfig)
		default:
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedChain, chainType)
		}
	}

	// Create the wallet
	wallet, err = createWallet()
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet for %s: %w", chainType, err)
	}

	f.wallets[chainType] = wallet
	return wallet, nil
}
