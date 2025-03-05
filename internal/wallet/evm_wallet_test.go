package wallet

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"vault0/internal/keygen"
	"vault0/internal/keystore"
)

// Test NewEVMWallet function
func TestNewEVMWallet(t *testing.T) {
	t.Run("Create EVM wallet with valid parameters", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, err := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeEthereum, wallet.chainType)
		assert.Equal(t, keyStore, wallet.keyStore)
		assert.NotNil(t, wallet.config)
	})

	t.Run("Create EVM wallet with custom config", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Custom config
		customConfig := &EVMConfig{
			ChainID:         big.NewInt(1),
			DefaultGasLimit: 50000,
			DefaultGasPrice: big.NewInt(30000000000),
		}

		wallet, err := NewEVMWallet(keyStore, ChainTypeEthereum, customConfig, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeEthereum, wallet.chainType)
		assert.Equal(t, keyStore, wallet.keyStore)
		assert.Equal(t, customConfig, wallet.config)
	})

	t.Run("Error when ChainID is not provided", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Custom config without ChainID
		customConfig := &EVMConfig{
			DefaultGasLimit: 50000,
			DefaultGasPrice: big.NewInt(30000000000),
		}

		wallet, err := NewEVMWallet(keyStore, ChainTypeEthereum, customConfig, appConfig)

		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.Contains(t, err.Error(), "Chain ID is required")
	})
}

// Test ChainType method
func TestEVMWallet_ChainType(t *testing.T) {
	t.Run("Return correct chain type", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)
		assert.Equal(t, ChainTypeEthereum, wallet.ChainType())

		wallet, _ = NewEVMWallet(keyStore, ChainTypePolygon, nil, appConfig)
		assert.Equal(t, ChainTypePolygon, wallet.ChainType())

		wallet, _ = NewEVMWallet(keyStore, ChainTypeBase, nil, appConfig)
		assert.Equal(t, ChainTypeBase, wallet.ChainType())
	})
}

// Test DeriveAddress method
func TestEVMWallet_DeriveAddress(t *testing.T) {
	t.Run("Derive address from valid public key", func(t *testing.T) {
		// Generate a test private key
		privateKey, _ := createTestPrivateKey()

		// Get the public key
		publicKey := privateKey.Public()
		publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
		assert.True(t, ok)

		// Get the expected address
		expectedAddress := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()

		// Convert the public key to bytes
		publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)

		// Create wallet
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Derive address
		address, err := wallet.DeriveAddress(context.Background(), publicKeyBytes)

		assert.NoError(t, err)
		assert.Equal(t, expectedAddress, address)
	})

	t.Run("Return error for invalid public key", func(t *testing.T) {
		// Create wallet
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Try to derive address from invalid public key
		address, err := wallet.DeriveAddress(context.Background(), []byte("invalid public key"))

		assert.Error(t, err)
		assert.Empty(t, address)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})
}

// Test CreateNativeTransaction method
func TestEVMWallet_CreateNativeTransaction(t *testing.T) {
	t.Run("Create native transaction with valid parameters", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000000000000000) // 1 ETH
		options := TransactionOptions{
			GasPrice: big.NewInt(25000000000), // 25 Gwei
			GasLimit: 30000,
			Nonce:    5,
			Data:     []byte{1, 2, 3, 4},
		}

		tx, err := wallet.CreateNativeTransaction(ctx, fromAddress, toAddress, amount, options)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, fromAddress, tx.From)
		assert.Equal(t, toAddress, tx.To)
		assert.Equal(t, amount, tx.Value)
		assert.Equal(t, options.Data, tx.Data)
		assert.Equal(t, options.Nonce, tx.Nonce)
		assert.Equal(t, options.GasPrice, tx.GasPrice)
		assert.Equal(t, options.GasLimit, tx.GasLimit)
		assert.Equal(t, TransactionTypeNative, tx.Type)
	})

	t.Run("Create native transaction with default gas options", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000000000000000) // 1 ETH
		options := TransactionOptions{
			Nonce: 5,
		}

		tx, err := wallet.CreateNativeTransaction(ctx, fromAddress, toAddress, amount, options)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, wallet.config.DefaultGasPrice, tx.GasPrice)
		assert.Equal(t, wallet.config.DefaultGasLimit, tx.GasLimit)
	})

	t.Run("Error with invalid address", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		toAddress := "invalid-address"
		amount := big.NewInt(1000000000000000000) // 1 ETH
		options := TransactionOptions{}

		tx, err := wallet.CreateNativeTransaction(ctx, fromAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})

	t.Run("Error with zero amount", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(0)
		options := TransactionOptions{}

		tx, err := wallet.CreateNativeTransaction(ctx, fromAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})
}

