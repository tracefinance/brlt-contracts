package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/handlers/blockchain"
	"vault0/internal/api/handlers/user"
	"vault0/internal/api/handlers/wallet"
	"vault0/internal/config"
	coreBlockchain "vault0/internal/core/blockchain"
	"vault0/internal/core/db"
	"vault0/internal/core/keystore"
	coreWallet "vault0/internal/core/wallet"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Server represents the API server
type Server struct {
	router            *gin.Engine
	db                *db.DB
	config            *config.Config
	keystore          keystore.KeyStore
	chainFactory      types.ChainFactory
	walletFactory     coreWallet.Factory
	blockchainFactory coreBlockchain.Factory
	logger            logger.Logger
}

// NewServer creates a new API server
func NewServer(
	db *db.DB,
	cfg *config.Config,
	keystore keystore.KeyStore,
	chainFactory types.ChainFactory,
	walletFactory coreWallet.Factory,
	blockchainFactory coreBlockchain.Factory,
	log logger.Logger,
) *Server {
	router := gin.Default()

	server := &Server{
		router:            router,
		db:                db,
		config:            cfg,
		keystore:          keystore,
		chainFactory:      chainFactory,
		walletFactory:     walletFactory,
		blockchainFactory: blockchainFactory,
		logger:            log,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures the API routes
func (s *Server) setupRoutes() {
	// Setup API routes
	api := s.router.Group("/api/v1")

	// Setup user routes
	user.SetupRoutes(api, s.db)

	// Setup wallet routes
	wallet.SetupRoutes(
		api,
		s.db,
		s.keystore,
		s.chainFactory,
		s.walletFactory,
		s.blockchainFactory,
		s.config,
		s.logger,
	)

	// Setup blockchain routes
	blockchain.SetupRoutes(
		api,
		s.db,
		s.keystore,
		s.config,
		s.logger,
	)

	// Serve static files for the UI
	if s.config.UIPath != "" {
		s.router.Static("/ui", s.config.UIPath)
		s.router.NoRoute(func(c *gin.Context) {
			c.File(filepath.Join(s.config.UIPath, "index.html"))
		})
	}
}

// healthHandler handles the health check endpoint
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Run starts the server
func (s *Server) Run() error {
	return s.router.Run(":" + s.config.Port)
}

// Shutdown performs cleanup before server shutdown
func (s *Server) Shutdown() {
	// Add cleanup logic here if needed
}
