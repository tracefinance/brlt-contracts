package abi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	coreerrors "vault0/internal/errors"
	"vault0/internal/testing/matchers"
	"vault0/internal/testing/mocks"
	"vault0/internal/types"
)

// setupTestABILoader creates a test ABILoader with mocked dependencies
func setupTestABILoader() (*abiLoader, *mocks.MockConfig, *mocks.MockBlockExplorer, *mocks.MockBlockchainClient, *mocks.MockABIUtils) {
	mockConfig := mocks.NewMockConfig()
	mockExplorer := mocks.NewMockBlockExplorer()
	mockBlockchainClient := mocks.NewMockBlockchainClient()
	mockABIUtils := mocks.NewMockABIUtils()
	mockLogger := mocks.NewNopLogger()

	testChain := types.Chain{
		Type: types.ChainTypeEthereum,
	}

	loader := &abiLoader{
		config:           mockConfig,
		explorer:         mockExplorer,
		blockchainClient: mockBlockchainClient,
		log:              mockLogger,
		chainType:        testChain.Type,
		abiCache:         &sync.Map{},
		abiUtils:         mockABIUtils,
	}

	return loader, mockConfig, mockExplorer, mockBlockchainClient, mockABIUtils
}

func createTestLoader(configProvider config.ABIConfigProvider) *abiLoader {
	return &abiLoader{
		config:   configProvider,
		abiCache: &sync.Map{},
	}
}

