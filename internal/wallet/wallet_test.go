package wallet

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vault0/internal/config"
	"vault0/internal/keygen"
	"vault0/internal/keystore"
	"vault0/internal/types"
)

// MockKeyStore is a mock implementation of the KeyStore interface for testing
type MockKeyStore struct {
	mock.Mock
}

func (m *MockKeyStore) Create(ctx context.Context, id, name string, keyType keygen.KeyType, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, id, name, keyType, tags)
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Import(ctx context.Context, id, name string, keyType keygen.KeyType, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, id, name, keyType, privateKey, publicKey, tags)
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Sign(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
	args := m.Called(ctx, id, data, dataType)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockKeyStore) GetPublicKey(ctx context.Context, id string) (*keystore.Key, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) List(ctx context.Context) ([]*keystore.Key, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, id, name, tags)
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockKeyStore) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockWallet is a mock implementation of the Wallet interface for testing
type MockWallet struct {
	mock.Mock
}

func (m *MockWallet) ChainType() types.ChainType {
	args := m.Called()
	return args.Get(0).(types.ChainType)
}

func (m *MockWallet) DeriveAddress(ctx context.Context, publicKey []byte) (string, error) {
	args := m.Called(ctx, publicKey)
	return args.String(0), args.Error(1)
}

func (m *MockWallet) CreateNativeTransaction(ctx context.Context, fromAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	args := m.Called(ctx, fromAddress, toAddress, amount, options)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWallet) CreateTokenTransaction(ctx context.Context, fromAddress, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	args := m.Called(ctx, fromAddress, tokenAddress, toAddress, amount, options)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWallet) SignTransaction(ctx context.Context, keyID string, tx *types.Transaction) ([]byte, error) {
	args := m.Called(ctx, keyID, tx)
	return args.Get(0).([]byte), args.Error(1)
}

// Helper functions for tests
func createTestPrivateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

func createTestConfig() *config.Config {
	return &config.Config{
		Blockchains: config.BlockchainsConfig{
			Ethereum: config.BlockchainConfig{
				RPCURL:          "https://mainnet.infura.io/v3/your-api-key",
				ChainID:         1,
				DefaultGasPrice: 20,
				DefaultGasLimit: 21000,
				ExplorerURL:     "https://etherscan.io",
			},
			Polygon: config.BlockchainConfig{
				RPCURL:          "https://polygon-mainnet.infura.io/v3/your-api-key",
				ChainID:         137,
				DefaultGasPrice: 30,
				DefaultGasLimit: 21000,
				ExplorerURL:     "https://polygonscan.com",
			},
			Base: config.BlockchainConfig{
				RPCURL:          "https://mainnet.base.org",
				ChainID:         8453,
				DefaultGasPrice: 10,
				DefaultGasLimit: 21000,
				ExplorerURL:     "https://basescan.org",
			},
		},
	}
}

func createTestTransaction() *types.Transaction {
	return &types.Transaction{
		Chain:     types.ChainTypeEthereum,
		Hash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		From:      "0xabcdef1234567890abcdef1234567890abcdef12",
		To:        "0x1234567890abcdef1234567890abcdef12345678",
		Value:     big.NewInt(100000000000000000), // 0.1 ETH
		Data:      []byte{},
		Nonce:     0,
		GasPrice:  big.NewInt(20000000000), // 20 Gwei
		GasLimit:  21000,
		Type:      types.TransactionTypeNative,
		Status:    "pending",
		Timestamp: time.Now().Unix(),
	}
}

// Test cases

func TestChainType(t *testing.T) {
	tests := []struct {
		name      string
		chainType types.ChainType
		expected  string
	}{
		{
			name:      "Ethereum Chain Type",
			chainType: types.ChainTypeEthereum,
			expected:  "ethereum",
		},
		{
			name:      "Polygon Chain Type",
			chainType: types.ChainTypePolygon,
			expected:  "polygon",
		},
		{
			name:      "Base Chain Type",
			chainType: types.ChainTypeBase,
			expected:  "base",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.chainType))
		})
	}
}

func TestTransactionType(t *testing.T) {
	tests := []struct {
		name            string
		transactionType types.TransactionType
		expected        string
	}{
		{
			name:            "Native Transaction Type",
			transactionType: types.TransactionTypeNative,
			expected:        "native",
		},
		{
			name:            "ERC20 Transaction Type",
			transactionType: types.TransactionTypeERC20,
			expected:        "erc20",
		},
		{
			name:            "Contract Transaction Type",
			transactionType: types.TransactionTypeContract,
			expected:        "contract",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.transactionType))
		})
	}
}

func TestTransaction(t *testing.T) {
	t.Run("Create and validate Transaction struct", func(t *testing.T) {
		// Create a test transaction
		tx := createTestTransaction()

		// Validate fields
		assert.Equal(t, types.ChainTypeEthereum, tx.Chain)
		assert.Equal(t, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", tx.Hash)
		assert.Equal(t, "0xabcdef1234567890abcdef1234567890abcdef12", tx.From)
		assert.Equal(t, "0x1234567890abcdef1234567890abcdef12345678", tx.To)
		assert.Equal(t, big.NewInt(100000000000000000), tx.Value)
		assert.Equal(t, []byte{}, tx.Data)
		assert.Equal(t, uint64(0), tx.Nonce)
		assert.Equal(t, big.NewInt(20000000000), tx.GasPrice)
		assert.Equal(t, uint64(21000), tx.GasLimit)
		assert.Equal(t, types.TransactionTypeNative, tx.Type)
		assert.Equal(t, "pending", tx.Status)
		assert.NotZero(t, tx.Timestamp)
	})
}

func TestTransactionOptions(t *testing.T) {
	t.Run("Create and validate TransactionOptions struct", func(t *testing.T) {
		// Create transaction options
		options := types.TransactionOptions{
			GasPrice: big.NewInt(25000000000), // 25 Gwei
			GasLimit: 30000,
			Nonce:    5,
			Data:     []byte{1, 2, 3, 4},
		}

		// Validate fields
		assert.Equal(t, big.NewInt(25000000000), options.GasPrice)
		assert.Equal(t, uint64(30000), options.GasLimit)
		assert.Equal(t, uint64(5), options.Nonce)
		assert.Equal(t, []byte{1, 2, 3, 4}, options.Data)
	})
}
