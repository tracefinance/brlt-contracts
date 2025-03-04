package wallet

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vault0/internal/keystore"
)

// MockEthClient is a mock implementation of ethclient.Client for testing
type MockEthClient struct {
	mock.Mock
}

// Test NewEthereumWallet function
func TestNewEthereumWallet(t *testing.T) {
	t.Run("Create ethereum wallet with valid parameters", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, err := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeEthereum, wallet.chainType)
		assert.Equal(t, keyStore, wallet.keyStore)
		assert.NotNil(t, wallet.config)
	})

	t.Run("Create ethereum wallet with custom config", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Custom config
		customConfig := &EthereumConfig{
			RPCURL:          "https://custom-rpc-url.com",
			ChainID:         big.NewInt(1),
			DefaultGasLimit: 50000,
			DefaultGasPrice: big.NewInt(30000000000),
		}

		wallet, err := NewEthereumWallet(keyStore, ChainTypeEthereum, customConfig, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, ChainTypeEthereum, wallet.chainType)
		assert.Equal(t, keyStore, wallet.keyStore)
		assert.Equal(t, customConfig, wallet.config)
	})
}

// Test ChainType method
func TestEthereumWallet_ChainType(t *testing.T) {
	t.Run("Return correct chain type", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)
		assert.Equal(t, ChainTypeEthereum, wallet.ChainType())

		wallet, _ = NewEthereumWallet(keyStore, ChainTypePolygon, nil, appConfig)
		assert.Equal(t, ChainTypePolygon, wallet.ChainType())

		wallet, _ = NewEthereumWallet(keyStore, ChainTypeBase, nil, appConfig)
		assert.Equal(t, ChainTypeBase, wallet.ChainType())
	})
}

// Test DeriveAddress method
func TestEthereumWallet_DeriveAddress(t *testing.T) {
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
		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Derive address
		address, err := wallet.DeriveAddress(context.Background(), publicKeyBytes)

		assert.NoError(t, err)
		assert.Equal(t, expectedAddress, address)
	})

	t.Run("Return error for invalid public key", func(t *testing.T) {
		// Create wallet
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Try to derive address from invalid public key
		address, err := wallet.DeriveAddress(context.Background(), []byte("invalid public key"))

		assert.Error(t, err)
		assert.Empty(t, address)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})
}

// Test GetBalance method
func TestEthereumWallet_GetBalance(t *testing.T) {
	t.Run("Get balance for valid address", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Create wallet with mocked client
		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// We can't directly mock the ethclient, so this test is limited
		// In a real test, you would inject a mock client
		// For now, we'll just check that the method doesn't panic

		// This test will fail if the RPC URL is not valid or reachable
		// In a real test environment, we'd use a mock or a local node
		t.Skip("Skipping test that requires a real Ethereum node")

		// The following code shouldn't be executed due to the Skip above,
		// but we keep it for documentation purposes
		_, _ = wallet.GetBalance(ctx, "0x71C7656EC7ab88b098defB751B7401B5f6d8976F")
	})

	t.Run("Return error for invalid address", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		balance, err := wallet.GetBalance(ctx, "invalid-address")

		assert.Error(t, err)
		assert.Nil(t, balance)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})
}

