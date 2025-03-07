package wallet

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vault0/internal/config"
	"vault0/internal/keygen"
	"vault0/internal/keystore"
	"vault0/internal/types"
)

// MockConfigWithoutChainID is a mock config where the GetBlockchainConfig always returns nil
type MockConfigWithoutChainID struct {
	*config.Config
}

func (m *MockConfigWithoutChainID) GetBlockchainConfig(chainType string) *config.BlockchainConfig {
	return nil
}

// MockEVMWallet is a mock for testing that returns a nil ChainID
type MockEVMWallet struct {
	*EVMWallet
}

// MockAppConfig is a mock implementation of config.Config that allows testing nil ChainID scenario
type MockAppConfig struct {
	*config.Config
}

func (m *MockAppConfig) GetBlockchainConfig(chainType string) *config.BlockchainConfig {
	return &config.BlockchainConfig{
		// ChainID is intentionally set to 0 which will cause a nil big.Int when converted
		ChainID:         0,
		DefaultGasPrice: 20,
		DefaultGasLimit: 21000,
		RPCURL:          "https://mock-url",
	}
}

// Test NewEVMWallet function
func TestNewEVMWallet(t *testing.T) {
	t.Run("Create EVM wallet with valid parameters", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyID := "test-key-id"

		wallet, err := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, wallet)
		assert.Equal(t, types.ChainTypeEthereum, wallet.chainType)
		assert.Equal(t, keyStore, wallet.keyStore)
		assert.Equal(t, keyID, wallet.keyID)
		assert.NotNil(t, wallet.config)
	})

	t.Run("Create EVM wallet with nil keystore", func(t *testing.T) {
		appConfig := createTestConfig()
		keyID := "test-key-id"

		wallet, err := NewEVMWallet(nil, types.ChainTypeEthereum, keyID, appConfig)

		assert.Error(t, err)
		assert.Nil(t, wallet)
	})

	t.Run("Error when ChainID is not provided", func(t *testing.T) {
		keyStore := new(MockKeyStore)
		mockConfig := &MockConfigWithoutChainID{Config: createTestConfig()}
		keyID := "test-key-id"

		// The mock returns nil blockchain config, so we should get an error
		wallet, err := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, mockConfig)

		// Now we expect an error since we changed NewEVMConfig to error when blockchainConfig is nil
		assert.Error(t, err)
		assert.Nil(t, wallet)
		assert.Contains(t, err.Error(), "blockchain configuration for ethereum not found")
	})
}

// Test ChainType method
func TestEVMWallet_ChainType(t *testing.T) {
	keyStore := new(MockKeyStore)
	appConfig := createTestConfig()
	keyID := "test-key-id"

	wallet, err := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)
	assert.NoError(t, err)

	chainType := wallet.ChainType()
	assert.Equal(t, types.ChainTypeEthereum, chainType)

	// Test with a different chain type
	wallet, err = NewEVMWallet(keyStore, types.ChainTypePolygon, keyID, appConfig)
	assert.NoError(t, err)

	chainType = wallet.ChainType()
	assert.Equal(t, types.ChainTypePolygon, chainType)
}

// Test DeriveAddress method
func TestEVMWallet_DeriveAddress(t *testing.T) {
	t.Run("Derive address from valid key ID", func(t *testing.T) {
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

		// Setup keyID and mock key
		keyID := "test-key-id"
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: publicKeyBytes,
		}

		// Create wallet and mock keystore
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		
		// Setup mock for GetPublicKey
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Derive address
		address, err := wallet.DeriveAddress(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, expectedAddress, address)
		keyStore.AssertExpectations(t)
	})

	t.Run("Return error when keystore returns error", func(t *testing.T) {
		// Setup keyID and expected error
		keyID := "non-existent-key"
		expectedError := errors.New("key not found")

		// Create wallet and mock keystore
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		
		// Setup mock for GetPublicKey to return error
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return((*keystore.Key)(nil), expectedError)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Try to derive address
		address, err := wallet.DeriveAddress(context.Background())

		assert.Error(t, err)
		assert.Empty(t, address)
		assert.Contains(t, err.Error(), expectedError.Error())
		keyStore.AssertExpectations(t)
	})

	t.Run("Return error for invalid public key", func(t *testing.T) {
		// Setup keyID and invalid key
		keyID := "invalid-key"
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: []byte("invalid public key"),
		}

		// Create wallet and mock keystore
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		
		// Setup mock for GetPublicKey to return invalid public key
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Try to derive address from invalid public key
		address, err := wallet.DeriveAddress(context.Background())

		assert.Error(t, err)
		assert.Empty(t, address)
		assert.ErrorIs(t, err, types.ErrInvalidAddress)
		keyStore.AssertExpectations(t)
	})
}

