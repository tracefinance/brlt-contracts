package blockchain

import (
	"crypto/elliptic"
	"fmt"
	"sync"
	"vault0/internal/config"
	"vault0/internal/keygen"
	"vault0/internal/types"
)

// Factory creates blockchain implementations
type Factory struct {
	cfg        *config.Config
	clients    map[types.ChainType]Blockchain
	clientsMux sync.RWMutex
}

// NewFactory creates a new blockchain factory with the given configuration
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		cfg:     cfg,
		clients: make(map[types.ChainType]Blockchain),
	}
}

// NewBlockchain creates a new blockchain client for the specified chain type
func (f *Factory) NewBlockchain(chainType types.ChainType) (Blockchain, error) {
	f.clientsMux.Lock()
	defer f.clientsMux.Unlock()

	// Check if we already have a client for this chain type
	if client, exists := f.clients[chainType]; exists {
		return client, nil
	}

	// Create a new client
	chain, err := f.NewChain(chainType)
	if err != nil {
		return nil, err
	}

	client, err := NewEVMBlockchain(chain)
	if err != nil {
		return nil, err
	}

	// Store the client in the map for future use
	f.clients[chainType] = client
	return client, nil
}

// NewChain creates a Chain struct for the specified chain type
func (f *Factory) NewChain(chainType types.ChainType) (Chain, error) {
	var chainCfg *config.BlockchainConfig
	var chainName string

	switch chainType {
	case types.ChainTypeEthereum:
		chainCfg = &f.cfg.Blockchains.Ethereum
		chainName = "Ethereum"
	case types.ChainTypePolygon:
		chainCfg = &f.cfg.Blockchains.Polygon
		chainName = "Polygon"
	case types.ChainTypeBase:
		chainCfg = &f.cfg.Blockchains.Base
		chainName = "Base"
	default:
		return Chain{}, fmt.Errorf("unsupported chain type %s: %w", chainType, ErrChainNotSupported)
	}

	if chainCfg.RPCURL == "" {
		return Chain{}, fmt.Errorf("missing RPC URL for %s: %w", chainName, ErrRPCConnectionFailed)
	}

	// Determine the key type and curve for the chain
	keyType, curve := getChainCryptoParams(chainType)

	return Chain{
		ID:          chainCfg.ChainID,
		Type:        chainType,
		Name:        chainName,
		Symbol:      getChainSymbol(chainType),
		RPCUrl:      chainCfg.RPCURL,
		ExplorerUrl: chainCfg.ExplorerURL,
		KeyType:     keyType,
		Curve:       curve,
	}, nil
}

// getChainCryptoParams returns the appropriate key type and elliptic curve for a given blockchain
func getChainCryptoParams(chainType types.ChainType) (keygen.KeyType, elliptic.Curve) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// All EVM-compatible chains use ECDSA with secp256k1
		return keygen.KeyTypeECDSA, keygen.Secp256k1Curve
	// Add more chain types as needed
	default:
		// For unknown chains, default to ECDSA with P-256
		return keygen.KeyTypeECDSA, elliptic.P256()
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