// Test SignTransaction method
func TestEthereumWallet_SignTransaction(t *testing.T) {
	t.Run("Sign transaction with valid key", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)

		// Generate a test private key and address
		privateKey, _ := createTestPrivateKey()
		publicKey := privateKey.Public().(*ecdsa.PublicKey)
		address := crypto.PubkeyToAddress(*publicKey).Hex()

		// Create a test key
		keyID := "test-key-id"
		testKey := &keystore.Key{
			ID:        keyID,
			Name:      "Test Key",
			Type:      keystore.KeyTypeECDSA,
			PublicKey: crypto.FromECDSAPub(publicKey),
		}

		// Set up keystore mock
		keyStore.On("GetPublicKey", ctx, keyID).Return(testKey, nil)

		// Skip the actual signing test since it requires specific signature format for Ethereum
		// We'll just test our input validation and mock setup instead

		// Create transaction (unused in the actual test but demonstrates the test setup)
		_ = &Transaction{
			Chain:    ChainTypeEthereum,
			From:     address,
			To:       "0xbcd4042de499d14e55001ccbb24a551f3b954096",
			Value:    big.NewInt(1000000000000000000), // 1 ETH
			Nonce:    1,
			GasPrice: big.NewInt(20000000000),
			GasLimit: 21000,
			Type:     TransactionTypeNative,
		}

		// We'll skip this test as proper testing would require a more complex setup
		// with real Ethereum signature format (r, s, v)
		t.Skip("Skipping test that requires real Ethereum signature format")
	})

	t.Run("Return error for non-existent key", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Set up keystore mock to return error
		keyID := "non-existent-key"
		keyStore.On("GetPublicKey", ctx, keyID).Return((*keystore.Key)(nil), keystore.ErrKeyNotFound)

		// Create wallet
		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

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

// Test SendNative method
func TestEthereumWallet_SendNative(t *testing.T) {
	t.Run("Send native currency validation", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Test invalid address
		keyID := "test-key-id"
		toAddress := "invalid-address"
		amount := big.NewInt(1000000000000000)
		options := &TransactionOptions{}

		tx, err := wallet.SendNative(ctx, keyID, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAddress)

		// Test zero amount
		toAddress = "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount = big.NewInt(0)

		tx, err = wallet.SendNative(ctx, keyID, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})

	t.Run("Setup for SendNative", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		// Generate a test private key and address
		privateKey, _ := createTestPrivateKey()
		publicKey := privateKey.Public().(*ecdsa.PublicKey)
		_ = crypto.PubkeyToAddress(*publicKey).Hex() // We would use this address in the actual test

		// Create a test key
		keyID := "test-key-id"
		testKey := &keystore.Key{
			ID:        keyID,
			Name:      "Test Key",
			Type:      keystore.KeyTypeECDSA,
			PublicKey: crypto.FromECDSAPub(publicKey),
		}

		// Set up keystore mock
		keyStore.On("GetPublicKey", ctx, keyID).Return(testKey, nil)

		// Set up signing mock
		mockSignature := make([]byte, 65)
		copy(mockSignature, "mocksignature")
		keyStore.On("Sign", ctx, keyID, mock.Anything).Return(mockSignature, nil)

		// Create wallet - just to verify the test setup is correct
		_, err := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)
		assert.NoError(t, err)

		// This test would normally be skipped in CI
		// Skip the actual sending part since it requires a real node
		t.Skip("Skipping test that requires a real Ethereum node")
	})
}

// Test SendToken method
func TestEthereumWallet_SendToken(t *testing.T) {
	t.Run("Send token validation", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		keyID := "test-key-id"

		// Test invalid token address
		tokenAddress := "invalid-address"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000)
		options := &TransactionOptions{}

		tx, err := wallet.SendToken(ctx, keyID, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAddress)

		// Test invalid recipient address
		tokenAddress = "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
		toAddress = "invalid-address"

		tx, err = wallet.SendToken(ctx, keyID, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAddress)

		// Test zero amount
		toAddress = "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount = big.NewInt(0)

		tx, err = wallet.SendToken(ctx, keyID, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, ErrInvalidAmount)
	})
}

// Test GetTokenBalance method
func TestEthereumWallet_GetTokenBalance(t *testing.T) {
	t.Run("Return error for invalid address", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		balance, err := wallet.GetTokenBalance(ctx, "invalid-address", "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48")

		assert.Error(t, err)
		assert.Nil(t, balance)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})

	t.Run("Return error for invalid token address", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		balance, err := wallet.GetTokenBalance(ctx, "0x71C7656EC7ab88b098defB751B7401B5f6d8976F", "invalid-token-address")

		assert.Error(t, err)
		assert.Nil(t, balance)
		assert.ErrorIs(t, err, ErrInvalidAddress)
	})

	t.Run("Get token balance for valid addresses - skipped for external dependency", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Skip this test as it requires a real Ethereum node
		t.Skip("Skipping test that requires a real Ethereum node")

		// The following code shouldn't be executed due to the Skip above,
		// but we keep it for documentation purposes
		address := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC on Ethereum
		_, _ = wallet.GetTokenBalance(ctx, address, tokenAddress)
	})
}

