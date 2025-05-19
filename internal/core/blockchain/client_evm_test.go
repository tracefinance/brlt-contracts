package blockchain

import (
	"context"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vault0/internal/errors"
	"vault0/internal/testing/mocks"
	"vault0/internal/types"
)

// TestNewEVMBlockchainClient tests the blockchain client creation
func TestNewEVMBlockchainClient(t *testing.T) {
	// Create mocks
	mockEth := mocks.NewMockEthClient()
	mockLogger := mocks.NewNopLogger()

	// Create test chain
	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
		ID:   1,
	}

	// Create the client
	client, err := NewEVMBlockchainClient(testChain, mockEth, mockLogger)

	// Verify
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, testChain, client.Chain())
}

// TestEVMClient_GetChainID tests the GetChainID method
func TestEVMClient_GetChainID(t *testing.T) {
	// Create mocks
	mockEth := mocks.NewMockEthClient()
	mockLogger := mocks.NewNopLogger()

	// Set up the expected return value
	expectedChainID := big.NewInt(1)
	ctx := context.Background()

	// Set up the mock to return the expected value
	mockEth.On("ChainID", ctx).Return(expectedChainID, nil)

	// Create test chain
	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
		ID:   1,
	}

	// Create the client
	client, err := NewEVMBlockchainClient(testChain, mockEth, mockLogger)
	require.NoError(t, err)

	// Test the method
	chainID, err := client.GetChainID(ctx)

	// Assert the results
	require.NoError(t, err)
	assert.Equal(t, expectedChainID.Int64(), chainID)
	mockEth.AssertExpectations(t)
}

// createTestEVMClient is a helper to create a test EVMClient with mock dependencies
func createTestEVMClient(t *testing.T) (*EVMClient, *mocks.MockEthClient) {
	mockEth := mocks.NewMockEthClient()
	mockLogger := mocks.NewNopLogger()

	// Create test chain
	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
		ID:   1,
	}

	// Create the client
	client, err := NewEVMBlockchainClient(testChain, mockEth, mockLogger)
	require.NoError(t, err)

	return client, mockEth
}

// TestEVMClient_GetBalance tests the GetBalance method
func TestEVMClient_GetBalance(t *testing.T) {
	ctx := context.Background()

	// Test address and balance
	testAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	expectedBalance := big.NewInt(100000000000000000) // 0.1 ETH

	// Create client with mocks
	client, mockEth := createTestEVMClient(t)

	// Set up the mock to return the expected value
	mockEth.On("BalanceAt", ctx, common.HexToAddress(testAddress), (*big.Int)(nil)).Return(expectedBalance, nil)

	// Test valid address
	balance, err := client.GetBalance(ctx, testAddress)
	require.NoError(t, err)
	assert.Equal(t, expectedBalance, balance)

	// Test invalid address
	balance, err = client.GetBalance(ctx, "invalid-address")
	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.True(t, errors.IsError(err, errors.ErrCodeInvalidAddress))

	mockEth.AssertExpectations(t)
}

// TestEVMClient_GetTokenBalance tests the GetTokenBalance method
func TestEVMClient_GetTokenBalance(t *testing.T) {
	ctx := context.Background()

	// Test addresses and balance
	testAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	testTokenAddress := "0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984" // UNI token
	expectedBalance := big.NewInt(500000000000000000)                // 0.5 UNI

	// Create client with mocks
	client, mockEth := createTestEVMClient(t)

	// Set up the mock to return the expected value for contract call
	mockCallResult := common.LeftPadBytes(expectedBalance.Bytes(), 32)
	mockEth.On("CallContract", ctx,
		mock.MatchedBy(func(msg ethereum.CallMsg) bool {
			return msg.To != nil && msg.To.Hex() == testTokenAddress
		}),
		(*big.Int)(nil)).Return(mockCallResult, nil)

	// Test valid addresses
	balance, err := client.GetTokenBalance(ctx, testAddress, testTokenAddress)
	require.NoError(t, err)
	assert.Equal(t, expectedBalance.String(), balance.String())

	// Test invalid address
	balance, err = client.GetTokenBalance(ctx, "invalid-address", testTokenAddress)
	assert.Error(t, err)
	assert.Nil(t, balance)
	assert.True(t, errors.IsError(err, errors.ErrCodeInvalidAddress))

	mockEth.AssertExpectations(t)
}

