package blockchain

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vault0/internal/testing/mocks"
	"vault0/internal/types"
)

func TestContractMonitor_GetSubscription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		chainType      types.ChainType
		contractAddr   string
		setupMonitor   func(*ContractMonitor)
		expectNil      bool
		expectedEvents EventSet
	}{
		{
			name:         "non_existent_chain",
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			setupMonitor: func(m *ContractMonitor) {
				// No setup needed - empty monitor
			},
			expectNil: true,
		},
		{
			name:         "non_existent_contract",
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x456", EventSet{"Transfer(address,address,uint256)": struct{}{}})
			},
			expectNil: true,
		},
		{
			name:         "existing_contract",
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x123", EventSet{"Transfer(address,address,uint256)": struct{}{}})
			},
			expectNil:      false,
			expectedEvents: EventSet{"Transfer(address,address,uint256)": struct{}{}},
		},
		{
			name:         "case_insensitive_address",
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0xABC",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0xabC", EventSet{"Transfer(address,address,uint256)": struct{}{}})
			},
			expectNil:      false,
			expectedEvents: EventSet{"Transfer(address,address,uint256)": struct{}{}},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewContractMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute
			sub := monitor.GetSubscription(tc.chainType, tc.contractAddr)

			// Verify
			if tc.expectNil {
				assert.Nil(t, sub)
			} else {
				require.NotNil(t, sub)
				assert.Equal(t, tc.chainType, sub.ChainType)
				assert.Equal(t, tc.expectedEvents, sub.Events)
			}
		})
	}
}

func TestContractMonitor_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		chainType    types.ChainType
		contractAddr string
		events       EventSet
	}{
		{
			name:         "add_new_contract",
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			events:       EventSet{"Transfer(address,address,uint256)": struct{}{}},
		},
		{
			name:         "update_existing_contract",
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			events:       EventSet{"Approval(address,address,uint256)": struct{}{}},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewContractMonitor(logger)

			// Execute
			monitor.Add(tc.chainType, tc.contractAddr, tc.events)

			// Verify
			sub := monitor.GetSubscription(tc.chainType, tc.contractAddr)
			require.NotNil(t, sub)
			assert.Equal(t, tc.chainType, sub.ChainType)
			assert.Equal(t, tc.events, sub.Events)
		})
	}
}

func TestContractMonitor_Remove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		setupMonitor  func(*ContractMonitor)
		removeChain   types.ChainType
		removeAddr    string
		checkChain    types.ChainType
		checkAddr     string
		expectPresent bool
	}{
		{
			name: "remove_existing_contract",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x123", EventSet{"Transfer(address,address,uint256)": struct{}{}})
			},
			removeChain:   types.ChainTypeEthereum,
			removeAddr:    "0x123",
			checkChain:    types.ChainTypeEthereum,
			checkAddr:     "0x123",
			expectPresent: false,
		},
		{
			name: "remove_nonexistent_contract",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x456", EventSet{"Transfer(address,address,uint256)": struct{}{}})
			},
			removeChain:   types.ChainTypeEthereum,
			removeAddr:    "0x123",
			checkChain:    types.ChainTypeEthereum,
			checkAddr:     "0x456",
			expectPresent: true,
		},
		{
			name: "case_insensitive_remove",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0xABC", EventSet{"Transfer(address,address,uint256)": struct{}{}})
			},
			removeChain:   types.ChainTypeEthereum,
			removeAddr:    "0xabc",
			checkChain:    types.ChainTypeEthereum,
			checkAddr:     "0xABC",
			expectPresent: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewContractMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute
			monitor.Remove(tc.removeChain, tc.removeAddr)

			// Verify
			sub := monitor.GetSubscription(tc.checkChain, tc.checkAddr)
			if tc.expectPresent {
				require.NotNil(t, sub)
			} else {
				assert.Nil(t, sub)
			}
		})
	}
}

