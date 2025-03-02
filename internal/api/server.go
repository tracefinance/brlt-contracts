package api

import (
	"log"
	"net/http"
	"path/filepath"

	"vault0/internal/config"
	"vault0/internal/db"

	"github.com/gin-gonic/gin"
)

// Server represents the API server
type Server struct {
	router *gin.Engine
	db     *db.DB
	config *config.Config
}

// New creates a new API server
func New(db *db.DB, cfg *config.Config) *Server {
	router := gin.Default()

	server := &Server{
		router: router,
		db:     db,
		config: cfg,
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
	log.Printf("Starting server on %s", address)
	return s.router.Run(address)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() {
	// Close the database connection
	if s.db != nil {
		s.db.Close()
	}
}
