package contract

import (
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/wallet"
	"vault0/internal/types"
)

// Factory creates SmartContract instances for different chains
type Factory struct {
	config config.Config
}

// NewFactory creates a new SmartContract factory
func NewFactory(
	config config.Config,
	blockchain blockchain.Blockchain,
	wallet wallet.Wallet,
) *Factory {
	return &Factory{
		config: config,
	}
}

// NewSmartContract returns a SmartContract instance for the specified chain type
func (f *Factory) NewSmartContract(blockchain blockchain.Blockchain, wallet wallet.Wallet) (SmartContract, error) {
	// Get the chain information from the blockchain and wallet
	blockchainChain := blockchain.Chain()
	walletChain := wallet.Chain()

	// Validate that the wallet and blockchain have matching chain types
	if blockchainChain.Type != walletChain.Type {
		return nil, fmt.Errorf("blockchain chain type %s does not match wallet chain type %s: %w",
			blockchainChain.Type, walletChain.Type, types.ErrUnsupportedChain)
	}

	// Create the appropriate implementation based on chain type
	switch blockchainChain.Type {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// These are all EVM-compatible chains, so use EVMSmartContract
		return NewEVMSmartContract(blockchain, wallet, &f.config)
	default:
		return nil, fmt.Errorf("unsupported chain type: %s: %w", blockchainChain.Type, types.ErrUnsupportedChain)
	}
}
