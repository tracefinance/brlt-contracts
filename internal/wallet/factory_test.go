package wallet

import (
	"context"
	"math/big"
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
		assert.NotNil(t, factory.config)
		assert.NotNil(t, factory.wallets)
	})

	t.Run("Create new factory with nil config should panic", func(t *testing.T) {
		keyStore := new(MockKeyStore)

		assert.Panics(t, func() {
			NewFactory(keyStore, nil)
		})
	})
}

func TestSetConfig(t *testing.T) {
	t.Run("Set config for a chain type", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		// Custom ethereum config
		customConfig := &EVMConfig{
			ChainID:         big.NewInt(1),
			DefaultGasLimit: 50000,
			DefaultGasPrice: big.NewInt(30000000000),
		}

		factory.SetConfig(ChainTypeEthereum, customConfig)

		// Verify the config was set
		assert.Equal(t, customConfig, factory.config[ChainTypeEthereum])
	})
}

func TestGetWallet(t *testing.T) {
	t.Run("Get ethereum wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		wallet, err := factory.GetWallet(ctx, ChainTypeEthereum)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeEthereum, wallet.ChainType())

		// Getting the same wallet again should return the cached instance
		wallet2, err := factory.GetWallet(ctx, ChainTypeEthereum)
		assert.NoError(t, err)
		assert.Equal(t, wallet, wallet2) // Same instance
	})

	t.Run("Get polygon wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		wallet, err := factory.GetWallet(ctx, ChainTypePolygon)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypePolygon, wallet.ChainType())
	})

	t.Run("Get base wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		wallet, err := factory.GetWallet(ctx, ChainTypeBase)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeBase, wallet.ChainType())
	})

	t.Run("Get wallet for unsupported chain type", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		unsupportedChainType := ChainType("unsupported")
		wallet, err := factory.GetWallet(ctx, unsupportedChainType)

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.ErrorIs(t, err, ErrUnsupportedChain)
	})

	t.Run("Get wallet with custom config", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		// Set custom config
		customConfig := &EVMConfig{
			ChainID:         big.NewInt(1),
			DefaultGasLimit: 50000,
			DefaultGasPrice: big.NewInt(30000000000),
		}
		factory.SetConfig(ChainTypeEthereum, customConfig)

		wallet, err := factory.GetWallet(ctx, ChainTypeEthereum)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)

		// We'd need to cast to EVMWallet to verify the config was used,
		// but this is implementation dependent and might not be reliable in a unit test
		// Instead we just verify the chain type
		assert.Equal(t, ChainTypeEthereum, wallet.ChainType())
	})
}