func TestContractMonitor_IsMonitored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupMonitor func(*ContractMonitor)
		chainType    types.ChainType
		contractAddr string
		eventSig     string
		expected     bool
	}{
		{
			name: "event_is_monitored",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x123", EventSet{
					"Transfer(address,address,uint256)": struct{}{},
				})
			},
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			eventSig:     "Transfer(address,address,uint256)",
			expected:     true,
		},
		{
			name: "event_not_monitored",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x123", EventSet{
					"Transfer(address,address,uint256)": struct{}{},
				})
			},
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			eventSig:     "Approval(address,address,uint256)",
			expected:     false,
		},
		{
			name: "contract_not_monitored",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x456", EventSet{
					"Transfer(address,address,uint256)": struct{}{},
				})
			},
			chainType:    types.ChainTypeEthereum,
			contractAddr: "0x123",
			eventSig:     "Transfer(address,address,uint256)",
			expected:     false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewContractMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute & Verify
			result := monitor.IsMonitored(tc.chainType, tc.contractAddr, tc.eventSig)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestContractMonitor_GetSubscriptionsForChain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupMonitor func(*ContractMonitor)
		chainType    types.ChainType
		expectedLen  int
	}{
		{
			name: "no_subscriptions",
			setupMonitor: func(m *ContractMonitor) {
				// Empty monitor
			},
			chainType:   types.ChainTypeEthereum,
			expectedLen: 0,
		},
		{
			name: "multiple_subscriptions",
			setupMonitor: func(m *ContractMonitor) {
				m.Add(types.ChainTypeEthereum, "0x123", EventSet{"Event1": struct{}{}})
				m.Add(types.ChainTypeEthereum, "0x456", EventSet{"Event2": struct{}{}})
				m.Add(types.ChainTypePolygon, "0x789", EventSet{"Event3": struct{}{}})
			},
			chainType:   types.ChainTypeEthereum,
			expectedLen: 2,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewContractMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute
			subs := monitor.GetSubscriptionsForChain(tc.chainType)

			// Verify
			assert.Len(t, subs, tc.expectedLen)
		})
	}
}

func TestContractMonitor_CancelAllSubscriptions(t *testing.T) {
	t.Parallel()

	// Setup
	logger := mocks.NewNopLogger()
	monitor := NewContractMonitor(logger)

	// Add some subscriptions
	monitor.Add(types.ChainTypeEthereum, "0x123", EventSet{"Event1": struct{}{}})
	monitor.Add(types.ChainTypePolygon, "0x456", EventSet{"Event2": struct{}{}})

	// Set cancel functions for testing
	ctx, cancel1 := context.WithCancel(context.Background())
	defer cancel1() // Ensure cleanup

	monitor.mutex.Lock()
	if sub := monitor.subscriptions[types.ChainTypeEthereum]["0x123"]; sub != nil {
		sub.CancelFunc = cancel1
	}
	monitor.mutex.Unlock()

	// Execute
	monitor.CancelAllSubscriptions()

	// Verify all subscriptions were removed
	assert.Empty(t, monitor.GetSubscriptionsForChain(types.ChainTypeEthereum))
	assert.Empty(t, monitor.GetSubscriptionsForChain(types.ChainTypePolygon))

	// Verify context was canceled
	select {
	case <-ctx.Done():
		// Good, context was canceled
	default:
		t.Error("Context should have been canceled")
	}
}

func TestContractSubscription_AddEvent(t *testing.T) {
	t.Parallel()

	sub := &ContractSubscription{
		ChainType:    types.ChainTypeEthereum,
		ContractAddr: "0x123",
		Events:       make(EventSet),
	}

	// Test adding an event
	sub.AddEvent("Transfer(address,address,uint256)")

	// Verify event was added
	_, exists := sub.Events["Transfer(address,address,uint256)"]
	assert.True(t, exists)
}

func TestContractSubscription_RemoveEvent(t *testing.T) {
	t.Parallel()

	// Setup subscription with an event
	sub := &ContractSubscription{
		ChainType:    types.ChainTypeEthereum,
		ContractAddr: "0x123",
		Events:       EventSet{"Transfer(address,address,uint256)": struct{}{}},
	}

	// Test removing the event
	sub.RemoveEvent("Transfer(address,address,uint256)")

	// Verify event was removed
	_, exists := sub.Events["Transfer(address,address,uint256)"]
	assert.False(t, exists)
}

func TestContractSubscription_Cancel(t *testing.T) {
	t.Parallel()

	// Setup context and cancel function
	ctx, cancel := context.WithCancel(context.Background())

	// Create subscription with cancel function
	sub := &ContractSubscription{
		ChainType:    types.ChainTypeEthereum,
		ContractAddr: "0x123",
		Events:       make(EventSet),
		CancelFunc:   cancel,
	}

	// Call Cancel
	sub.Cancel()

	// Verify context was canceled
	select {
	case <-ctx.Done():
		// Good, context was canceled
	default:
		t.Error("Context should have been canceled")
	}
}
