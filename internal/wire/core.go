package wire

import (
	"github.com/google/wire"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/contract"
	"vault0/internal/core/keystore"
	"vault0/internal/core/pricefeed"
	"vault0/internal/core/tokenstore"
	"vault0/internal/core/wallet"
	"vault0/internal/db"
	"vault0/internal/logger"
	"vault0/internal/types"
)

func NewSnowflake(config *config.Config) (*db.Snowflake, error) {
	return db.NewSnowflake(config.Snowflake.DataCenterID, config.Snowflake.MachineID)
}

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
	PriceFeed            pricefeed.PriceFeed
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
	priceFeed pricefeed.PriceFeed,
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
		PriceFeed:            priceFeed,
	}
}

// CoreSet combines all core dependencies
var CoreSet = wire.NewSet(
	config.LoadConfig,
	NewSnowflake,
	db.NewDatabase,
	logger.NewLogger,
	keystore.NewKeyStore,
	tokenstore.NewTokenStore,
	blockchain.NewRegistry,
	wallet.NewFactory,
	blockexplorer.NewFactory,
	contract.NewFactory,
	types.NewChains,
	pricefeed.NewPriceFeed,
	NewCore,
)
