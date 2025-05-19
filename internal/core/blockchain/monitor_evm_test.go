package blockchain

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vault0/internal/errors"
	"vault0/internal/testing/mocks"
	"vault0/internal/types"
)

// setupTestEVMMonitor creates a test EVMMonitor with mocked dependencies
func setupTestEVMMonitor() (*EVMMonitor, *mocks.MockBlockchainClient, *mocks.NopLogger) {
	mockClient := mocks.NewMockBlockchainClient()
	mockLogger := mocks.NewNopLogger()

	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
	}

	mockClient.On("Chain").Return(testChain).Maybe()

	monitor := NewEVMMonitor(mockLogger, mockClient).(*EVMMonitor)
	return monitor, mockClient, mockLogger
}

func TestNewEVMMonitor(t *testing.T) {
	t.Parallel()

	// Setup
	mockClient := mocks.NewMockBlockchainClient()
	mockLogger := mocks.NewNopLogger()

	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
	}
	mockClient.On("Chain").Return(testChain).Maybe()

	// Execute
	monitor := NewEVMMonitor(mockLogger, mockClient)

	// Verify
	assert.NotNil(t, monitor)
	evmMonitor, ok := monitor.(*EVMMonitor)
	assert.True(t, ok)
	assert.NotNil(t, evmMonitor.transactionEvents)
	assert.NotNil(t, evmMonitor.addressMonitor)
	assert.NotNil(t, evmMonitor.contractMonitor)
	assert.NotNil(t, evmMonitor.eventHandlers)

	// Verify that event handlers were registered
	assert.Greater(t, len(evmMonitor.eventHandlers), 0)
	mockClient.AssertExpectations(t)
}

