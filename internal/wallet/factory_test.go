package wallet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"vault0/internal/types"
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

		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.CreateWallet(ctx, types.ChainTypeEthereum, keyID)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, types.ChainTypeEthereum, wallet.ChainType())
	})

	t.Run("Create Polygon wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.CreateWallet(ctx, types.ChainTypePolygon, keyID)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, types.ChainTypePolygon, wallet.ChainType())
	})

	t.Run("Create Base wallet", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.CreateWallet(ctx, types.ChainTypeBase, keyID)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, types.ChainTypeBase, wallet.ChainType())
	})

	t.Run("Create wallet for unsupported chain", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		factory := NewFactory(keyStore, appConfig)

		unsupportedChainType := types.ChainType("unsupported")
		// Use a test keyID
		keyID := "test-key-id"
		wallet, err := factory.CreateWallet(ctx, unsupportedChainType, keyID)

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.Contains(t, err.Error(), "unsupported blockchain")
	})
}