// Test CreateNativeTransaction method
func TestEVMWallet_CreateNativeTransaction(t *testing.T) {
	t.Run("Create native transaction with valid parameters", func(t *testing.T) {
		ctx := context.Background()
		
		// Setup test data
		keyID := "test-key-id"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000000000000000) // 1 ETH
		options := types.TransactionOptions{
			GasPrice: big.NewInt(25000000000), // 25 Gwei
			GasLimit: 30000,
			Nonce:    5,
			Data:     []byte{1, 2, 3, 4},
		}

		// Generate a private key for testing
		privateKey, _ := createTestPrivateKey()
		publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)
		expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

		// Setup mock key and keystore
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: publicKey,
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create transaction
		tx, err := wallet.CreateNativeTransaction(ctx, toAddress, amount, options)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, expectedAddress, tx.From)
		assert.Equal(t, toAddress, tx.To)
		assert.Equal(t, amount, tx.Value)
		assert.Equal(t, options.Data, tx.Data)
		assert.Equal(t, options.Nonce, tx.Nonce)
		assert.Equal(t, options.GasPrice, tx.GasPrice)
		assert.Equal(t, options.GasLimit, tx.GasLimit)
		assert.Equal(t, types.TransactionTypeNative, tx.Type)

		keyStore.AssertExpectations(t)
	})

	t.Run("Return error for invalid to address", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		keyID := "test-key-id"
		toAddress := "invalid-address"
		amount := big.NewInt(1000000000000000000) // 1 ETH
		options := types.TransactionOptions{}

		// Setup mock key and keystore
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: []byte{0x04, 1, 2, 3, 4}, // Dummy public key
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create transaction
		tx, err := wallet.CreateNativeTransaction(ctx, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, types.ErrInvalidAddress)

		keyStore.AssertExpectations(t)
	})

	t.Run("Return error for zero amount", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		keyID := "test-key-id"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(0)
		options := types.TransactionOptions{}

		// Setup mock for keystore with a valid public key
		privateKey, _ := createTestPrivateKey()
		publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)

		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: publicKey,
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create transaction
		tx, err := wallet.CreateNativeTransaction(ctx, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, types.ErrInvalidAmount)

		keyStore.AssertExpectations(t)
	})
}

// Test CreateTokenTransaction method
func TestEVMWallet_CreateTokenTransaction(t *testing.T) {
	t.Run("Create token transaction with valid parameters", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		keyID := "test-key-id"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC on Ethereum
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000) // 1 USDC (6 decimals)
		options := types.TransactionOptions{
			GasPrice: big.NewInt(25000000000), // 25 Gwei
			GasLimit: 65000,
			Nonce:    5,
		}

		// Generate a private key for testing
		privateKey, _ := createTestPrivateKey()
		publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)
		expectedAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

		// Setup mock key and keystore
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: publicKey,
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create token transaction
		tx, err := wallet.CreateTokenTransaction(ctx, tokenAddress, toAddress, amount, options)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, expectedAddress, tx.From)
		assert.Equal(t, tokenAddress, tx.To)
		assert.Equal(t, big.NewInt(0), tx.Value) // 0 ETH for token transfers
		assert.NotEmpty(t, tx.Data)              // Should contain the token transfer data
		assert.Equal(t, options.Nonce, tx.Nonce)
		assert.Equal(t, options.GasPrice, tx.GasPrice)
		assert.Equal(t, options.GasLimit, tx.GasLimit)
		assert.Equal(t, types.TransactionTypeERC20, tx.Type)
		assert.Equal(t, tokenAddress, tx.TokenAddress)

		keyStore.AssertExpectations(t)
	})

	t.Run("Create token transaction with default gas options", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		keyID := "test-key-id"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48" // USDC on Ethereum
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000) // 1 USDC (6 decimals)
		options := types.TransactionOptions{
			Nonce: 5,
		}

		// Generate a private key for testing
		privateKey, _ := createTestPrivateKey()
		publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)

		// Setup mock key and keystore
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: publicKey,
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create token transaction
		tx, err := wallet.CreateTokenTransaction(ctx, tokenAddress, toAddress, amount, options)

		assert.NoError(t, err)
		assert.NotNil(t, tx)
		assert.Equal(t, wallet.config.DefaultGasPrice, tx.GasPrice)
		assert.Equal(t, uint64(65000), tx.GasLimit) // Default for ERC20 transfers

		keyStore.AssertExpectations(t)
	})

	t.Run("Error with invalid token address", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		keyID := "test-key-id"
		tokenAddress := "invalid-token-address"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(1000000)
		options := types.TransactionOptions{}

		// Setup mock key and keystore
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: []byte{0x04, 1, 2, 3, 4}, // Dummy public key
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create token transaction
		tx, err := wallet.CreateTokenTransaction(ctx, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, types.ErrInvalidAddress)

		keyStore.AssertExpectations(t)
	})

	t.Run("Error with invalid recipient address", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		keyID := "test-key-id"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
		toAddress := "invalid-recipient-address"
		amount := big.NewInt(1000000)
		options := types.TransactionOptions{}

		// Setup mock key and keystore
		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: []byte{0x04, 1, 2, 3, 4}, // Dummy public key
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create token transaction
		tx, err := wallet.CreateTokenTransaction(ctx, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, types.ErrInvalidAddress)

		keyStore.AssertExpectations(t)
	})

	t.Run("Error with zero amount", func(t *testing.T) {
		ctx := context.Background()

		// Setup test data
		keyID := "test-key-id"
		tokenAddress := "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"
		toAddress := "0xbcd4042de499d14e55001ccbb24a551f3b954096"
		amount := big.NewInt(0)
		options := types.TransactionOptions{}

		// Setup mock for keystore with a valid public key
		privateKey, _ := createTestPrivateKey()
		publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)

		mockKey := &keystore.Key{
			ID:        keyID,
			PublicKey: publicKey,
		}
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", mock.Anything, keyID).Return(mockKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create token transaction
		tx, err := wallet.CreateTokenTransaction(ctx, tokenAddress, toAddress, amount, options)

		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorIs(t, err, types.ErrInvalidAmount)

		keyStore.AssertExpectations(t)
	})
}