// Test CreateTokenTransaction method
func TestEVMWallet_CreateTokenTransaction(t *testing.T) {
	t.Run("Create token transaction with valid parameters", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC on Ethereum
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000) // 1 USDC (6 decimals)
		options := TransactionOptions{
			GasPrice: big.NewInt(25000000000), // 25 Gwei
			GasLimit: 65000,
			Nonce:    5,
		}

		tx, err := wallet.CreateTokenTransaction(ctx, fromAddress, tokenAddress, toAddress, amount, options)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, fromAddress, tx.From)
		assert.Equal(t, tokenAddress, tx.To)
		assert.Equal(t, big.NewInt(0), tx.Value) // 0 ETH for token transfers
		assert.NotEmpty(t, tx.Data)              // Should contain the token transfer data
		assert.Equal(t, options.Nonce, tx.Nonce)
		assert.Equal(t, options.GasPrice, tx.GasPrice)
		assert.Equal(t, options.GasLimit, tx.GasLimit)
		assert.Equal(t, TransactionTypeERC20, tx.Type)
		assert.Equal(t, tokenAddress, tx.TokenAddress)
	})

	t.Run("Create token transaction with default gas options", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC on Ethereum
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000) // 1 USDC (6 decimals)
		options := TransactionOptions{
			Nonce: 5,
		}

		tx, err := wallet.CreateTokenTransaction(ctx, fromAddress, tokenAddress, toAddress, amount, options)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, wallet.config.DefaultGasPrice, tx.GasPrice)
		assert.Equal(t, uint64(65000), tx.GasLimit) // Default for ERC20 transfers
	})

	t.Run("Error with invalid address", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC on Ethereum
		toAddress := "invalid-address"
		amount := big.NewInt(1000000) // 1 USDC (6 decimals)
		options := TransactionOptions{}

		tx, err := wallet.CreateTokenTransaction(ctx, fromAddress, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})

	t.Run("Error with invalid token address", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		tokenAddress := "invalid-token-address"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000) // 1 USDC (6 decimals)
		options := TransactionOptions{}

		tx, err := wallet.CreateTokenTransaction(ctx, fromAddress, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})

	t.Run("Error with zero amount", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC on Ethereum
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(0)
		options := TransactionOptions{}

		tx, err := wallet.CreateTokenTransaction(ctx, fromAddress, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})
}

// Test SignTransaction method
func TestEVMWallet_SignTransaction(t *testing.T) {
	t.Run("Error on address mismatch", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Generate a test key
		privateKey, _ := createTestPrivateKey()
		publicKey := privateKey.Public().(*ecdsa.PublicKey)

		// Create a test key
		keyID := "test-key-id"
		testKey := &keystore.Key{
			ID:        keyID,
			Name:      "Test Key",
			Type:      keygen.KeyTypeECDSA,
			PublicKey: crypto.FromECDSAPub(publicKey),
		}

		// Create a transaction with a different from address
		tx := &Transaction{
			Chain:    ChainTypeEthereum,
			From:     "0xDifferentAddress",
			To:       "0xbcd4042de499d14e55001ccbb24a551f3b954096",
			Value:    big.NewInt(1000000000000000000), // 1 ETH
			Nonce:    1,
			GasPrice: big.NewInt(20000000000),
			GasLimit: 21000,
			Type:     TransactionTypeNative,
		}

		// Set up keystore mock
		keyStore.On("GetPublicKey", ctx, keyID).Return(testKey, nil)

		// Create wallet
		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Sign transaction should fail due to address mismatch
		sig, err := wallet.SignTransaction(ctx, keyID, tx)

		assert.Error(t, err)
		assert.Nil(t, sig)
		assert.Contains(t, err.Error(), "transaction from address does not match key")
	})

	t.Run("Return error for non-existent key", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Set up keystore mock to return error
		keyID := "non-existent-key"
		keyStore.On("GetPublicKey", ctx, keyID).Return((*keystore.Key)(nil), keystore.ErrKeyNotFound)

		// Create wallet
		wallet, _ := NewEVMWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Create transaction
		tx := createTestTransaction()

		// Sign transaction
		signedTx, err := wallet.SignTransaction(ctx, keyID, tx)

		assert.Error(t, err)
		assert.Nil(t, signedTx)
		assert.ErrorIs(t, err, keystore.ErrKeyNotFound)

		// Verify the mock was called
		keyStore.AssertCalled(t, "GetPublicKey", ctx, keyID)
	})
}

// Test NewEVMConfig function
func TestNewEVMConfig(t *testing.T) {
	t.Run("Create config for Ethereum", func(t *testing.T) {
		appConfig := createTestConfig()

		evmConfig := NewEVMConfig(ChainTypeEthereum, appConfig)

		assert.NotNil(t, evmConfig)
		assert.Equal(t, big.NewInt(1), evmConfig.ChainID)
		assert.Equal(t, uint64(21000), evmConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(20000000000), evmConfig.DefaultGasPrice) // 20 Gwei
	})

	t.Run("Create config for Polygon", func(t *testing.T) {
		appConfig := createTestConfig()

		polygonConfig := NewEVMConfig(ChainTypePolygon, appConfig)

		assert.NotNil(t, polygonConfig)
		assert.Equal(t, big.NewInt(137), polygonConfig.ChainID)
		assert.Equal(t, uint64(21000), polygonConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(30000000000), polygonConfig.DefaultGasPrice) // 30 Gwei
	})

	t.Run("Create config for Base", func(t *testing.T) {
		appConfig := createTestConfig()

		baseConfig := NewEVMConfig(ChainTypeBase, appConfig)

		assert.NotNil(t, baseConfig)
		assert.Equal(t, big.NewInt(8453), baseConfig.ChainID)
		assert.Equal(t, uint64(21000), baseConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(10000000000), baseConfig.DefaultGasPrice) // 10 Gwei
	})

	t.Run("Panic when appConfig is nil", func(t *testing.T) {
		assert.Panics(t, func() {
			NewEVMConfig(ChainTypeEthereum, nil)
		})
	})

	t.Run("Create minimal config for unsupported chain", func(t *testing.T) {
		appConfig := createTestConfig()

		unsupportedConfig := NewEVMConfig(ChainType("unsupported"), appConfig)

		assert.NotNil(t, unsupportedConfig)
		assert.Nil(t, unsupportedConfig.ChainID)
		assert.Equal(t, uint64(21000), unsupportedConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(20000000000), unsupportedConfig.DefaultGasPrice) // 20 Gwei
	})
}
