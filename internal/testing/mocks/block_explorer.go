package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"vault0/internal/core/blockexplorer"
	coreerrors "vault0/internal/errors"
	"vault0/internal/types"
)

// MockBlockExplorer implements the blockexplorer.BlockExplorer interface for testing
type MockBlockExplorer struct {
	mock.Mock
}

func (m *MockBlockExplorer) GetTransactionHistory(ctx context.Context, address string, options blockexplorer.TransactionHistoryOptions, nextToken string) (*types.Page[types.CoreTransaction], error) {
	args := m.Called(ctx, address, options, nextToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Page[types.CoreTransaction]), args.Error(1)
}

func (m *MockBlockExplorer) GetTransactionByHash(ctx context.Context, hash string) (*types.Transaction, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *MockBlockExplorer) GetTransactionReceiptByHash(ctx context.Context, hash string) (*types.TransactionReceipt, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TransactionReceipt), args.Error(1)
}

func (m *MockBlockExplorer) GetContract(ctx context.Context, address string) (*blockexplorer.ContractInfo, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*blockexplorer.ContractInfo), args.Error(1)
}

func (m *MockBlockExplorer) GetTokenURL(address string) string {
	args := m.Called(address)
	return args.String(0)
}

func (m *MockBlockExplorer) Chain() types.Chain {
	args := m.Called()
	return args.Get(0).(types.Chain)
}

// WithSuccessfulContractFetch sets up a mock expectation for a successful GetContract call
func (m *MockBlockExplorer) WithSuccessfulContractFetch(address string, abi string) *MockBlockExplorer {
	m.On("GetContract", mock.Anything, address).Return(&blockexplorer.ContractInfo{
		ABI:          abi,
		ContractName: "TestContract",
		IsVerified:   true,
	}, nil)
	return m
}

// WithContractNotFound sets up a mock expectation for a GetContract call that fails with contract not found
func (m *MockBlockExplorer) WithContractNotFound(address string) *MockBlockExplorer {
	m.On("GetContract", mock.Anything, address).Return(nil, coreerrors.NewContractNotFoundError(address, "ethereum"))
	return m
}

// NewMockBlockExplorer creates and returns a new instance of MockBlockExplorer
// with common expectations already set up
func NewMockBlockExplorer() *MockBlockExplorer {
	mockExplorer := &MockBlockExplorer{}

	// Set up common expectations
	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
	}
	mockExplorer.On("Chain").Return(testChain).Maybe()

	return mockExplorer
}