// Test SignTransaction method
func TestEVMWallet_SignTransaction(t *testing.T) {
	t.Run("Error on address mismatch", func(t *testing.T) {
		ctx := context.Background()

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
		tx := &types.Transaction{
			Chain:    types.ChainTypeEthereum,
			From:     "0xDifferentAddress",
			To:       "0xbcd4042de499d14e55001ccbb24a551f3b954096",
			Value:    big.NewInt(1000000000000000000), // 1 ETH
			Nonce:    1,
			GasPrice: big.NewInt(20000000000),
			GasLimit: 21000,
			Type:     types.TransactionTypeNative,
		}

		// Set up keystore mock
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", ctx, keyID).Return(testKey, nil)

		// Create wallet with keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Sign transaction should fail due to address mismatch
		sig, err := wallet.SignTransaction(ctx, tx)

		assert.Error(t, err)
		assert.Nil(t, sig)
		assert.Contains(t, err.Error(), "transaction from address does not match key")
	})

	t.Run("Return error for non-existent key", func(t *testing.T) {
		ctx := context.Background()

		// Set up keystore mock to return error
		keyID := "non-existent-key"
		keyStore := new(MockKeyStore)
		appConfig := createTestConfig()
		keyStore.On("GetPublicKey", ctx, keyID).Return((*keystore.Key)(nil), keystore.ErrKeyNotFound)

		// Create wallet with the keyID
		wallet, _ := NewEVMWallet(keyStore, types.ChainTypeEthereum, keyID, appConfig)

		// Create transaction
		tx := createTestTransaction()

		// Sign transaction
		signedTx, err := wallet.SignTransaction(ctx, tx)

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

		evmConfig, err := NewEVMConfig(types.ChainTypeEthereum, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, evmConfig)
		assert.Equal(t, big.NewInt(1), evmConfig.ChainID)
		assert.Equal(t, uint64(21000), evmConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(20), evmConfig.DefaultGasPrice)
	})

	t.Run("Create config for Polygon", func(t *testing.T) {
		appConfig := createTestConfig()

		polygonConfig, err := NewEVMConfig(types.ChainTypePolygon, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, polygonConfig)
		assert.Equal(t, big.NewInt(137), polygonConfig.ChainID)
		assert.Equal(t, uint64(21000), polygonConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(30), polygonConfig.DefaultGasPrice)
	})

	t.Run("Create config for Base", func(t *testing.T) {
		appConfig := createTestConfig()

		baseConfig, err := NewEVMConfig(types.ChainTypeBase, appConfig)

		assert.NoError(t, err)
		assert.NotNil(t, baseConfig)
		assert.Equal(t, big.NewInt(8453), baseConfig.ChainID)
		assert.Equal(t, uint64(21000), baseConfig.DefaultGasLimit)
		assert.Equal(t, big.NewInt(10), baseConfig.DefaultGasPrice)
	})

	t.Run("Panic when appConfig is nil", func(t *testing.T) {
		assert.Panics(t, func() {
			_, _ = NewEVMConfig(types.ChainTypeEthereum, nil)
		})
	})

	t.Run("Error for unsupported chain", func(t *testing.T) {
		appConfig := createTestConfig()

		unsupportedConfig, err := NewEVMConfig(types.ChainType("unsupported"), appConfig)

		assert.Error(t, err)
		assert.Nil(t, unsupportedConfig)
		assert.ErrorIs(t, err, types.ErrUnsupportedChain)
	})

	t.Run("Error when blockchain config is nil", func(t *testing.T) {
		mockConfig := &MockConfigWithoutChainID{Config: createTestConfig()}

		config, err := NewEVMConfig(types.ChainTypeEthereum, mockConfig)

		assert.Error(t, err)
		assert.Nil(t, config)
		assert.ErrorIs(t, err, types.ErrUnsupportedChain)
	})
}
