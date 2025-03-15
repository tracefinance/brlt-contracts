package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/handlers/transaction"
	"vault0/internal/api/handlers/user"
	"vault0/internal/api/handlers/wallet"
	"vault0/internal/config"
	"vault0/internal/logger"
)

// Server represents the API server
type Server struct {
	router             *gin.Engine
	config             *config.Config
	logger             logger.Logger
	userHandler        *user.Handler
	walletHandler      *wallet.Handler
	transactionHandler *transaction.Handler
}

// NewServer creates a new API server
func NewServer(
	logger logger.Logger,
	config *config.Config,
	userHandler *user.Handler,
	walletHandler *wallet.Handler,
	transactionHandler *transaction.Handler,
) *Server {
	router := gin.Default()
	return &Server{
		router:             router,
		logger:             logger,
		config:             config,
		userHandler:        userHandler,
		walletHandler:      walletHandler,
		transactionHandler: transactionHandler,
	}
}

// setupRoutes configures the API routes
func (s *Server) SetupRoutes() {
	// Setup API routes
	api := s.router.Group("/api/v1")

	// Setup user routes
	s.userHandler.SetupRoutes(api)
	s.walletHandler.SetupRoutes(api)
	s.transactionHandler.SetupRoutes(api)

	// Health check endpoint
	api.GET("/health", s.healthHandler)

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
