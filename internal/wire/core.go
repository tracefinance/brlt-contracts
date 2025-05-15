package wire

import (
	"github.com/google/wire"

	"vault0/internal/config"
	"vault0/internal/core/abi"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/contract"
	"vault0/internal/core/keystore"
	"vault0/internal/core/pricefeed"
	"vault0/internal/core/tokenstore"
	"vault0/internal/core/transaction"
	"vault0/internal/core/wallet"
	"vault0/internal/db"
	"vault0/internal/logger"
	"vault0/internal/types"
)

func NewSnowflake(config *config.Config) (*db.Snowflake, error) {
	return db.NewSnowflake(config.Snowflake.DataCenterID, config.Snowflake.MachineID)
}

// CoreSet combines all core dependencies
var CoreSet = wire.NewSet(
	config.LoadConfig,
	NewSnowflake,
	db.NewDatabase,
	logger.NewLogger,
	keystore.NewKeyStore,
	tokenstore.NewTokenStore,
	types.NewChains,
	pricefeed.NewPriceFeed,
	blockchain.NewFactory,
	wallet.NewFactory,
	blockexplorer.NewFactory,
	contract.NewFactory,
	abi.NewFactory,
	transaction.NewFactory,
	NewCore,
)

// Core holds all core infrastructure dependencies
type Core struct {
	Config                  *config.Config
	DB                      *db.DB
	Logger                  logger.Logger
	KeyStore                keystore.KeyStore
	TokenStore              tokenstore.TokenStore
	Chains                  *types.Chains
	WalletFactory           wallet.Factory
	BlockchainClientFactory blockchain.Factory
	ContractManagerFactory  contract.Factory
	BlockExplorerFactory    blockexplorer.Factory
	ABIFactory              abi.Factory
	PriceFeed               pricefeed.PriceFeed
	TransactionFactory      transaction.Factory
}

// NewCore creates a new Core instance with all core dependencies
func NewCore(
	config *config.Config,
	db *db.DB,
	logger logger.Logger,
	keyStore keystore.KeyStore,
	tokenStore tokenstore.TokenStore,
	chains *types.Chains,
	priceFeed pricefeed.PriceFeed,
	walletFactory wallet.Factory,
	blockchainClientFactory blockchain.Factory,
	contractManagerFactory contract.Factory,
	blockExplorerFactory blockexplorer.Factory,
	abiFactory abi.Factory,
	transactionFactory transaction.Factory,
) *Core {
	return &Core{
		Config:                  config,
		DB:                      db,
		Logger:                  logger,
		KeyStore:                keyStore,
		TokenStore:              tokenStore,
		Chains:                  chains,
		PriceFeed:               priceFeed,
		WalletFactory:           walletFactory,
		BlockchainClientFactory: blockchainClientFactory,
		ContractManagerFactory:  contractManagerFactory,
		BlockExplorerFactory:    blockExplorerFactory,
		ABIFactory:              abiFactory,
		TransactionFactory:      transactionFactory,
	}
}
