package wallet

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"vault0/internal/config"
	"vault0/internal/keystore"
)

// MockKeyStore is a mock implementation of the KeyStore interface for testing
type MockKeyStore struct {
	mock.Mock
}

func (m *MockKeyStore) Create(ctx context.Context, id, name string, keyType keystore.KeyType, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, id, name, keyType, tags)
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Import(ctx context.Context, id, name string, keyType keystore.KeyType, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	args := m.Called(ctx, id, name, keyType, privateKey, publicKey, tags)
	return args.Get(0).(*keystore.Key), args.Error(1)
}

func (m *MockKeyStore) Sign(ctx context.Context, id string, data []byte) ([]byte, error) {
	args := m.Called(ctx, id, data)
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

func (m *MockWallet) ChainType() ChainType {
	args := m.Called()
	return args.Get(0).(ChainType)
}

func (m *MockWallet) DeriveAddress(ctx context.Context, publicKey []byte) (string, error) {
	args := m.Called(ctx, publicKey)
	return args.String(0), args.Error(1)
}

func (m *MockWallet) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockWallet) GetTokenBalance(ctx context.Context, address, tokenAddress string) (*big.Int, error) {
	args := m.Called(ctx, address, tokenAddress)
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockWallet) SendNative(ctx context.Context, keyID, toAddress string, amount *big.Int, options *TransactionOptions) (*Transaction, error) {
	args := m.Called(ctx, keyID, toAddress, amount, options)
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *MockWallet) SendToken(ctx context.Context, keyID, tokenAddress, toAddress string, amount *big.Int, options *TransactionOptions) (*Transaction, error) {
	args := m.Called(ctx, keyID, tokenAddress, toAddress, amount, options)
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *MockWallet) SignTransaction(ctx context.Context, keyID string, tx *Transaction) ([]byte, error) {
	args := m.Called(ctx, keyID, tx)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockWallet) BroadcastTransaction(ctx context.Context, signedTx []byte) (*Transaction, error) {
	args := m.Called(ctx, signedTx)
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *MockWallet) GetTransaction(ctx context.Context, hash string) (*Transaction, error) {
	args := m.Called(ctx, hash)
	return args.Get(0).(*Transaction), args.Error(1)
}

// Helper functions for tests
func createTestPrivateKey() (*ecdsa.PrivateKey, error) {
	return crypto.GenerateKey()
}

func createTestAddress() (common.Address, *ecdsa.PrivateKey, error) {
	privateKey, err := createTestPrivateKey()
	if err != nil {
		return common.Address{}, nil, err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, nil, err
	}

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, privateKey, nil
}

func createTestConfig() *config.Config {
	return &config.Config{
		Blockchains: config.BlockchainsConfig{
			Ethereum: config.BlockchainConfig{
				RPCURL:          "https://eth-mainnet.alchemyapi.io/v2/test-key",
				ChainID:         1,
				DefaultGasPrice: 20,
				DefaultGasLimit: 21000,
			},
			Polygon: config.BlockchainConfig{
				RPCURL:          "https://polygon-mainnet.g.alchemy.com/v2/test-key",
				ChainID:         137,
				DefaultGasPrice: 30,
				DefaultGasLimit: 21000,
			},
			Base: config.BlockchainConfig{
				RPCURL:          "https://mainnet.base.org",
				ChainID:         8453,
				DefaultGasPrice: 10,
				DefaultGasLimit: 21000,
			},
		},
	}
}

func createTestTransaction() *Transaction {
	return &Transaction{
		Chain:     ChainTypeEthereum,
		Hash:      "0x123456789abcdef",
		From:      "0x71C7656EC7ab88b098defB751B7401B5f6d8976F",
		To:        "0xbcd4042de499d14e55001ccbb24a551f3b954096",
		Value:     big.NewInt(1000000000000000000), // 1 ETH
		Data:      []byte{},
		Nonce:     0,
		GasPrice:  big.NewInt(20000000000), // 20 Gwei
		GasLimit:  21000,
		Type:      TransactionTypeNative,
		Status:    "pending",
		Timestamp: time.Now().Unix(),
	}
}

// Test cases

func TestChainType(t *testing.T) {
	tests := []struct {
		name      string
		chainType ChainType
		expected  string
	}{
		{
			name:      "Ethereum Chain Type",
			chainType: ChainTypeEthereum,
			expected:  "ethereum",
		},
		{
			name:      "Polygon Chain Type",
			chainType: ChainTypePolygon,
			expected:  "polygon",
		},
		{
			name:      "Base Chain Type",
			chainType: ChainTypeBase,
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
		transactionType TransactionType
		expected        string
	}{
		{
			name:            "Native Transaction Type",
			transactionType: TransactionTypeNative,
			expected:        "native",
		},
		{
			name:            "ERC20 Transaction Type",
			transactionType: TransactionTypeERC20,
			expected:        "erc20",
		},
		{
			name:            "Contract Transaction Type",
			transactionType: TransactionTypeContract,
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
		assert.Equal(t, ChainTypeEthereum, tx.Chain)
		assert.Equal(t, "0x123456789abcdef", tx.Hash)
		assert.Equal(t, "0x71C7656EC7ab88b098defB751B7401B5f6d8976F", tx.From)
		assert.Equal(t, "0xbcd4042de499d14e55001ccbb24a551f3b954096", tx.To)
		assert.Equal(t, big.NewInt(1000000000000000000), tx.Value)
		assert.Equal(t, []byte{}, tx.Data)
		assert.Equal(t, uint64(0), tx.Nonce)
		assert.Equal(t, big.NewInt(20000000000), tx.GasPrice)
		assert.Equal(t, uint64(21000), tx.GasLimit)
		assert.Equal(t, TransactionTypeNative, tx.Type)
		assert.Equal(t, "pending", tx.Status)
		assert.NotZero(t, tx.Timestamp)
	})
}

func TestTransactionOptions(t *testing.T) {
	t.Run("Create and validate TransactionOptions struct", func(t *testing.T) {
		// Create transaction options
		nonce := uint64(5)
		options := &TransactionOptions{
			GasPrice: big.NewInt(25000000000), // 25 Gwei
			GasLimit: 30000,
			Nonce:    &nonce,
			Data:     []byte{1, 2, 3, 4},
		}

		// Validate fields
		assert.Equal(t, big.NewInt(25000000000), options.GasPrice)
		assert.Equal(t, uint64(30000), options.GasLimit)
		assert.Equal(t, &nonce, options.Nonce)
		assert.Equal(t, []byte{1, 2, 3, 4}, options.Data)
	})

	t.Run("Create TransactionOptions with nil Nonce", func(t *testing.T) {
		options := &TransactionOptions{
			GasPrice: big.NewInt(20000000000),
			GasLimit: 21000,
			Nonce:    nil,
			Data:     []byte{},
		}

		assert.Nil(t, options.Nonce)
		assert.Equal(t, big.NewInt(20000000000), options.GasPrice)
	})
}
