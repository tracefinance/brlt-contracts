package blockchain

import (
	"fmt"
	"vault0/internal/config"
)

// Factory creates blockchain implementations
type Factory struct {
	cfg *config.Config
}

// NewFactory creates a new blockchain factory with the given configuration
func NewFactory(cfg *config.Config) *Factory {
	return &Factory{
		cfg: cfg,
	}
}

// NewBlockchain creates a new blockchain implementation for the given chain type
func (f *Factory) NewBlockchain(chainType ChainType) (Blockchain, error) {
	var chainCfg *config.BlockchainConfig
	var chainName string

	switch chainType {
	case Ethereum:
		chainCfg = &f.cfg.Blockchains.Ethereum
		chainName = "Ethereum"
	case Polygon:
		chainCfg = &f.cfg.Blockchains.Polygon
		chainName = "Polygon"
	case Base:
		chainCfg = &f.cfg.Blockchains.Base
		chainName = "Base"
	default:
		return nil, fmt.Errorf("unsupported chain type %s: %w", chainType, ErrChainNotSupported)
	}

	if chainCfg.RPCURL == "" {
		return nil, fmt.Errorf("missing RPC URL for %s: %w", chainName, ErrRPCConnectionFailed)
	}

	chain := Chain{
		ID:          chainCfg.ChainID,
		Type:        chainType,
		Name:        chainName,
		Symbol:      getChainSymbol(chainType),
		RPCUrl:      chainCfg.RPCURL,
		ExplorerUrl: chainCfg.ExplorerURL,
	}

	return NewEVMClient(chain)
}

// getChainSymbol returns the symbol for a given chain type
func getChainSymbol(chainType ChainType) string {
	switch chainType {
	case Ethereum, Base:
		return "ETH"
	case Polygon:
		return "MATIC"
	default:
		return ""
	}
}
