package wallet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewFactory(t *testing.T) {
	t.Run("Create new factory with valid parameters", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		factory := NewFactory(keyStore, appConfig)

		assert.NotNil(t, factory)
		assert.Equal(t, keyStore, factory.keyStore)
		assert.Equal(t, appConfig, factory.appConfig)
	})

	t.Run("Create new factory with nil config should panic", func(t *testing.T) {
		keyStore := new(MockKeyStore)

		assert.Panics(t, func() {
			NewFactory(keyStore, nil)
		})
	})
}

func TestCreateWallet(t *testing.T) {
	t.Run("Create Ethereum wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		wallet, err := factory.CreateWallet(ctx, ChainTypeEthereum)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeEthereum, wallet.ChainType())
	})

	t.Run("Create Polygon wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		wallet, err := factory.CreateWallet(ctx, ChainTypePolygon)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypePolygon, wallet.ChainType())
	})

	t.Run("Create Base wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		wallet, err := factory.CreateWallet(ctx, ChainTypeBase)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeBase, wallet.ChainType())
	})

	t.Run("Create wallet for unsupported chain", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		unsupportedChainType := ChainType("unsupported")
		wallet, err := factory.CreateWallet(ctx, unsupportedChainType)

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.Contains(t, err.Error(), "unsupported blockchain")
	})
}
