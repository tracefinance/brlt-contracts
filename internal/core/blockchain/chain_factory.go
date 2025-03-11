package blockchain

import (
	"fmt"
	"vault0/internal/config"
	"vault0/internal/types"
)

type ChainFactory interface {
	// NewChain creates a Chain struct for the specified blockchain type.
	// It loads configuration from the provided config object and sets up
	// all necessary parameters for interacting with the blockchain.
	//
	// Parameters:
	//   - chainType: The type of blockchain to create (e.g., Ethereum, Polygon)
	//   - config: Configuration object containing blockchain-specific settings
	//
	// Returns:
	//   - A fully initialized Chain struct if successful
	//   - Error if the chain type is unsupported or if required configuration is missing
	NewChain(chainType types.ChainType) (Chain, error)
}

type chainFactory struct {
	cfg *config.Config
}

func NewChainFactory(cfg *config.Config) ChainFactory {
	return &chainFactory{
		cfg: cfg,
	}
}

func (f *chainFactory) NewChain(chainType types.ChainType) (Chain, error) {
	chainCfg, err := getChainConfig(chainType, f.cfg)
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
