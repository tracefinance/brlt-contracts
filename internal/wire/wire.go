//go:build wireinject

package wire

import (
	"vault0/internal/api"
	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	"vault0/internal/core/wallet"
	"vault0/internal/logger"
	"vault0/internal/types"

	"github.com/google/wire"
)

// Container holds all application dependencies
type Container struct {
	Config               *config.Config
	DB                   *db.DB
	Logger               logger.Logger
	KeyStore             keystore.KeyStore
	Chains               types.Chains
	WalletFactory        wallet.Factory
	BlockchainRegistry   blockchain.Registry
	BlockExplorerFactory blockexplorer.Factory
	Server               *api.Server
	Services             *Services
}

// NewContainer creates a new dependency injection container
func NewContainer(
	config *config.Config,
	db *db.DB,
	logger logger.Logger,
	keyStore keystore.KeyStore,
	chains types.Chains,
	walletFactory wallet.Factory,
	blockchainRegistry blockchain.Registry,
	blockExplorerFactory blockexplorer.Factory,
	server *api.Server,
	services *Services,
) *Container {
	return &Container{
		Config:               config,
		DB:                   db,
		Logger:               logger,
		KeyStore:             keyStore,
		Chains:               chains,
		WalletFactory:        walletFactory,
		BlockchainRegistry:   blockchainRegistry,
		BlockExplorerFactory: blockExplorerFactory,
		Server:               server,
		Services:             services,
	}
}

// ContainerSet combines all dependency sets
var ContainerSet = wire.NewSet(
	CoreSet,
	ServerSet,
	ServicesSet,
	NewContainer,
)

// InitializeContainer creates a new container with all dependencies wired up
// BuildContainer is a placeholder function that will be replaced by wire with the actual implementation
func BuildContainer() (*Container, error) {
	wire.Build(ContainerSet)
	return nil, nil
}
