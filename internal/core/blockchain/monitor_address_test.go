package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vault0/internal/testing/mocks"
	"vault0/internal/types"
)

func TestAddressMonitor_Add(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		addr          *types.Address
		setupMonitor  func(*AddressMonitor)
		expectError   bool
		errorContains string
	}{
		{
			name: "add_valid_address",
			addr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			setupMonitor: func(m *AddressMonitor) {},
			expectError:  false,
		},
		{
			name:          "add_nil_address",
			addr:          nil,
			setupMonitor:  func(m *AddressMonitor) {},
			expectError:   true,
			errorContains: "Address cannot be nil",
		},
		{
			name: "add_duplicate_address",
			addr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				})
			},
			expectError: false, // Should not error on duplicate
		},
		{
			name: "case_insensitive_add",
			addr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0xABCDEF0123456789ABCDEF0123456789ABCDEF01",
			},
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0xabcdef0123456789abcdef0123456789abcdef01",
				})
			},
			expectError: false, // Should handle different case
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewAddressMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute
			err := monitor.Add(tc.addr)

			// Verify
			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				if tc.addr != nil {
					// Verify address is now monitored
					isMonitored := monitor.IsMonitored(tc.addr.ChainType, []string{tc.addr.Address})
					assert.True(t, isMonitored)
				}
			}
		})
	}
}

func TestAddressMonitor_Remove(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupMonitor    func(*AddressMonitor)
		removeAddr      *types.Address
		checkAddr       *types.Address
		expectError     bool
		errorContains   string
		expectMonitored bool
	}{
		{
			name: "remove_existing_address",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				})
			},
			removeAddr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			checkAddr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			expectError:     false,
			expectMonitored: false,
		},
		{
			name:         "remove_nonexistent_address",
			setupMonitor: func(m *AddressMonitor) {},
			removeAddr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			checkAddr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0x1234567890123456789012345678901234567890",
			},
			expectError:     false, // No error on removing nonexistent
			expectMonitored: false,
		},
		{
			name: "case_insensitive_remove",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0xABCDEF0123456789ABCDEF0123456789ABCDEF01",
				})
			},
			removeAddr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0xabcdef0123456789abcdef0123456789abcdef01",
			},
			checkAddr: &types.Address{
				ChainType: types.ChainTypeEthereum,
				Address:   "0xABCDEF0123456789ABCDEF0123456789ABCDEF01",
			},
			expectError:     false,
			expectMonitored: false,
		},
		{
			name:          "remove_nil_address",
			setupMonitor:  func(m *AddressMonitor) {},
			removeAddr:    nil,
			expectError:   true,
			errorContains: "Address cannot be nil",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewAddressMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute
			err := monitor.Remove(tc.removeAddr)

			// Verify
			if tc.expectError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
			} else {
				require.NoError(t, err)
				if tc.checkAddr != nil {
					isMonitored := monitor.IsMonitored(tc.checkAddr.ChainType, []string{tc.checkAddr.Address})
					assert.Equal(t, tc.expectMonitored, isMonitored)
				}
			}
		})
	}
}

func TestAddressMonitor_IsMonitored(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		setupMonitor func(*AddressMonitor)
		chainType    types.ChainType
		addresses    []string
		expected     bool
	}{
		{
			name: "address_is_monitored",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				})
			},
			chainType: types.ChainTypeEthereum,
			addresses: []string{"0x1234567890123456789012345678901234567890"},
			expected:  true,
		},
		{
			name: "address_not_monitored",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x2234567890123456789012345678901234567890",
				})
			},
			chainType: types.ChainTypeEthereum,
			addresses: []string{"0x1234567890123456789012345678901234567890"},
			expected:  false,
		},
		{
			name: "chain_not_monitored",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				})
			},
			chainType: types.ChainTypePolygon,
			addresses: []string{"0x1234567890123456789012345678901234567890"},
			expected:  false,
		},
		{
			name: "one_of_many_monitored",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				})
			},
			chainType: types.ChainTypeEthereum,
			addresses: []string{
				"0x2234567890123456789012345678901234567890",
				"0x1234567890123456789012345678901234567890",
				"0x3234567890123456789012345678901234567890",
			},
			expected: true, // Only one address needs to be monitored
		},
		{
			name: "case_insensitive_check",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0xABCDEF0123456789ABCDEF0123456789ABCDEF01",
				})
			},
			chainType: types.ChainTypeEthereum,
			addresses: []string{"0xabcdef0123456789abcdef0123456789abcdef01"},
			expected:  true,
		},
		{
			name: "empty_address_list",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{
					ChainType: types.ChainTypeEthereum,
					Address:   "0x1234567890123456789012345678901234567890",
				})
			},
			chainType: types.ChainTypeEthereum,
			addresses: []string{},
			expected:  false, // No addresses to check
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewAddressMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute & Verify
			result := monitor.IsMonitored(tc.chainType, tc.addresses)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestAddressMonitor_GetAllAddresses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		setupMonitor   func(*AddressMonitor)
		chainType      types.ChainType
		expectedLength int
		expectedAddrs  []string
	}{
		{
			name: "no_addresses",
			setupMonitor: func(m *AddressMonitor) {
				// Empty monitor
			},
			chainType:      types.ChainTypeEthereum,
			expectedLength: 0,
			expectedAddrs:  []string{},
		},
		{
			name: "multiple_addresses_same_chain",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{ChainType: types.ChainTypeEthereum, Address: "0x1234567890123456789012345678901234567890"})
				_ = m.Add(&types.Address{ChainType: types.ChainTypeEthereum, Address: "0x2234567890123456789012345678901234567890"})
				_ = m.Add(&types.Address{ChainType: types.ChainTypePolygon, Address: "0x3234567890123456789012345678901234567890"})
			},
			chainType:      types.ChainTypeEthereum,
			expectedLength: 2,
			expectedAddrs:  []string{"0x1234567890123456789012345678901234567890", "0x2234567890123456789012345678901234567890"},
		},
		{
			name: "addresses_different_chain",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{ChainType: types.ChainTypeEthereum, Address: "0x1234567890123456789012345678901234567890"})
				_ = m.Add(&types.Address{ChainType: types.ChainTypePolygon, Address: "0x3234567890123456789012345678901234567890"})
			},
			chainType:      types.ChainTypePolygon,
			expectedLength: 1,
			expectedAddrs:  []string{"0x3234567890123456789012345678901234567890"},
		},
		{
			name: "nonexistent_chain",
			setupMonitor: func(m *AddressMonitor) {
				_ = m.Add(&types.Address{ChainType: types.ChainTypeEthereum, Address: "0x1234567890123456789012345678901234567890"})
			},
			chainType:      types.ChainTypePolygon,
			expectedLength: 0,
			expectedAddrs:  []string{},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Setup
			logger := mocks.NewNopLogger()
			monitor := NewAddressMonitor(logger)
			tc.setupMonitor(monitor)

			// Execute
			addresses := monitor.GetAllAddresses(tc.chainType)

			// Verify
			assert.Len(t, addresses, tc.expectedLength)

			// Check that all expected addresses are present
			// (Ignoring order since it's not guaranteed in map iteration)
			for _, expected := range tc.expectedAddrs {
				assert.Contains(t, addresses, expected)
			}
		})
	}
}