// TestEVMClient_GetNonce tests the GetNonce method
func TestEVMClient_GetNonce(t *testing.T) {
	ctx := context.Background()

	// Test address and expected nonce
	testAddress := "0x71C7656EC7ab88b098defB751B7401B5f6d8976F"
	expectedNonce := uint64(42)

	// Create client with mocks
	client, mockEth := createTestEVMClient(t)

	// Set up the mock to return the expected value
	mockEth.On("PendingNonceAt", ctx, common.HexToAddress(testAddress)).Return(expectedNonce, nil)

	// Test valid address
	nonce, err := client.GetNonce(ctx, testAddress)
	require.NoError(t, err)
	assert.Equal(t, expectedNonce, nonce)

	// Test invalid address
	nonce, err = client.GetNonce(ctx, "invalid-address")
	assert.Error(t, err)
	assert.Zero(t, nonce)
	assert.True(t, errors.IsError(err, errors.ErrCodeInvalidAddress))

	mockEth.AssertExpectations(t)
}

// TestEVMClient_GetGasPrice tests the GetGasPrice method
func TestEVMClient_GetGasPrice(t *testing.T) {
	ctx := context.Background()

	// Test data
	expectedGasPrice := big.NewInt(20000000000) // 20 Gwei

	// Create client with mocks
	client, mockEth := createTestEVMClient(t)

	// Set up the mock to return the expected value
	mockEth.On("SuggestGasPrice", ctx).Return(expectedGasPrice, nil).Once()

	// Test success case
	gasPrice, err := client.GetGasPrice(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedGasPrice, gasPrice)

	// Test RPC error
	mockEth.On("SuggestGasPrice", ctx).Return(nil, ethereum.NotFound).Once()
	gasPrice, err = client.GetGasPrice(ctx)
	assert.Error(t, err)
	assert.Nil(t, gasPrice)
	assert.True(t, errors.IsError(err, errors.ErrCodeRPCError), "Error should be an RPCError")

	mockEth.AssertExpectations(t)
}

// TestEVMClient_GetTransaction tests the GetTransaction method
func TestEVMClient_GetTransaction(t *testing.T) {
	ctx := context.Background()

	// Test transaction hash
	testTxHash := "0x5e9c12e53471f5fdf47faea193897695c32449f249ca548de18b2a42c297aa19"

	// Create client with mocks
	client, mockEth := createTestEVMClient(t)

	// Create a mock transaction
	mockTx := ethTypes.NewTransaction(
		1,                                  // nonce
		common.HexToAddress("0xrecipient"), // to
		big.NewInt(1000000000000000000),    // amount
		21000,                              // gas limit
		big.NewInt(20000000000),            // gas price
		[]byte("data"),                     // data
	)

	// Create a mock receipt
	mockReceipt := &ethTypes.Receipt{
		Status:            1,
		CumulativeGasUsed: 21000,
		GasUsed:           21000,
		BlockNumber:       big.NewInt(12345),
		BlockHash:         common.HexToHash("0xblockhash"),
		TxHash:            mockTx.Hash(),
	}

	// Create a mock block
	mockHeader := &ethTypes.Header{
		Time: 1620000000, // Unix timestamp
	}
	mockBlock := ethTypes.NewBlockWithHeader(mockHeader)

	// Set up the mock expectations
	mockEth.On("TransactionByHash", ctx, common.HexToHash(testTxHash)).Return(mockTx, false, nil)
	mockEth.On("TransactionReceipt", ctx, common.HexToHash(testTxHash)).Return(mockReceipt, nil)
	mockEth.On("BlockByNumber", ctx, mockReceipt.BlockNumber).Return(mockBlock, nil)

	// Not found case
	mockEth.On("TransactionByHash", ctx, common.HexToHash("0xnotfound")).Return(nil, false, ethereum.NotFound)

	// Test valid transaction hash
	tx, err := client.GetTransaction(ctx, testTxHash)
	require.NoError(t, err)
	assert.NotNil(t, tx)
	assert.Equal(t, mockTx.Hash().Hex(), tx.Hash)
	assert.Equal(t, types.ChainTypeEthereum, tx.ChainType)
	assert.Equal(t, int64(1620000000), tx.Timestamp)
	assert.Equal(t, big.NewInt(12345), tx.BlockNumber)
	assert.Equal(t, types.TransactionStatusSuccess, tx.Status)

	// Test transaction not found
	tx, err = client.GetTransaction(ctx, "0xnotfound")
	assert.Error(t, err)
	assert.Nil(t, tx)
	assert.True(t, errors.IsError(err, errors.ErrCodeTransactionNotFound))

	mockEth.AssertExpectations(t)
}

// TestEVMClient_Close tests the Close method
func TestEVMClient_Close(t *testing.T) {
	// Create client with mocks
	client, mockEth := createTestEVMClient(t)

	// Set up the mock
	mockEth.On("Close").Return()

	// Test
	client.Close()

	// Assert
	mockEth.AssertExpectations(t)
}
