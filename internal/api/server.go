package api

import (
	"net/http"
	"path/filepath"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	// Import generated docs
	_ "vault0/internal/api/docs"
	"vault0/internal/api/handlers/reference"
	"vault0/internal/api/handlers/signer"
	"vault0/internal/api/handlers/token"
	"vault0/internal/api/handlers/tokenprice"
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
	tokenHandler       *token.Handler
	signerHandler      *signer.Handler
	tokenPriceHandler  *tokenprice.Handler
	referenceHandler   *reference.Handler
}

// NewServer creates a new API server
func NewServer(
	logger logger.Logger,
	config *config.Config,
	userHandler *user.Handler,
	walletHandler *wallet.Handler,
	transactionHandler *transaction.Handler,
	tokenHandler *token.Handler,
	signerHandler *signer.Handler,
	tokenPriceHandler *tokenprice.Handler,
	referenceHandler *reference.Handler,
) *Server {
	router := gin.Default()
	router.Use(cors.Default())

	return &Server{
		router:             router,
		logger:             logger,
		config:             config,
		userHandler:        userHandler,
		walletHandler:      walletHandler,
		transactionHandler: transactionHandler,
		tokenHandler:       tokenHandler,
		signerHandler:      signerHandler,
		tokenPriceHandler:  tokenPriceHandler,
		referenceHandler:   referenceHandler,
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
	s.tokenHandler.SetupRoutes(api)
	s.signerHandler.SetupRoutes(api)
	s.tokenPriceHandler.RegisterRoutes(api)
	s.referenceHandler.SetupRoutes(api)

	// Health check endpoint
	api.GET("/health", s.healthHandler)

	// Swagger documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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
