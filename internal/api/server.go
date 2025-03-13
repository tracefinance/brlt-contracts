package api

import (
	"net/http"
	"path/filepath"

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

	"github.com/gin-gonic/gin"
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
	// API routes group
	apiGroup := s.router.Group("/api")
	{
		// Health check endpoint
		apiGroup.GET("/health", s.healthHandler)

		// Setup user routes with core dependencies
		user.SetupRoutes(apiGroup, s.db)

		// Setup wallet routes with core dependencies
		wallet.SetupRoutes(apiGroup, s.db, s.keystore, s.chainFactory, s.walletFactory, s.config)

		// Setup blockchain routes with core dependencies
		blockchain.SetupRoutes(apiGroup, s.db, s.keystore, s.config)

		// Add more API routes here as needed
	}

	// Serve static files from the UI directory
	s.router.Static("/assets", filepath.Join(s.config.UIPath, "assets"))

	// Setup a catch-all route to serve index.html for SPA routing
	s.router.NoRoute(func(c *gin.Context) {
		// If the request path starts with /api, return 404
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// For all other routes, serve the index.html file to support SPA routing
		c.File(filepath.Join(s.config.UIPath, "index.html"))
	})
}

// healthHandler handles the health check endpoint
func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// Run starts the server on the specified port
func (s *Server) Run() error {
	address := ":" + s.config.Port
	s.logger.Info("Starting server", logger.String("address", address))
	return s.router.Run(address)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	s.logger.Info("Shutting down server")
	// Close the database connection
	if s.db != nil {
		s.db.Close()
	}
}
