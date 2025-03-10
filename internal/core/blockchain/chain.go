package blockchain

import (
	"crypto/elliptic"
	"fmt"
	"vault0/internal/config"
	"vault0/internal/core/keygen"
	"vault0/internal/types"
)

// Chain represents information about a specific blockchain
type Chain struct {
	ID          int64           // Chain ID
	Type        types.ChainType // Chain type
	Name        string          // Human-readable name
	Symbol      string          // Native currency symbol
	RPCUrl      string          // RPC URL for the chain
	ExplorerUrl string          // Block explorer URL
	Curve       elliptic.Curve  // Elliptic curve for key generation
	KeyType     keygen.KeyType  // Key type for the blockchain
}

// NewChain creates a Chain struct for the specified chain type
func NewChain(chainType types.ChainType, config *config.Config) (Chain, error) {
	chainCfg, err := getChainConfig(chainType, config)
	if err != nil {
		return Chain{}, err
	}

	if chainCfg.RPCURL == "" {
		return Chain{}, fmt.Errorf("missing RPC URL for %s: %w", chainType, ErrRPCConnectionFailed)
	}

	// Determine the key type and curve for the chain
	keyType, curve := getChainCryptoParams(chainType)

	return Chain{
		ID:          chainCfg.ChainID,
		Type:        chainType,
		Name:        getChainName(chainType),
		Symbol:      getChainSymbol(chainType),
		RPCUrl:      chainCfg.RPCURL,
		ExplorerUrl: chainCfg.ExplorerURL,
		KeyType:     keyType,
		Curve:       curve,
	}, nil
}

// getChainConfig returns the configuration for a given chain type
func getChainConfig(chainType types.ChainType, config *config.Config) (*config.BlockchainConfig, error) {
	switch chainType {
	case types.ChainTypeEthereum:
		return &config.Blockchains.Ethereum, nil
	case types.ChainTypePolygon:
		return &config.Blockchains.Polygon, nil
	case types.ChainTypeBase:
		return &config.Blockchains.Base, nil
	default:
		return nil, fmt.Errorf("unsupported chain type %s: %w", chainType, ErrChainNotSupported)
	}
}

// getChainName returns the human-readable name for a given blockchain
func getChainName(chainType types.ChainType) string {
	switch chainType {
	case types.ChainTypeEthereum:
		return "Ethereum"
	case types.ChainTypePolygon:
		return "Polygon"
	case types.ChainTypeBase:
		return "Base"
	default:
		return "Unknown"
	}
}

// getChainSymbol returns the native currency symbol for a given blockchain
func getChainSymbol(chainType types.ChainType) string {
	switch chainType {
	case types.ChainTypeEthereum:
		return "ETH"
	case types.ChainTypePolygon:
		return "MATIC"
	case types.ChainTypeBase:
		return "ETH"
	default:
		return "UNKNOWN"
	}
}

// getChainCryptoParams returns the appropriate key type and elliptic curve for a given blockchain
func getChainCryptoParams(chainType types.ChainType) (keygen.KeyType, elliptic.Curve) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// All EVM-compatible chains use ECDSA with secp256k1
		return keygen.KeyTypeECDSA, keygen.Secp256k1Curve
	default:
		// For unknown chains, default to ECDSA with P-256
		return keygen.KeyTypeECDSA, elliptic.P256()
	}
}
