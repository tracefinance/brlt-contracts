package wallet

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/mock"

	"vault0/internal/types"
)

// MockWallet is a mock implementation of the Wallet interface for testing
type MockWallet struct {
	mock.Mock
}

func (m *MockWallet) Chain() types.Chain {
	args := m.Called()
	return args.Get(0).(types.Chain)
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
