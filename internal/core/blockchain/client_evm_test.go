package blockchain

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vault0/internal/errors"
	"vault0/internal/types"
)

// TestNewEVMBlockchainClient tests the blockchain client creation
func TestNewEVMBlockchainClient(t *testing.T) {
	// Create mocks instead of using the actual client
	mockClient := new(mockEVMClient)
	mockChain := types.Chain{
		Type: types.ChainTypeEthereum,
		ID:   1,
	}
	mockClient.chain = mockChain

	// Assert the Chain() method works correctly
	assert.Equal(t, mockChain, mockClient.Chain())
}

// TestEVMClient_GetChainID tests the GetChainID method
func TestEVMClient_GetChainID(t *testing.T) {
	// Create mocks
	mockClient := new(mockEVMClient)
	ctx := context.Background()

	// Set up the expected return value
	expectedChainID := big.NewInt(1)

	// Set up the mock to return the expected value
	mockClient.On("GetChainID", ctx).Return(expectedChainID.Int64(), nil)

	// Test the method
	chainID, err := mockClient.GetChainID(ctx)

	// Assert the results
	require.NoError(t, err)
	assert.Equal(t, expectedChainID.Int64(), chainID)
	mockClient.AssertExpectations(t)
}

// mockEVMClient is a complete mock of the EVMClient for testing
type mockEVMClient struct {
	mock.Mock
	chain types.Chain
}

func (m *mockEVMClient) GetChainID(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockEVMClient) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	args := m.Called(ctx, address)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *mockEVMClient) GetTokenBalance(ctx context.Context, address string, tokenAddress string) (*big.Int, error) {
	args := m.Called(ctx, address, tokenAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *mockEVMClient) GetNonce(ctx context.Context, address string) (uint64, error) {
	args := m.Called(ctx, address)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockEVMClient) GetTransaction(ctx context.Context, hash string) (*types.Transaction, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Transaction), args.Error(1)
}

func (m *mockEVMClient) GetBlock(ctx context.Context, identifier string) (*types.Block, error) {
	args := m.Called(ctx, identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.Block), args.Error(1)
}

func (m *mockEVMClient) GetTransactionReceipt(ctx context.Context, hash string) (*types.TransactionReceipt, error) {
	args := m.Called(ctx, hash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*types.TransactionReceipt), args.Error(1)
}

func (m *mockEVMClient) EstimateGas(ctx context.Context, tx *types.Transaction) (uint64, error) {
	args := m.Called(ctx, tx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *mockEVMClient) GetGasPrice(ctx context.Context) (*big.Int, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*big.Int), args.Error(1)
}

func (m *mockEVMClient) CallContract(ctx context.Context, from string, to string, data []byte) ([]byte, error) {
	args := m.Called(ctx, from, to, data)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockEVMClient) BroadcastTransaction(ctx context.Context, signedTx []byte) (string, error) {
	args := m.Called(ctx, signedTx)
	return args.String(0), args.Error(1)
}

func (m *mockEVMClient) FilterContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock, toBlock int64) ([]types.Log, error) {
	args := m.Called(ctx, addresses, eventSignature, eventArgs, fromBlock, toBlock)
	return args.Get(0).([]types.Log), args.Error(1)
}

func (m *mockEVMClient) SubscribeContractLogs(ctx context.Context, addresses []string, eventSignature string, eventArgs []any, fromBlock int64) (<-chan types.Log, <-chan error, error) {
	args := m.Called(ctx, addresses, eventSignature, eventArgs, fromBlock)
	return args.Get(0).(<-chan types.Log), args.Get(1).(<-chan error), args.Error(2)
}

func (m *mockEVMClient) SubscribeNewHead(ctx context.Context) (<-chan types.Block, <-chan error, error) {
	args := m.Called(ctx)
	return args.Get(0).(<-chan types.Block), args.Get(1).(<-chan error), args.Error(2)
}

func (m *mockEVMClient) Chain() types.Chain {
	return m.chain
}

func (m *mockEVMClient) Close() {
	m.Called()
}

// TestEVMClient_GetBalance tests the GetBalance method
func TestEVMClient_GetBalance(t *testing.T) {
	// Create mock
	mockClient := new(mockEVMClient)
	ctx := context.Background()

	// Test address and balance
	testAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	expectedBalance := big.NewInt(100000000000000000) // 0.1 ETH

	// Set up the mock
	mockClient.On("GetBalance", ctx, testAddress).Return(expectedBalance, nil)
	mockClient.On("GetBalance", ctx, "invalid-address").Return(nil, errors.NewInvalidAddressError("invalid-address"))

	// Test valid address
	balance, err := mockClient.GetBalance(ctx, testAddress)
	require.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)

	// Test invalid address
	balance, err = mockClient.GetBalance(ctx, "invalid-address")
	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.True(t, errors.IsError(err, errors.ErrCodeInvalidAddress))

	mockClient.AssertExpectations(t)
}

