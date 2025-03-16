package wire

import (
	"github.com/google/wire"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/contract"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	"vault0/internal/core/tokenstore"
	"vault0/internal/core/wallet"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Core holds all core infrastructure dependencies
type Core struct {
	Config               *config.Config
	DB                   *db.DB
	Logger               logger.Logger
	KeyStore             keystore.KeyStore
	TokenStore           tokenstore.TokenStore
	Chains               *types.Chains
	WalletFactory        wallet.Factory
	BlockchainRegistry   blockchain.Registry
	ContractFactory      contract.Factory
	BlockExplorerFactory blockexplorer.Factory
}

// NewCore creates a new Core instance with all core dependencies
func NewCore(
	config *config.Config,
	db *db.DB,
	logger logger.Logger,
	keyStore keystore.KeyStore,
	tokenStore tokenstore.TokenStore,
	chains *types.Chains,
	walletFactory wallet.Factory,
	blockchainRegistry blockchain.Registry,
	contractFactory contract.Factory,
	blockExplorerFactory blockexplorer.Factory,
) *Core {
	return &Core{
		Config:               config,
		DB:                   db,
		Logger:               logger,
		KeyStore:             keyStore,
		TokenStore:           tokenStore,
		Chains:               chains,
		WalletFactory:        walletFactory,
		BlockchainRegistry:   blockchainRegistry,
		ContractFactory:      contractFactory,
		BlockExplorerFactory: blockExplorerFactory,
	}
}

// CoreSet combines all core dependencies
var CoreSet = wire.NewSet(
	config.LoadConfig,
	db.NewDatabase,
	logger.NewLogger,
	keystore.NewKeyStore,
	tokenstore.NewTokenStore,
	blockchain.NewRegistry,
	wallet.NewFactory,
	blockexplorer.NewFactory,
	contract.NewFactory,
	types.NewChains,
	NewCore,
)
