package wallet

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/mock"

	"vault0/internal/config"
	"vault0/internal/types"
)

// MockWallet is a mock implementation of the Wallet interface for testing
type MockWallet struct {
	mock.Mock
}

func (m *MockWallet) ChainType() types.ChainType {
	args := m.Called()
	return args.Get(0).(types.ChainType)
}

func (m *MockWallet) DeriveAddress(ctx context.Context, keyID string) (string, error) {
	args := m.Called(ctx, keyID)
	return args.String(0), args.Error(1)
}

func (m *MockWallet) CreateNativeTransaction(ctx context.Context, keyID string, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	args := m.Called(ctx, keyID, toAddress, amount, options)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWallet) CreateTokenTransaction(ctx context.Context, keyID string, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	args := m.Called(ctx, keyID, tokenAddress, toAddress, amount, options)
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockWallet) SignTransaction(ctx context.Context, keyID string, tx *types.Transaction) ([]byte, error) {
	args := m.Called(ctx, keyID, tx)
	return args.Get(0).([]byte), args.Error(1)
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
