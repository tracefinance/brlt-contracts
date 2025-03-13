package wallet

import (
	"context"

	"github.com/gin-gonic/gin"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	coreWallet "vault0/internal/core/wallet"
	"vault0/internal/logger"
	walletService "vault0/internal/services/wallet"
	"vault0/internal/types"
)

// SetupRoutes configures all wallet-related routes and their dependencies
func SetupRoutes(
	router *gin.RouterGroup,
	db *db.DB,
	keyStore keystore.KeyStore,
	chainFactory types.ChainFactory,
	walletFactory coreWallet.Factory,
	blockchainFactory blockchain.Factory,
	cfg *config.Config,
	log logger.Logger,
) {
	// Create wallet repository
	walletRepo := walletService.NewSQLiteRepository(db)

	// Create wallet service
	walletSvc := walletService.NewService(cfg, walletRepo, keyStore, chainFactory, walletFactory, blockchainFactory, log)

	// Create wallet handler
	walletHandler := NewHandler(walletSvc)

	// Register wallet routes
	walletRoutes := router.Group("/wallets")
	{
		walletRoutes.POST("/", walletHandler.CreateWallet)
		walletRoutes.GET("/", walletHandler.ListWallets)
		walletRoutes.GET("/:id", walletHandler.GetWallet)
		walletRoutes.PUT("/:id", walletHandler.UpdateWallet)
		walletRoutes.DELETE("/:id", walletHandler.DeleteWallet)
	}

	// Subscribe to events for all wallets
	if err := walletSvc.SubscribeToEvents(context.Background()); err != nil {
		log.Error("Failed to subscribe to wallet events", logger.Error(err))
	}
}
