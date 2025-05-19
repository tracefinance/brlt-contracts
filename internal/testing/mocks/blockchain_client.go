package mocks

import (
	"context"
	"math/big"

	"github.com/stretchr/testify/mock"

	"vault0/internal/types"
)

// MockBlockchainClient implements the blockchain.BlockchainClient interface for testing
type MockBlockchainClient struct {
	mock.Mock
}

func (m *MockBlockchainClient) GetTransaction(ctx context.Context, hash string) (*types.Transaction, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBlockchainClient) GetBlock(ctx context.Context, identifier string) (*types.Block, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockBlockchainClient) GetTransactionReceipt(ctx context.Context, hash string) (*types.TransactionReceipt, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TransactionReceipt), args.Error(1)
}

func (m *MockBlockchainClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBlockchainClient) GetTokenBalance(ctx context.Context, address string, tokenAddress string) (*big.Int, error) {
	args := m.Called(ctx, address, tokenAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBlockchainClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockBlockchainClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockBlockchainClient) CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error) {
	args := m.Called(ctx, from, to, data)
	err := args.Error(1)
	if args.Get(0) == nil {
		return nil, err
	}
	return args.Get(0).([]byte), err
}

func (m *MockBlockchainClient) FilterContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock, toBlock int64) ([]types.Log, error) {
	args := m.Called(ctx, addresses, eventSignature, eventArgs, fromBlock, toBlock)
	return args.Get(0).([]types.Log), args.Error(1)
}

func (m *MockBlockchainClient) SubscribeContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock int64) (<-chan types.Log, <-chan error, error) {
	args := m.Called(ctx, addresses, eventSignature, eventArgs, fromBlock)
	return args.Get(0).(<-chan types.Log), args.Get(1).(<-chan error), args.Error(2)
}

func (m *MockBlockchainClient) BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error) {
	args := m.Called(ctx, signedTx)
	return args.String(0), args.Error(1)
}

func (m *MockBlockchainClient) EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockBlockchainClient) SubscribeNewHead(ctx context.Context) (<-chan types.Block, <-chan error, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan types.Block), args.Get(1).(<-chan error), args.Error(2)
}

func (m *MockBlockchainClient) Chain() types.Chain {
	args := m.Called()
	return args.Get(0).(types.Chain)
}

func (m *MockBlockchainClient) Close() {
	m.Called()
}

// NewMockBlockchainClient creates and returns a new instance of MockBlockchainClient
// with common expectations already set up
func NewMockBlockchainClient() *MockBlockchainClient {
	mockClient := &MockBlockchainClient{}

	// Set up common expectations
	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
	}
	mockClient.On("Chain").Return(testChain).Maybe()

	return mockClient
}