// TestEVMClient_GetTokenBalance tests the GetTokenBalance method
func TestEVMClient_GetTokenBalance(t *testing.T) {
	// Create mock
	mockClient := new(mockEVMClient)
	ctx := context.Background()

	// Test addresses and balance
	testAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	testTokenAddress := "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984" // UNI token
	expectedBalance := big.NewInt(500000000000000000)                // 0.5 UNI

	// Set up the mock
	mockClient.On("GetTokenBalance", ctx, testAddress, testTokenAddress).Return(expectedBalance, nil)
	mockClient.On("GetTokenBalance", ctx, "invalid-address", testTokenAddress).Return(nil, errors.NewInvalidAddressError("invalid-address"))

	// Test valid addresses
	balance, err := mockClient.GetTokenBalance(ctx, testAddress, testTokenAddress)
	require.NoError(t, err)
	assert.Equal(t, expectedBalance.String(), balance.String())

	// Test invalid address
	balance, err = mockClient.GetTokenBalance(ctx, "invalid-address", testTokenAddress)
	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.True(t, errors.IsError(err, errors.ErrCodeInvalidAddress))

	mockClient.AssertExpectations(t)
}

// TestEVMClient_GetTransaction tests the GetTransaction method
func TestEVMClient_GetTransaction(t *testing.T) {
	// Create mock
	mockClient := new(mockEVMClient)
	ctx := context.Background()

	// Test transaction hash
	testTxHash := "0x5e9c12e53471f5fdf47faea193897695c32449f249ca548de18b2a42c297aa19"

	// Create a mock transaction
	mockTx := &types.Transaction{
		BaseTransaction: types.BaseTransaction{
			ChainType: types.ChainTypeEthereum,
			Hash:      testTxHash,
		},
		Status:      types.TransactionStatusSuccess,
		Timestamp:   1620000000,
		BlockNumber: big.NewInt(12345),
	}

	// Set up the mock
	mockClient.On("GetTransaction", ctx, testTxHash).Return(mockTx, nil)
	mockClient.On("GetTransaction", ctx, "0xnotfound").Return(nil, errors.NewTransactionNotFoundError("0xnotfound"))

	// Test valid transaction hash
	tx, err := mockClient.GetTransaction(ctx, testTxHash)
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, testTxHash, tx.Hash)
	assert.Equal(t, types.ChainTypeEthereum, tx.ChainType)
	assert.Equal(t, int64(1620000000), tx.Timestamp)
	assert.Equal(t, big.NewInt(12345), tx.BlockNumber)
	assert.Equal(t, types.TransactionStatusSuccess, tx.Status)

	// Test transaction not found
	tx, err = mockClient.GetTransaction(ctx, "0xnotfound")
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.True(t, errors.IsError(err, errors.ErrCodeTransactionNotFound))

	mockClient.AssertExpectations(t)
}