func TestEVMMonitor_MonitorAddress(t *testing.T) {
	t.Parallel()

	// Test cases
	tests := []struct {
		name      string
		address   *types.Address
		expectErr bool
		errorCode string
	}{
		{
			name: "Valid address",
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			expectErr: false,
		},
		{
			name:      "Nil address",
			address:   nil,
			expectErr: true,
			errorCode: errors.ErrCodeInvalidInput,
		},
		{
			name: "Invalid address",
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "invalid",
			},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			monitor, mockClient, _ := setupTestEVMMonitor()

			// Execute
			err := monitor.MonitorAddress(tc.address)

			// Verify
			if tc.expectErr {
				assert.Error(t, err)
				if tc.errorCode != "" {
					assert.True(t, errors.IsError(err, tc.errorCode), "Expected error code %s, got %v", tc.errorCode, err)
				}
			} else {
				assert.NoError(t, err)

				// Verify address is monitored
				if tc.address != nil {
					isMonitored := monitor.addressMonitor.IsMonitored(tc.address.ChainType, []string{tc.address.Address})
					assert.True(t, isMonitored, "Address should be monitored")
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestEVMMonitor_UnmonitorAddress(t *testing.T) {
	t.Parallel()

	// Test cases
	tests := []struct {
		name      string
		setup     func(*EVMMonitor)
		address   *types.Address
		expectErr bool
		errorCode string
	}{
		{
			name: "Remove existing address",
			setup: func(m *EVMMonitor) {
				addr := &types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				}
				_ = m.MonitorAddress(addr)
			},
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			expectErr: false,
		},
		{
			name:  "Remove non-existing address",
			setup: func(m *EVMMonitor) {},
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			expectErr: false, // No error, just a no-op
		},
		{
			name:      "Nil address",
			setup:     func(m *EVMMonitor) {},
			address:   nil,
			expectErr: true,
			errorCode: errors.ErrCodeInvalidInput,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			monitor, mockClient, _ := setupTestEVMMonitor()
			tc.setup(monitor)

			// Execute
			err := monitor.UnmonitorAddress(tc.address)

			// Verify
			if tc.expectErr {
				assert.Error(t, err)
				if tc.errorCode != "" {
					assert.True(t, errors.IsError(err, tc.errorCode), "Expected error code %s, got %v", tc.errorCode, err)
				}
			} else {
				assert.NoError(t, err)

				// Verify address is not monitored
				if tc.address != nil {
					isMonitored := monitor.addressMonitor.IsMonitored(tc.address.ChainType, []string{tc.address.Address})
					assert.False(t, isMonitored, "Address should not be monitored")
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestEVMMonitor_MonitorContractAddress(t *testing.T) {
	t.Parallel()

	// Test cases
	tests := []struct {
		name      string
		address   *types.Address
		events    []string
		withCtx   bool // If true, set up an event context
		expectErr bool
		errorCode string
	}{
		{
			name: "Valid contract address and events",
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			events:    []string{string(types.ERC20TransferEvent)},
			expectErr: false,
		},
		{
			name: "Valid contract address and events with context",
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			events:    []string{string(types.ERC20TransferEvent)},
			withCtx:   true,
			expectErr: false,
		},
		{
			name:      "Nil address",
			address:   nil,
			events:    []string{string(types.ERC20TransferEvent)},
			expectErr: true,
			errorCode: errors.ErrCodeInvalidInput,
		},
		{
			name: "Invalid address",
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "invalid",
			},
			events:    []string{string(types.ERC20TransferEvent)},
			expectErr: true,
		},
		{
			name: "Empty events list",
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			events:    []string{},
			expectErr: true,
			errorCode: errors.ErrCodeInvalidInput,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			monitor, mockClient, _ := setupTestEVMMonitor()

			if tc.withCtx {
				// Set up context if needed
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				monitor.eventCtx = ctx
				monitor.eventCancel = cancel

				if tc.address != nil && !tc.expectErr {
					// Create the channels that will be returned by mock
					logCh := make(chan types.Log)
					logChTyped := (<-chan types.Log)(logCh)
					errCh := make(chan error)
					errChTyped := (<-chan error)(errCh)

					// Use mock.Anything for all parameters that might vary
					mockClient.On("SubscribeContractLogs",
						mock.Anything,
						[]string{tc.address.Address},
						mock.Anything,
						mock.Anything,
						int64(0)).
						Return(logChTyped, errChTyped, nil).Maybe()
				}
			}

			// Execute
			err := monitor.MonitorContractAddress(tc.address, tc.events)

			// Verify
			if tc.expectErr {
				assert.Error(t, err)
				if tc.errorCode != "" {
					assert.True(t, errors.IsError(err, tc.errorCode), "Expected error code %s, got %v", tc.errorCode, err)
				}
			} else {
				assert.NoError(t, err)

				// Verify contract is monitored
				if tc.address != nil && len(tc.events) > 0 {
					sub := monitor.contractMonitor.GetSubscription(tc.address.ChainType, tc.address.Address)
					assert.NotNil(t, sub, "Contract subscription should exist")
					assert.Contains(t, sub.Events, tc.events[0], "Event should be in subscription")
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestEVMMonitor_UnmonitorContractAddress(t *testing.T) {
	t.Parallel()

	// Test cases
	tests := []struct {
		name      string
		setup     func(*EVMMonitor)
		address   *types.Address
		expectErr bool
		errorCode string
	}{
		{
			name: "Remove existing contract",
			setup: func(m *EVMMonitor) {
				addr := &types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				}
				_ = m.MonitorContractAddress(addr, []string{string(types.ERC20TransferEvent)})
			},
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			expectErr: false,
		},
		{
			name:  "Remove non-existing contract",
			setup: func(m *EVMMonitor) {},
			address: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			expectErr: false, // No error, just a no-op
		},
		{
			name:      "Nil address",
			setup:     func(m *EVMMonitor) {},
			address:   nil,
			expectErr: true,
			errorCode: errors.ErrCodeInvalidInput,
		},
	}

	for _, tc := range tests {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			monitor, mockClient, _ := setupTestEVMMonitor()
			tc.setup(monitor)

			// Execute
			err := monitor.UnmonitorContractAddress(tc.address)

			// Verify
			if tc.expectErr {
				assert.Error(t, err)
				if tc.errorCode != "" {
					assert.True(t, errors.IsError(err, tc.errorCode), "Expected error code %s, got %v", tc.errorCode, err)
				}
			} else {
				assert.NoError(t, err)

				// Verify contract is not monitored
				if tc.address != nil {
					sub := monitor.contractMonitor.GetSubscription(tc.address.ChainType, tc.address.Address)
					assert.Nil(t, sub, "Contract subscription should not exist")
				}
			}

			mockClient.AssertExpectations(t)
		})
	}
}

func TestEVMMonitor_SubscribeToTransactionEvents(t *testing.T) {
	// This test is simplified since the function starts a goroutine that would run indefinitely
	// Setup
	monitor, mockClient, _ := setupTestEVMMonitor()

	// Create test channels
	blockCh := make(chan types.Block)
	blockChTyped := (<-chan types.Block)(blockCh)
	errCh := make(chan error)
	errChTyped := (<-chan error)(errCh)

	// We need a temporary context we can control
	ctx, cancel := context.WithCancel(context.Background())

	// Setup blockchain client mock expectations
	// Note: Since SubscribeToTransactionEvents starts a goroutine that never exits during the test,
	// we don't expect the mock call to be made in this test.
	// Instead, we just verify that it calls the correct context setting methods.
	mockClient.On("SubscribeNewHead", mock.Anything).Return(blockChTyped, errChTyped, nil).Maybe()

	// Execute - since this starts goroutines, immediately cancel to prevent hanging
	monitor.SubscribeToTransactionEvents(ctx)
	cancel() // Stop the goroutine right away

	// Verify context is set
	assert.NotNil(t, monitor.eventCtx)
	assert.NotNil(t, monitor.eventCancel)
}

func TestEVMMonitor_UnsubscribeFromTransactionEvents(t *testing.T) {
	// Setup
	monitor, mockClient, _ := setupTestEVMMonitor()

	// Set up a subscription first
	ctx, cancel := context.WithCancel(context.Background())
	monitor.eventCtx = ctx
	monitor.eventCancel = cancel

	// Execute
	monitor.UnsubscribeFromTransactionEvents()

	// Verify cancel function was called (eventCancel should be nil)
	assert.Nil(t, monitor.eventCancel, "Event cancel function should be nil after unsubscribe")

	// Try reading from transactionEvents channel, it should be closed
	_, ok := <-monitor.transactionEvents
	assert.False(t, ok, "Transaction events channel should be closed")

	mockClient.AssertExpectations(t)
}

func TestEVMMonitor_processBlock(t *testing.T) {
	// Setup
	monitor, mockClient, _ := setupTestEVMMonitor()

	// Initialize the transaction events channel with a buffer
	monitor.transactionEvents = make(chan *types.Transaction, 10)

	// Explicitly set up the monitored addresses map directly
	// This bypasses the normal MonitorAddress method which might have issues in tests
	monitoredAddr := "0xbbb"
	normalizedAddr := strings.ToLower(monitoredAddr)

	// Initialize the address monitor's internal map
	monitor.addressMonitor.monitoredAddresses = map[types.ChainType]AddressSet{
		types.ChainTypeEthereum: {
			normalizedAddr: struct{}{},
		},
	}

	// Create a test block with proper timestamp and big.Int usage
	currentTime := time.Now()
	block := &types.Block{
		ChainType:        types.ChainTypeEthereum,
		Hash:             "0xblockhash",
		Number:           big.NewInt(100),
		Timestamp:        currentTime,
		TransactionCount: 2,
		Transactions: []*types.Transaction{
			{
				BaseTransaction: types.BaseTransaction{
					ChainType: types.ChainTypeEthereum,
					Hash:      "0xtx1",
					From:      "0xaaa",
					To:        monitoredAddr, // Use the same address format here
				},
			},
			{
				BaseTransaction: types.BaseTransaction{
					ChainType: types.ChainTypeEthereum,
					Hash:      "0xtx2",
					From:      "0xccc",
					To:        "0xddd",
				},
			},
		},
	}

	// Verify that the address is actually monitored before proceeding
	isMonitored := monitor.addressMonitor.IsMonitored(types.ChainTypeEthereum, []string{monitoredAddr})
	require.True(t, isMonitored, "Address should be monitored before running the test")

	// Execute the processBlock method
	monitor.processBlock(block)

	// Since transactions are emitted to a channel, we need to check if anything was sent
	// This should receive the first transaction since it involves the monitored address
	select {
	case receivedTx := <-monitor.transactionEvents:
		// Got a transaction - verify it's the right one
		assert.Equal(t, "0xtx1", receivedTx.Hash, "Should receive the transaction with the monitored address")
		assert.Equal(t, currentTime.Unix(), receivedTx.Timestamp, "Transaction timestamp should match block timestamp")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("No transaction received within timeout")
	}

	// Verify no more transactions were emitted
	select {
	case unexpectedTx := <-monitor.transactionEvents:
		t.Fatalf("Unexpected transaction received: %s", unexpectedTx.Hash)
	case <-time.After(100 * time.Millisecond):
		// Good - no more transactions received
	}

	mockClient.AssertExpectations(t)
}

func TestEVMMonitor_logBasicEvent(t *testing.T) {
	t.Parallel()

	// Setup
	monitor, _, _ := setupTestEVMMonitor()

	// Create test log
	testLog := types.Log{
		ChainType:       types.ChainTypeEthereum,
		Address:         "0x1234567890123456789012345678901234567890",
		TransactionHash: "0xtxhash",
		Topics:          []string{"0xtopic1", "0xtopic2"},
	}

	// Create event handler
	handler := monitor.logBasicEvent("TestEvent")

	// Execute - this just logs, no return value to test
	// Just verify it doesn't panic
	assert.NotPanics(t, func() {
		handler(context.Background(), testLog)
	})
}

func TestEVMMonitor_processContractEventLog(t *testing.T) {
	t.Parallel()

	// Setup
	monitor, _, _ := setupTestEVMMonitor()

	// Create a contract subscription
	contractAddr := "0x1234567890123456789012345678901234567890"
	addr, err := types.NewAddress(types.ChainTypeEthereum, contractAddr)
	require.NoError(t, err)

	// Add contract to monitoring
	eventType := string(types.ERC20TransferEvent)
	err = monitor.MonitorContractAddress(addr, []string{eventType})
	require.NoError(t, err)

	// Create a test log
	testLog := types.Log{
		ChainType:       types.ChainTypeEthereum,
		Address:         contractAddr,
		TransactionHash: "0xtxhash",
		Topics:          []string{eventType, "0xtopic2", "0xtopic3"},
	}

	// Mock the event handler
	handlerCalled := false
	monitor.eventHandlers[eventType] = func(ctx context.Context, log types.Log) {
		handlerCalled = true
		assert.Equal(t, testLog, log)
	}

	// Execute
	monitor.processContractEventLog(context.Background(), testLog, eventType)

	// Verify handler was called
	assert.True(t, handlerCalled, "Event handler should have been called")
}

func TestEVMMonitor_TransactionEvents(t *testing.T) {
	t.Parallel()

	// Setup
	monitor, _, _ := setupTestEVMMonitor()

	// Execute
	evtChan := monitor.TransactionEvents()

	// Verify - just check that the returned channel is from the monitor
	assert.NotNil(t, evtChan)
	// Don't compare the channels directly as they have different types in the interface
}
