package blockchain

import (
	"vault0/internal/config"
	blockchainCore "vault0/internal/core/blockchain"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	"vault0/internal/core/wallet"
	"vault0/internal/logger"
	blockchainService "vault0/internal/services/blockchain"
	walletService "vault0/internal/services/wallet"
	"vault0/internal/types"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all blockchain-related routes and their dependencies
func SetupRoutes(router *gin.RouterGroup, db *db.DB, keyStore keystore.KeyStore, cfg *config.Config, logger logger.Logger) {
	// Create chain factory
	chainFactory := types.NewChainFactory(cfg)

	// Create wallet factory
	walletFactory := wallet.NewFactory(keyStore, cfg)

	// Create blockchain factory
	blockchainFactory := blockchainCore.NewFactory(cfg)

	// Create wallet repository and service
	walletRepo := walletService.NewSQLiteRepository(db)
	walletSvc := walletService.NewService(cfg, walletRepo, keyStore, chainFactory, walletFactory, blockchainFactory, logger)

	// Create blockchain repository
	blockchainRepo := blockchainService.NewRepository(db)

	// Create blockchain service
	blockchainSvc := blockchainService.NewService(blockchainRepo, walletSvc, blockchainFactory)

	// Create blockchain handler
	blockchainHandler := NewHandler(blockchainSvc)

	// Register blockchain routes
	blockchainRoutes := router.Group("/blockchains")
	blockchainRoutes.POST("/:chain_type/activate", blockchainHandler.ActivateBlockchain)
	blockchainRoutes.POST("/:chain_type/deactivate", blockchainHandler.DeactivateBlockchain)
	blockchainRoutes.GET("/:chain_type", blockchainHandler.GetBlockchain)
	blockchainRoutes.GET("/", blockchainHandler.ListActiveBlockchains)
}