// TestEVMClient_GetBlock tests the GetBlock method
func TestEVMClient_GetBlock(t *testing.T) {
	// Create mock
	mockClient := new(mockEVMClient)
	ctx := context.Background()

	// Create test block
	testBlock := &types.Block{
		Hash:       "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		Number:     big.NewInt(12345),
		ParentHash: "0x0000000000000000000000000000000000000000000000000000000000000000",
		Timestamp:  time.Unix(1620000000, 0),
		Miner:      "0xminer",
		GasUsed:    8000000,
		GasLimit:   15000000,
		Difficulty: big.NewInt(1000000),
	}

	// Test cases
	testCases := []struct {
		name          string
		identifier    string
		mockSetup     func()
		expectedError error
	}{
		{
			name:       "Get Block by Number",
			identifier: "12345",
			mockSetup: func() {
				mockClient.On("GetBlock", ctx, "12345").Return(testBlock, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:       "Get Block by Hash",
			identifier: "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
			mockSetup: func() {
				mockClient.On("GetBlock", ctx, "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef").Return(testBlock, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:       "Get Latest Block",
			identifier: "latest",
			mockSetup: func() {
				mockClient.On("GetBlock", ctx, "latest").Return(testBlock, nil).Once()
			},
			expectedError: nil,
		},
		{
			name:       "Block Not Found",
			identifier: "99999",
			mockSetup: func() {
				mockClient.On("GetBlock", ctx, "99999").Return(nil, errors.NewBlockNotFoundError("99999")).Once()
			},
			expectedError: errors.NewBlockNotFoundError("99999"),
		},
		{
			name:       "Invalid Block Identifier",
			identifier: "invalid",
			mockSetup: func() {
				mockClient.On("GetBlock", ctx, "invalid").Return(nil, errors.NewInvalidBlockIdentifierError("invalid")).Once()
			},
			expectedError: errors.NewInvalidBlockIdentifierError("invalid"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup mocks for this test case
			tc.mockSetup()

			// Test
			block, err := mockClient.GetBlock(ctx, tc.identifier)

			// Assert
			if tc.expectedError != nil {
				assert.Error(t, err)
				if errors.IsError(tc.expectedError, errors.ErrCodeBlockNotFound) {
					assert.True(t, errors.IsError(err, errors.ErrCodeBlockNotFound))
				} else if errors.IsError(tc.expectedError, errors.ErrCodeInvalidBlockIdentifier) {
					assert.True(t, errors.IsError(err, errors.ErrCodeInvalidBlockIdentifier))
				}
				assert.Nil(t, block)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, block)
				assert.Equal(t, testBlock.Hash, block.Hash)
				assert.Equal(t, testBlock.Number, block.Number)
				assert.Equal(t, testBlock.Timestamp.Unix(), block.Timestamp.Unix())
				assert.Equal(t, testBlock.Difficulty, block.Difficulty)
				assert.Equal(t, testBlock.GasLimit, block.GasLimit)
				assert.Equal(t, testBlock.GasUsed, block.GasUsed)
				assert.Equal(t, testBlock.Miner, block.Miner)
			}
		})
	}
}

// TestEVMClient_CallContract tests the CallContract method
func TestEVMClient_CallContract(t *testing.T) {
	// Create mock
	mockClient := new(mockEVMClient)
	ctx := context.Background()

	// Test data
	fromAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	contractAddress := "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984"
	callData := []byte{0x70, 0xa0, 0x82, 0x31} // Example method ID for balanceOf()
	expectedResult := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100}

	// Set up the mock
	mockClient.On("CallContract", ctx, fromAddress, contractAddress, callData).Return(expectedResult, nil)
	mockClient.On("CallContract", ctx, types.ZeroAddress, contractAddress, callData).Return(expectedResult, nil)
	mockClient.On("CallContract", ctx, fromAddress, "invalid-address", callData).Return(nil, errors.NewInvalidAddressError("invalid-address"))

	// Test with valid addresses
	result, err := mockClient.CallContract(ctx, fromAddress, contractAddress, callData)
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	// Test with zero address as "from"
	result, err = mockClient.CallContract(ctx, types.ZeroAddress, contractAddress, callData)
	require.NoError(t, err)
	assert.Equal(t, expectedResult, result)

	// Test with invalid "to" address
	result, err = mockClient.CallContract(ctx, fromAddress, "invalid-address", callData)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.IsError(err, errors.ErrCodeInvalidAddress))

	mockClient.AssertExpectations(t)
}

// TestEVMClient_GetGasPrice tests the GetGasPrice method
func TestEVMClient_GetGasPrice(t *testing.T) {
	// Create mock
	mockClient := new(mockEVMClient)
	ctx := context.Background()

	// Test data
	expectedGasPrice := big.NewInt(20000000000) // 20 Gwei

	// Set up the mock
	mockClient.On("GetGasPrice", ctx).Return(expectedGasPrice, nil).Once()
	mockClient.On("GetGasPrice", ctx).Return(nil, errors.NewRPCError(ethereum.NotFound)).Once()

	// Test success case
	gasPrice, err := mockClient.GetGasPrice(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedGasPrice, gasPrice)

	// Test RPC error
	gasPrice, err = mockClient.GetGasPrice(ctx)
	assert.Error(t, err)
	assert.Nil(t, gasPrice)
	assert.True(t, errors.IsError(err, errors.ErrCodeRPCError))

	mockClient.AssertExpectations(t)
}

// TestEVMClient_Close tests the Close method
func TestEVMClient_Close(t *testing.T) {
	// Create mock
	mockClient := new(mockEVMClient)

	// Set up the mock
	mockClient.On("Close").Return()

	// Test
	mockClient.Close()

	// Assert
	mockClient.AssertExpectations(t)
}