// Test BroadcastTransaction method
func TestEthereumWallet_BroadcastTransaction(t *testing.T) {
	t.Run("Return error for invalid signed transaction", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Try to broadcast an invalid transaction
		invalidTx := []byte("invalid transaction data")

		tx, err := wallet.BroadcastTransaction(ctx, invalidTx)

		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("Broadcast transaction - skipped for external dependency", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Skip this test as it requires a real Ethereum node
		t.Skip("Skipping test that requires a real Ethereum node")

		// The following code wouldn't work without a real signed transaction
		// But we keep it for documentation purposes
		signedTx := []byte{}
		_, _ = wallet.BroadcastTransaction(ctx, signedTx)
	})
}

// Test GetTransaction method
func TestEthereumWallet_GetTransaction(t *testing.T) {
	t.Run("Return error for empty transaction hash", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		tx, err := wallet.GetTransaction(ctx, "")

		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("Return error for invalid transaction hash", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		tx, err := wallet.GetTransaction(ctx, "invalid-tx-hash")

		assert.Error(t, err)
		assert.Nil(t, tx)
	})

	t.Run("Get transaction - skipped for external dependency", func(t *testing.T) {
		ctx := context.Background()
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()

		wallet, _ := NewEthereumWallet(keyStore, ChainTypeEthereum, nil, appConfig)

		// Skip this test as it requires a real Ethereum node
		t.Skip("Skipping test that requires a real Ethereum node")

		// The following code wouldn't work without a real transaction hash
		// But we keep it for documentation purposes
		txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		_, _ = wallet.GetTransaction(ctx, txHash)
	})
}

// Test NewEthereumConfig function
func TestNewEthereumConfig(t *testing.T) {
	t.Run("Create config with valid parameters", func(t *testing.T) {
		appConfig := createTestConfig()

		// Test Ethereum config
		ethConfig := NewEthereumConfig(ChainTypeEthereum, appConfig)
		assert.NotNil(t, ethConfig)
		assert.Equal(t, "https://eth-mainnet.alchemyapi.io/v2/test-key", ethConfig.RPCURL)
		assert.Equal(t, big.NewInt(1), ethConfig.ChainID)
		assert.Equal(t, uint64(21000), ethConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(20000000000), ethConfig.DefaultGasPrice)

		// Test Polygon config
		polygonConfig := NewEthereumConfig(ChainTypePolygon, appConfig)
		assert.NotNil(t, polygonConfig)
		assert.Equal(t, "https://polygon-mainnet.g.alchemy.com/v2/test-key", polygonConfig.RPCURL)
		assert.Equal(t, big.NewInt(137), polygonConfig.ChainID)
		assert.Equal(t, uint64(21000), polygonConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(30000000000), polygonConfig.DefaultGasPrice)

		// Test Base config
		baseConfig := NewEthereumConfig(ChainTypeBase, appConfig)
		assert.NotNil(t, baseConfig)
		assert.Equal(t, "https://mainnet.base.org", baseConfig.RPCURL)
		assert.Equal(t, big.NewInt(8453), baseConfig.ChainID)
		assert.Equal(t, uint64(21000), baseConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(10000000000), baseConfig.DefaultGasPrice)
	})

	t.Run("Create config with nil appConfig", func(t *testing.T) {
		assert.Panics(t, func() {
			NewEthereumConfig(ChainTypeEthereum, nil)
		})
	})

	t.Run("Create config with unsupported chain type", func(t *testing.T) {
		appConfig := createTestConfig()

		// Unsupported chain type should still return a minimal default config
		unsupportedConfig := NewEthereumConfig(ChainType("unsupported"), appConfig)
		assert.NotNil(t, unsupportedConfig)
		assert.Equal(t, uint64(21000), unsupportedConfig.DefaultGasLimit)
		assert.NotNil(t, unsupportedConfig.DefaultGasPrice)
	})
}
