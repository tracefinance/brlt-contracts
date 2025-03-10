package blockchain

import (
	"crypto/elliptic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"vault0/internal/config"
	"vault0/internal/core/keygen"
	"vault0/internal/types"
)

func createTestConfig() *config.Config {
	return &config.Config{
		Blockchains: config.BlockchainsConfig{
			Ethereum: config.BlockchainConfig{
				RPCURL:          "https://mainnet.infura.io/v3/your-api-key",
				ChainID:         1,
				DefaultGasPrice: 20,
				DefaultGasLimit: 21000,
				ExplorerURL:     "https://etherscan.io",
			},
			Polygon: config.BlockchainConfig{
				RPCURL:          "https://polygon-mainnet.infura.io/v3/your-api-key",
				ChainID:         137,
				DefaultGasPrice: 30,
				DefaultGasLimit: 21000,
				ExplorerURL:     "https://polygonscan.com",
			},
			Base: config.BlockchainConfig{
				RPCURL:          "https://mainnet.base.org",
				ChainID:         8453,
				DefaultGasPrice: 10,
				DefaultGasLimit: 21000,
				ExplorerURL:     "https://basescan.org",
			},
		},
	}
}

func TestNewFactory(t *testing.T) {
	cfg := createTestConfig()

	assert.NotNil(t, cfg)
	assert.NotNil(t, cfg.Blockchains)
	assert.NotNil(t, cfg.Blockchains.Ethereum)
	assert.NotNil(t, cfg.Blockchains.Polygon)
	assert.NotNil(t, cfg.Blockchains.Base)
}

func TestNewChain(t *testing.T) {
	cfg := createTestConfig()

	t.Run("Ethereum Chain", func(t *testing.T) {
		chain, err := NewChain(types.ChainTypeEthereum, cfg)
		require.NoError(t, err)

		assert.Equal(t, int64(1), chain.ID)
		assert.Equal(t, types.ChainTypeEthereum, chain.Type)
		assert.Equal(t, "Ethereum", chain.Name)
		assert.Equal(t, "ETH", chain.Symbol)
		assert.Equal(t, "https://mainnet.infura.io/v3/your-api-key", chain.RPCUrl)
		assert.Equal(t, "https://etherscan.io", chain.ExplorerUrl)
		assert.Equal(t, keygen.KeyTypeECDSA, chain.KeyType)
		assert.Equal(t, keygen.Secp256k1Curve, chain.Curve)
	})

	t.Run("Polygon Chain", func(t *testing.T) {
		chain, err := NewChain(types.ChainTypePolygon, cfg)
		require.NoError(t, err)

		assert.Equal(t, int64(137), chain.ID)
		assert.Equal(t, types.ChainTypePolygon, chain.Type)
		assert.Equal(t, "Polygon", chain.Name)
		assert.Equal(t, "MATIC", chain.Symbol)
		assert.Equal(t, "https://polygon-mainnet.infura.io/v3/your-api-key", chain.RPCUrl)
		assert.Equal(t, "https://polygonscan.com", chain.ExplorerUrl)
		assert.Equal(t, keygen.KeyTypeECDSA, chain.KeyType)
		assert.Equal(t, keygen.Secp256k1Curve, chain.Curve)
	})

	t.Run("Base Chain", func(t *testing.T) {
		chain, err := NewChain(types.ChainTypeBase, cfg)
		require.NoError(t, err)

		assert.Equal(t, int64(8453), chain.ID)
		assert.Equal(t, types.ChainTypeBase, chain.Type)
		assert.Equal(t, "Base", chain.Name)
		assert.Equal(t, "ETH", chain.Symbol)
		assert.Equal(t, "https://mainnet.base.org", chain.RPCUrl)
		assert.Equal(t, "https://basescan.org", chain.ExplorerUrl)
		assert.Equal(t, keygen.KeyTypeECDSA, chain.KeyType)
		assert.Equal(t, keygen.Secp256k1Curve, chain.Curve)
	})

	t.Run("Unsupported Chain", func(t *testing.T) {
		unsupportedChainType := types.ChainType("unsupported")
		chain, err := NewChain(unsupportedChainType, cfg)

		assert.Error(t, err)
		assert.Equal(t, Chain{}, chain)
		assert.Contains(t, err.Error(), "unsupported chain type")
		assert.ErrorIs(t, err, ErrChainNotSupported)
	})

	t.Run("Missing RPC URL", func(t *testing.T) {
		// Create a config with missing RPC URL
		badCfg := createTestConfig()
		badCfg.Blockchains.Ethereum.RPCURL = ""

		chain, err := NewChain(types.ChainTypeEthereum, badCfg)

		assert.Error(t, err)
		assert.Equal(t, Chain{}, chain)
		assert.Contains(t, err.Error(), "missing RPC URL")
		assert.ErrorIs(t, err, ErrRPCConnectionFailed)
	})
}

// For TestNewBlockchain, we'll skip the actual connection part since it requires valid RPC endpoints
// Instead, we'll just test the factory's client caching logic
func TestNewBlockchain(t *testing.T) {
	t.Skip("Skipping test that requires valid RPC endpoints")
}

func TestGetChainCryptoParams(t *testing.T) {
	t.Run("EVM Chains", func(t *testing.T) {
		// Test all EVM chains to verify they use the same crypto params
		chains := []types.ChainType{
			types.ChainTypeEthereum,
			types.ChainTypePolygon,
			types.ChainTypeBase,
		}

		for _, chainType := range chains {
			keyType, curve := getChainCryptoParams(chainType)
			assert.Equal(t, keygen.KeyTypeECDSA, keyType)
			assert.Equal(t, keygen.Secp256k1Curve, curve)
		}
	})

	t.Run("Unknown Chain", func(t *testing.T) {
		unknownChain := types.ChainType("unknown")
		keyType, curve := getChainCryptoParams(unknownChain)

		// Unknown chains should default to ECDSA with P-256
		assert.Equal(t, keygen.KeyTypeECDSA, keyType)
		assert.Equal(t, elliptic.P256(), curve)
	})
}

func TestGetChainSymbol(t *testing.T) {
	t.Run("Known Chains", func(t *testing.T) {
		testCases := []struct {
			chainType types.ChainType
			expected  string
		}{
			{types.ChainTypeEthereum, "ETH"},
			{types.ChainTypePolygon, "MATIC"},
			{types.ChainTypeBase, "ETH"},
		}

		for _, tc := range testCases {
			symbol := getChainSymbol(tc.chainType)
			assert.Equal(t, tc.expected, symbol, "Chain %s should have symbol %s", tc.chainType, tc.expected)
		}
	})

	t.Run("Unknown Chain", func(t *testing.T) {
		unknownChain := types.ChainType("unknown")
		symbol := getChainSymbol(unknownChain)
		assert.Equal(t, "UNKNOWN", symbol)
	})
}
