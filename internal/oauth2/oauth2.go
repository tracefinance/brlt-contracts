package oauth2

import (
	"time"
	"vault0/internal/core/db"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
)

// Config represents OAuth2 configuration options
type Config struct {
	// Token expiration times
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration
	// Other configurations
	AllowedGrantTypes []oauth2.GrantType
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		AccessTokenExp:    time.Hour * 2,
		RefreshTokenExp:   time.Hour * 24 * 7,
		AllowedGrantTypes: []oauth2.GrantType{oauth2.AuthorizationCode, oauth2.Refreshing},
	}
}

// Service manages the OAuth2 service
type Service struct {
	config *Config
	db     *db.DB
	server *server.Server
}

// New creates a new OAuth2 service
func New(database *db.DB, config *Config) (*Service, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Create token store
	tokenStore, err := NewTokenStore(database)
	if err != nil {
		return nil, err
	}

	// Create client store
	clientStore, err := NewClientStore(database)
	if err != nil {
		return nil, err
	}

	// Create OAuth2 manager
	manager := manage.NewManager()
	manager.MapTokenStorage(tokenStore)
	manager.MapClientStorage(clientStore)

	// Set token configurations
	manager.SetAuthorizeCodeTokenCfg(&manage.Config{
		AccessTokenExp:    config.AccessTokenExp,
		RefreshTokenExp:   config.RefreshTokenExp,
		IsGenerateRefresh: true,
	})

	// Create OAuth2 server
	srv := server.NewServer(&server.Config{
		TokenType:            "Bearer",
		AllowedResponseTypes: []oauth2.ResponseType{oauth2.Code, oauth2.Token},
		AllowedGrantTypes:    config.AllowedGrantTypes,
	}, manager)

	// Set custom error handler
	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		return errors.NewResponse(err, 500)
	})

	// Set response error handler
	srv.SetResponseErrorHandler(func(re *errors.Response) {
		// Log errors or handle them as needed
	})

	return &Service{
		config: config,
		db:     database,
		server: srv,
	}, nil
}

// Server returns the OAuth2 server instance
func (s *Service) Server() *server.Server {
	return s.server
}

// RegisterRoutes registers OAuth2 routes on a Gin router
func (s *Service) RegisterRoutes(r *gin.Engine) {
	// Create a router group for OAuth2 endpoints
	oauth := r.Group("/oauth2")

	// Register handlers
	handlers := NewHandlers(s)
	oauth.GET("/authorize", handlers.AuthorizeHandler)
	oauth.POST("/token", handlers.TokenHandler)
	oauth.GET("/userinfo", handlers.UserInfoHandler)

	// Add the login handler
	oauth.POST("/login", handlers.LoginHandler)
}
