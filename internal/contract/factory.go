package contract

import (
	"fmt"

	"vault0/internal/blockchain"
	"vault0/internal/config"
	"vault0/internal/types"
	"vault0/internal/wallet"
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

// Create returns a SmartContract instance for the specified chain type
func (f *Factory) Create(blockchain blockchain.Blockchain, wallet wallet.Wallet) (SmartContract, error) {
	// Get the chain type from the blockchain
	blockchainType := blockchain.ChainType()
	walletType := wallet.ChainType()

	// Validate that the wallet and blockchain have matching chain types
	if blockchainType != walletType {
		return nil, fmt.Errorf("blockchain chain type %s does not match wallet chain type %s: %w",
			blockchainType, walletType, types.ErrUnsupportedChain)
	}

	// Create the appropriate implementation based on chain type
	switch blockchainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// These are all EVM-compatible chains, so use EVMSmartContract
		return NewEVMSmartContract(blockchain, wallet, &f.config)
	default:
		return nil, fmt.Errorf("unsupported chain type: %s: %w", blockchainType, types.ErrUnsupportedChain)
	}
}
