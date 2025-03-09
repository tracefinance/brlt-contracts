package wallet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"vault0/internal/types"
)

func TestNewFactory(t *testing.T) {
	t.Run("Create new factory with valid parameters", func(t *testing.T) {
		keyStore := &MockKeyStore{}
		appConfig := createTestConfig()

		factory := NewFactory(keyStore, appConfig)

		assert.NotNil(t, factory)
		assert.Equal(t, keyStore, factory.keyStore)
		assert.Equal(t, appConfig, factory.appConfig)
		assert.NotNil(t, factory.blockchainFactory, "blockchainFactory should be initialized")
	})

	t.Run("Create new factory with nil config should panic", func(t *testing.T) {
		keyStore := &MockKeyStore{}

		assert.Panics(t, func() {
			NewFactory(keyStore, nil)
		})
	})
}

func TestNewWallet(t *testing.T) {
	t.Run("Create Ethereum wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := &MockKeyStore{}
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.NewWallet(ctx, types.ChainTypeEthereum, keyID)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		chain := wallet.Chain()
		assert.Equal(t, types.ChainTypeEthereum, chain.Type)
		assert.Equal(t, "Ethereum", chain.Name)
		assert.Equal(t, "ETH", chain.Symbol)
	})

	t.Run("Create Polygon wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := &MockKeyStore{}
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.NewWallet(ctx, types.ChainTypePolygon, keyID)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		chain := wallet.Chain()
		assert.Equal(t, types.ChainTypePolygon, chain.Type)
		assert.Equal(t, "Polygon", chain.Name)
		assert.Equal(t, "MATIC", chain.Symbol)
	})

	t.Run("Create Base wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := &MockKeyStore{}
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.NewWallet(ctx, types.ChainTypeBase, keyID)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		chain := wallet.Chain()
		assert.Equal(t, types.ChainTypeBase, chain.Type)
		assert.Equal(t, "Base", chain.Name)
		assert.Equal(t, "ETH", chain.Symbol)
	})

	t.Run("Create wallet for unsupported chain", func(t *testing.T) {
		ctx := context.Background()
		keyStore := &MockKeyStore{}
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		unsupportedChainType := types.ChainType("unsupported")
		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.NewWallet(ctx, unsupportedChainType, keyID)

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.Contains(t, err.Error(), "unsupported blockchain")
	})
}
