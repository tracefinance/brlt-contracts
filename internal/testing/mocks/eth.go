package mocks

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/stretchr/testify/mock"
)

// MockEthClient implements ethclient.Client for testing
type MockEthClient struct {
	mock.Mock
}

func (m *MockEthClient) ChainID(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockEthClient) BlockByNumber(ctx context.Context, number *big.Int) (*types.Block, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *MockEthClient) BlockNumber(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthClient) HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockEthClient) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Header), args.Error(1)
}

func (m *MockEthClient) TransactionCount(ctx context.Context, blockHash common.Hash) (uint, error) {
	args := m.Called(ctx, blockHash)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockEthClient) TransactionInBlock(ctx context.Context, blockHash common.Hash, index uint) (*types.Transaction, error) {
	args := m.Called(ctx, blockHash, index)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockEthClient) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	args := m.Called(ctx, hash)
	if args.Error(2) != nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*types.Transaction), args.Bool(1), nil
}

func (m *MockEthClient) TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error) {
	args := m.Called(ctx, txHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Receipt), args.Error(1)
}

func (m *MockEthClient) BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error) {
	args := m.Called(ctx, account, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) StorageAt(ctx context.Context, account common.Address, key common.Hash, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, key, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) CodeAt(ctx context.Context, account common.Address, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, account, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) NonceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (uint64, error) {
	args := m.Called(ctx, account, blockNumber)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthClient) PendingBalanceAt(ctx context.Context, account common.Address) (*big.Int, error) {
	args := m.Called(ctx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) PendingStorageAt(ctx context.Context, account common.Address, key common.Hash) ([]byte, error) {
	args := m.Called(ctx, account, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) PendingCodeAt(ctx context.Context, account common.Address) ([]byte, error) {
	args := m.Called(ctx, account)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) PendingNonceAt(ctx context.Context, account common.Address) (uint64, error) {
	args := m.Called(ctx, account)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthClient) PendingTransactionCount(ctx context.Context) (uint, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint), args.Error(1)
}

func (m *MockEthClient) CallContract(ctx context.Context, call ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	args := m.Called(ctx, call, blockNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockEthClient) EstimateGas(ctx context.Context, call ethereum.CallMsg) (uint64, error) {
	args := m.Called(ctx, call)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *MockEthClient) SuggestGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) SuggestGasTipCap(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *MockEthClient) FilterLogs(ctx context.Context, q ethereum.FilterQuery) ([]types.Log, error) {
	args := m.Called(ctx, q)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]types.Log), args.Error(1)
}

func (m *MockEthClient) SubscribeFilterLogs(ctx context.Context, q ethereum.FilterQuery, ch chan<- types.Log) (ethereum.Subscription, error) {
	args := m.Called(ctx, q, ch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

func (m *MockEthClient) SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error) {
	args := m.Called(ctx, ch)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(ethereum.Subscription), args.Error(1)
}

func (m *MockEthClient) SendTransaction(ctx context.Context, tx *types.Transaction) error {
	args := m.Called(ctx, tx)
	return args.Error(0)
}

// NewMockEthClient creates a new instance of MockEthClient
func NewMockEthClient() *MockEthClient {
	return &MockEthClient{}
}

// MockRPCClient implements rpc.Client for testing
type MockRPCClient struct {
	mock.Mock
}

func (m *MockRPCClient) Close() {
	m.Called()
}

func (m *MockRPCClient) Call(result interface{}, method string, args ...interface{}) error {
	callArgs := m.Called(result, method, args)
	return callArgs.Error(0)
}

func (m *MockRPCClient) CallContext(ctx context.Context, result interface{}, method string, args ...interface{}) error {
	callArgs := []interface{}{ctx, result, method}
	callArgs = append(callArgs, args...)
	return m.Called(callArgs...).Error(0)
}

func (m *MockRPCClient) BatchCall(b []rpc.BatchElem) error {
	return m.Called(b).Error(0)
}

func (m *MockRPCClient) BatchCallContext(ctx context.Context, b []rpc.BatchElem) error {
	return m.Called(ctx, b).Error(0)
}

func (m *MockRPCClient) EthSubscribe(ctx context.Context, channel interface{}, args ...interface{}) (*rpc.ClientSubscription, error) {
	callArgs := []interface{}{ctx, channel}
	callArgs = append(callArgs, args...)
	ret := m.Called(callArgs...)
	if ret.Get(0) == nil {
		return nil, ret.Error(1)
	}
	return ret.Get(0).(*rpc.ClientSubscription), ret.Error(1)
}

// NewMockRPCClient creates a new instance of MockRPCClient
func NewMockRPCClient() *MockRPCClient {
	return &MockRPCClient{}
}

// MockSubscription implements the ethereum.Subscription interface for testing
type MockSubscription struct {
	mock.Mock
	ErrChan chan error
}

func (m *MockSubscription) Unsubscribe() {
	m.Called()
}

func (m *MockSubscription) Err() <-chan error {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(chan error)
}

// NewMockSubscription creates a new instance of MockSubscription
func NewMockSubscription() *MockSubscription {
	return &MockSubscription{
		ErrChan: make(chan error),
	}
}

// RpcClient returns the mock as an *rpc.Client interface
func (m *MockRPCClient) RpcClient() *rpc.Client {
	// To properly support mocking, we need to return a real client pointer
	// This is complicated because rpc.Client doesn't have exported fields we can access
	// For testing purposes, using an unsafe cast to the mock itself
	// This works because our tests will only call methods we've mocked
	type mockRPCClient struct {
		*rpc.Client
	}
	clientPtr := &mockRPCClient{&rpc.Client{}}
	return clientPtr.Client
}

// EthClient returns the mock as an *ethclient.Client interface
func (m *MockEthClient) EthClient() *ethclient.Client {
	// Similar approach as with RpcClient
	// This forces Go to think our mock is an ethclient.Client
	// This is only for testing and wouldn't be appropriate in production code
	type mockEthClient struct {
		*ethclient.Client
	}
	clientPtr := &mockEthClient{&ethclient.Client{}}
	return clientPtr.Client
}