// TestLoadABIByType tests the LoadABIByType method using real files
func TestLoadABIByType(t *testing.T) {
	// Create temp directory for test artifacts
	tmpDir, err := os.MkdirTemp("", "abi-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test artifact files
	erc20ArtifactPath := filepath.Join(tmpDir, "erc20.json")
	erc20ArtifactContent := `{
		"abi": [
			{
				"constant": true,
				"inputs": [{"name": "_owner", "type": "address"}],
				"name": "balanceOf",
				"outputs": [{"name": "balance", "type": "uint256"}],
				"payable": false,
				"stateMutability": "view",
				"type": "function"
			}
		]
	}`
	err = os.WriteFile(erc20ArtifactPath, []byte(erc20ArtifactContent), 0644)
	require.NoError(t, err)

	// Invalid artifact (missing ABI)
	invalidArtifactPath := filepath.Join(tmpDir, "invalid.json")
	invalidArtifactContent := `{"name": "Invalid"}`
	err = os.WriteFile(invalidArtifactPath, []byte(invalidArtifactContent), 0644)
	require.NoError(t, err)

	// Invalid JSON artifact
	badJsonPath := filepath.Join(tmpDir, "badjson.json")
	badJsonContent := `{not valid json`
	err = os.WriteFile(badJsonPath, []byte(badJsonContent), 0644)
	require.NoError(t, err)

	// Create a test context
	ctx := context.Background()

	t.Run("LoadABI with valid ERC20 artifact", func(t *testing.T) {
		configWithValidPaths := &testConfig{
			paths: map[string]string{
				"erc20": erc20ArtifactPath,
			},
		}

		loader := createTestLoader(configWithValidPaths)

		// Test loading ERC20 ABI
		abi, err := loader.LoadABIByType(ctx, ABITypeERC20)
		assert.NoError(t, err)
		assert.Contains(t, abi, "balanceOf")
		assert.Contains(t, abi, "_owner")
		assert.Contains(t, abi, "address")

		// Verify caching works
		abi2, err := loader.LoadABIByType(ctx, ABITypeERC20)
		assert.NoError(t, err)
		assert.Equal(t, abi, abi2)
	})

	t.Run("LoadABI with invalid JSON", func(t *testing.T) {
		configWithInvalidJSON := &testConfig{
			paths: map[string]string{
				"erc20": badJsonPath,
			},
		}

		loader := createTestLoader(configWithInvalidJSON)

		// Test loading with invalid JSON
		_, err := loader.LoadABIByType(ctx, ABITypeERC20)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse artifact JSON")
	})

	t.Run("LoadABI with missing ABI field", func(t *testing.T) {
		configWithMissingABI := &testConfig{
			paths: map[string]string{
				"erc20": invalidArtifactPath,
			},
		}

		loader := createTestLoader(configWithMissingABI)

		// Test loading with missing ABI field
		_, err := loader.LoadABIByType(ctx, ABITypeERC20)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "abi not found")
	})

	t.Run("LoadABI with nonexistent file", func(t *testing.T) {
		configWithNonexistentFile := &testConfig{
			paths: map[string]string{
				"erc20": filepath.Join(tmpDir, "nonexistent.json"),
			},
		}

		loader := createTestLoader(configWithNonexistentFile)

		// Test loading with nonexistent file
		_, err := loader.LoadABIByType(ctx, ABITypeERC20)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("LoadABI with missing path in config", func(t *testing.T) {
		configWithMissingPath := &testConfig{
			paths: map[string]string{
				// no erc20 path
			},
			err: assert.AnError,
		}

		loader := createTestLoader(configWithMissingPath)

		// Test loading with missing path in config
		_, err := loader.LoadABIByType(ctx, ABITypeERC20)
		assert.Error(t, err)
	})
}

// Simple test implementations

type testConfig struct {
	paths map[string]string
	err   error
}

func (c *testConfig) GetArtifactPathForType(abiType string) (string, error) {
	if c.err != nil {
		return "", c.err
	}
	path, exists := c.paths[abiType]
	if !exists {
		if abiType == "erc20" && c.paths[abiType] == "" {
			return "", errors.New("simulated config error for missing path")
		}
		return "", errors.New("path not found for " + abiType)
	}
	return path, nil
}

func TestLoadABIByAddress(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		mockSetup      func(*mocks.MockBlockExplorer, *mocks.MockBlockchainClient, *mocks.MockABIUtils, *sync.Map)
		expectedResult string
		expectError    bool
		errorCode      string
		address        string
	}{
		{
			name:    "Cache hit",
			address: "0x1234567890123456789012345678901234567890",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				cache.Store("0x1234567890123456789012345678901234567890", `[{"cached": true}]`)
			},
			expectedResult: `[{"cached": true}]`,
			expectError:    false,
		},
		{
			name:    "Explorer success - not a proxy (implementation method not found)",
			address: "0x1234567890123456789012345678901234567890",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0x1234567890123456789012345678901234567890"
				contractABI := `[{"name": "getValue", "type": "function"}]`
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{
					ABI:          contractABI,
					ContractName: "TestContract",
				}, nil)
				ma.On("Pack", contractABI, "implementation", mock.Anything).Return(nil, coreerrors.NewABIProxyMethodNotFoundError(addr, "implementation"))
			},
			expectedResult: `[{"name": "getValue", "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Explorer success - not a proxy (implementation signature invalid)",
			address: "0x1234567890123456789012345678901234567890",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0x1234567890123456789012345678901234567890"
				contractABI := `[{"name": "getValue", "type": "function"}]`
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{
					ABI:          contractABI,
					ContractName: "TestContract",
				}, nil)
				ma.On("Pack", contractABI, "implementation", mock.Anything).Return(nil, &coreerrors.Vault0Error{Code: coreerrors.ErrCodeABIProxyMethodSignatureInvalid, Message: "simulated invalid signature for implementation", Details: map[string]any{"method": "implementation"}}).Once()
			},
			expectedResult: `[{"name": "getValue", "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Explorer success - proxy contract",
			address: "0xA00000000000000000000000000000000000000A",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				proxyAddrChecksum := "0xA00000000000000000000000000000000000000A"
				implAddrOriginal := "0xb00000000000000000000000000000000000000b" // Renamed for clarity
				proxyABI := `[{"name": "implementation", "type": "function", "outputs": [{"type": "address"}]}]`
				implABI := `[{"name": "actualFunction", "type": "function"}]`
				packedCall := []byte{1, 2, 3, 4}
				implTypeAddr, _ := types.NewAddress(types.ChainTypeEthereum, implAddrOriginal)
				paddedImplAddrBytes := make([]byte, 32)
				copy(paddedImplAddrBytes[12:], common.HexToAddress(implTypeAddr.Address).Bytes())
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(proxyAddrChecksum)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedCall, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(proxyAddrChecksum), matchers.AddressMatcher(proxyAddrChecksum), packedCall).Return(paddedImplAddrBytes, nil).Once()
				ma.On("Unpack", proxyABI, "implementation", paddedImplAddrBytes).Return(map[string]any{"0": *implTypeAddr}, nil).Once()
				ma.On("GetAddressFromArgs", map[string]any{"0": *implTypeAddr}, "0").Return(*implTypeAddr, nil).Once()
				// Use AddressMatcher to handle case variations
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(implTypeAddr.Address)).Return(&blockexplorer.ContractInfo{ABI: implABI}, nil).Once()
			},
			expectedResult: `[{"name": "actualFunction", "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Explorer error",
			address: "0xBAD00000000000000000000000000000000000DD",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				checksummedAddr := "0xbAd00000000000000000000000000000000000DD"
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(checksummedAddr)).Return(nil, coreerrors.NewContractNotFoundError(checksummedAddr, "ethereum"))
			},
			expectedResult: "",
			expectError:    true,
			errorCode:      coreerrors.ErrCodeContractNotFound,
		},
		{
			name:    "Empty ABI from explorer",
			address: "0xABCDEF0123456789ABCDEF0123456789ABCDEF01",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				checksummedAddr := "0xabCDeF0123456789AbcdEf0123456789aBCDEF01"
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(checksummedAddr)).Return(&blockexplorer.ContractInfo{ABI: ""}, nil)
			},
			expectedResult: "",
			expectError:    true,
			errorCode:      coreerrors.ErrCodeABIUnavailableOrUnverified,
		},
		{
			name:    "Proxy: Pack fails (not MethodNotFound), fallback to proxy ABI",
			address: "0xC00000000000000000000000000000000000000C",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0xC00000000000000000000000000000000000000C"
				proxyABI := `[{"name": "someFunction", "type": "function"}]`
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(nil, coreerrors.NewABIProxyPackError(errors.New("some pack error"), addr, "implementation")).Once()
			},
			expectedResult: `[{"name": "someFunction", "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy: CallContract fails, fallback to proxy ABI",
			address: "0xD00000000000000000000000000000000000000D",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0xD00000000000000000000000000000000000000D"
				proxyABI := `[{"name": "proxyFunction", "outputs": [{"type": "address"}], "type": "function"}]`
				packedCall := []byte{1, 2, 3, 4}
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedCall, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(addr), matchers.AddressMatcher(addr), packedCall).Return(nil, coreerrors.NewABIProxyCallError(errors.New("rpc error"), addr, "implementation", "ethereum")).Once()
			},
			expectedResult: `[{"name": "proxyFunction", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy: CallContract returns empty, fallback to proxy ABI",
			address: "0xE00000000000000000000000000000000000000E",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0xE00000000000000000000000000000000000000E"
				proxyABI := `[{"name": "proxyFunctionEmpty", "outputs": [{"type": "address"}], "type": "function"}]`
				packedCall := []byte{1, 2, 3, 4}
				emptyBytes := []byte{}

				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()

				// Setting up ALL expected calls in the sequence they'll be called
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedCall, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(addr), matchers.AddressMatcher(addr), packedCall).Return(emptyBytes, nil).Once()
				ma.On("Unpack", proxyABI, "implementation", matchers.EmptyBytesMatcher()).Return(nil, coreerrors.NewABIProxyEmptyResultError(addr, "implementation")).Maybe()
			},
			expectedResult: `[{"name": "proxyFunctionEmpty", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy: Unpack fails, fallback to proxy ABI",
			address: "0xF00000000000000000000000000000000000000F",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0xF00000000000000000000000000000000000000F"
				proxyABI := `[{"name": "proxyFunctionUnpack", "outputs": [{"type": "address"}], "type": "function"}]`
				packedCall := []byte{1, 2, 3, 4}
				callResult := []byte{5, 6, 7, 8}
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedCall, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(addr), matchers.AddressMatcher(addr), packedCall).Return(callResult, nil).Once()
				ma.On("Unpack", proxyABI, "implementation", callResult).Return(nil, coreerrors.NewABIProxyUnpackError(errors.New("unpack error"), addr, "implementation", "detail")).Once()
			},
			expectedResult: `[{"name": "proxyFunctionUnpack", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy: GetAddressFromArgs fails, fallback to proxy ABI",
			address: "0xA1000000000000000000000000000000000000A1",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0xA1000000000000000000000000000000000000A1"
				proxyABI := `[{"name": "proxyFunctionGetAddr", "outputs": [{"type": "address"}], "type": "function"}]`
				packedCall := []byte{1, 2, 3, 4}
				callResult := []byte{5, 6, 7, 8}
				unpackedArgs := map[string]any{"0": "notAnAddress"}
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedCall, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(addr), matchers.AddressMatcher(addr), packedCall).Return(callResult, nil).Once()
				ma.On("Unpack", proxyABI, "implementation", callResult).Return(unpackedArgs, nil).Once()
				ma.On("GetAddressFromArgs", unpackedArgs, "0").Return(types.Address{}, coreerrors.NewABIProxyAddressConversionError(errors.New("conversion error"), "impl", "eth")).Once()
			},
			expectedResult: `[{"name": "proxyFunctionGetAddr", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy: Impl contract not found, fallback to proxy ABI",
			address: "0xB1000000000000000000000000000000000000B1",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				proxyAddrStr := "0xB1000000000000000000000000000000000000B1"
				implAddrStr := "0xB2000000000000000000000000000000000000B2"
				proxyABI := `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`
				packedCall := []byte{1, 2, 3, 4}
				implTypeAddr, _ := types.NewAddress(types.ChainTypeEthereum, implAddrStr)
				paddedImplAddrBytes := make([]byte, 32)
				copy(paddedImplAddrBytes[12:], common.HexToAddress(implTypeAddr.Address).Bytes())
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(proxyAddrStr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedCall, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(proxyAddrStr), matchers.AddressMatcher(proxyAddrStr), packedCall).Return(paddedImplAddrBytes, nil).Once()
				ma.On("Unpack", proxyABI, "implementation", paddedImplAddrBytes).Return(map[string]any{"0": *implTypeAddr}, nil).Once()
				ma.On("GetAddressFromArgs", map[string]any{"0": *implTypeAddr}, "0").Return(*implTypeAddr, nil).Once()
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(implTypeAddr.Address)).Return(nil, coreerrors.NewContractNotFoundError(implTypeAddr.Address, "ethereum")).Once()
			},
			expectedResult: `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy: Impl ABI unavailable, fallback to proxy ABI",
			address: "0xC1000000000000000000000000000000000000C1",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				proxyAddrStr := "0xC1000000000000000000000000000000000000C1"
				implAddrStr := "0xC2000000000000000000000000000000000000C2"
				proxyABI := `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`
				packedCall := []byte{1, 2, 3, 4}
				implTypeAddr, _ := types.NewAddress(types.ChainTypeEthereum, implAddrStr)
				paddedImplAddrBytes := make([]byte, 32)
				copy(paddedImplAddrBytes[12:], common.HexToAddress(implTypeAddr.Address).Bytes())
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(proxyAddrStr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedCall, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(proxyAddrStr), matchers.AddressMatcher(proxyAddrStr), packedCall).Return(paddedImplAddrBytes, nil).Once()
				ma.On("Unpack", proxyABI, "implementation", paddedImplAddrBytes).Return(map[string]any{"0": *implTypeAddr}, nil).Once()
				ma.On("GetAddressFromArgs", map[string]any{"0": *implTypeAddr}, "0").Return(*implTypeAddr, nil).Once()
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(implTypeAddr.Address)).Return(&blockexplorer.ContractInfo{ABI: ""}, nil).Once()
			},
			expectedResult: `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy resolution error - call fails",
			address: "0xD1000000000000000000000000000000000000D1",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0xD1000000000000000000000000000000000000D1"
				proxyABI := `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`
				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil)
				packedImpl := []byte{0x12, 0x34, 0x56, 0x78}
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedImpl, nil)
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(addr), matchers.AddressMatcher(addr), packedImpl).Return([]byte{}, errors.New("rpc error"))
			},
			expectedResult: `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
		{
			name:    "Proxy implementation returns empty",
			address: "0xE1000000000000000000000000000000000000E1",
			mockSetup: func(me *mocks.MockBlockExplorer, mbc *mocks.MockBlockchainClient, ma *mocks.MockABIUtils, cache *sync.Map) {
				addr := "0xE1000000000000000000000000000000000000E1"
				proxyABI := `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`
				packedImpl := []byte{0x12, 0x34, 0x56, 0x78}
				emptyBytes := []byte{}

				me.On("GetContract", mock.Anything, matchers.AddressMatcher(addr)).Return(&blockexplorer.ContractInfo{ABI: proxyABI}, nil).Once()

				// Setting up ALL expected calls in the sequence they'll be called
				ma.On("Pack", proxyABI, "implementation", mock.Anything).Return(packedImpl, nil).Once()
				mbc.On("CallContract", mock.Anything, matchers.AddressMatcher(addr), matchers.AddressMatcher(addr), packedImpl).Return(emptyBytes, nil).Once()
				ma.On("Unpack", proxyABI, "implementation", matchers.EmptyBytesMatcher()).Return(nil, coreerrors.NewABIProxyEmptyResultError(addr, "implementation")).Maybe()
			},
			expectedResult: `[{"name": "implementation", "outputs": [{"type": "address"}], "type": "function"}]`,
			expectError:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loader, _, mockExplorer, mockBlockchainClient, mockABIUtils := setupTestABILoader()

			// Parse test address
			addr, err := types.NewAddress(types.ChainTypeEthereum, tc.address)
			require.NoError(t, err)

			// Setup mocks
			tc.mockSetup(mockExplorer, mockBlockchainClient, mockABIUtils, loader.abiCache)

			// Call the method
			result, err := loader.LoadABIByAddress(context.Background(), *addr)

			// Check results
			if tc.expectError {
				assert.Error(t, err)
				if tc.errorCode != "" {
					assert.True(t, coreerrors.IsError(err, tc.errorCode), "Expected error code %s, got %v", tc.errorCode, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)

				// Check if result is cached
				cachedResult, found := loader.abiCache.Load(addr.Address)
				assert.True(t, found)
				assert.Equal(t, tc.expectedResult, cachedResult)
			}

			// Verify mock expectations for all mocks
			mockExplorer.AssertExpectations(t)
			mockBlockchainClient.AssertExpectations(t)
			mockABIUtils.AssertExpectations(t)
		})
	}
}
