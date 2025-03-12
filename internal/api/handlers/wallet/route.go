package wallet

import (
	"github.com/gin-gonic/gin"

	"vault0/internal/config"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	coreWallet "vault0/internal/core/wallet"
	walletService "vault0/internal/services/wallet"
	"vault0/internal/types"
)

// SetupRoutes configures all wallet-related routes and their dependencies
func SetupRoutes(router *gin.RouterGroup, db *db.DB, keyStore keystore.KeyStore, chainFactory types.ChainFactory, walletFactory coreWallet.Factory, cfg *config.Config) {
	// Create wallet repository
	walletRepo := walletService.NewSQLiteRepository(db)

	// Create wallet service
	walletSvc := walletService.NewService(cfg, walletRepo, keyStore, chainFactory, walletFactory)

	// Create wallet handler
	walletHandler := NewHandler(walletSvc)

	// Register wallet routes directly
	router.POST("/wallets", walletHandler.CreateWallet)
	router.GET("/wallets", walletHandler.ListWallets)
	router.GET("/wallets/:id", walletHandler.GetWallet)
	router.PUT("/wallets/:id", walletHandler.UpdateWallet)
	router.DELETE("/wallets/:id", walletHandler.DeleteWallet)
}
