package oauth2

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/server"
)

// Handlers manages the OAuth2 route handlers
type Handlers struct {
	service *Service
	server  *server.Server
}

// NewHandlers creates a new handlers instance
func NewHandlers(service *Service) *Handlers {
	return &Handlers{
		service: service,
		server:  service.Server(),
	}
}

// AuthorizeHandler handles the authorization endpoint
func (h *Handlers) AuthorizeHandler(c *gin.Context) {
	// Create an HTTP request object for the OAuth2 server
	r := c.Request

	// Get the HTTP response writer
	w := c.Writer

	// Handle the authorization request
	err := h.server.HandleAuthorizeRequest(w, r)
	if err != nil {
		// Format and return error
		h.handleError(c, err)
		return
	}
}

// TokenHandler handles the token endpoint
func (h *Handlers) TokenHandler(c *gin.Context) {
	// Create an HTTP request object for the OAuth2 server
	r := c.Request

	// Get the HTTP response writer
	w := c.Writer

	// Handle the token request
	err := h.server.HandleTokenRequest(w, r)
	if err != nil {
		// Format and return error
		h.handleError(c, err)
		return
	}
}

// UserInfoHandler handles the userinfo endpoint, protected by OAuth2
func (h *Handlers) UserInfoHandler(c *gin.Context) {
	// Extract token from the request
	tokenString, err := h.extractBearerToken(c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Validate the token
	token, err := h.server.Manager.LoadAccessToken(c, tokenString)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	// Check if the token is expired
	if token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
		return
	}

	// Return user information
	c.JSON(http.StatusOK, gin.H{
		"user_id":    token.GetUserID(),
		"client_id":  token.GetClientID(),
		"scope":      token.GetScope(),
		"expires_in": int64(token.GetAccessExpiresIn().Seconds()),
	})
}

// HandleValidateToken middleware validates an OAuth2 token
func (h *Handlers) HandleValidateToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from the request
		tokenString, err := h.extractBearerToken(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		// Validate the token
		token, err := h.server.Manager.LoadAccessToken(c, tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Check if the token is expired
		if token.GetAccessCreateAt().Add(token.GetAccessExpiresIn()).Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}

		// Store token info in the context for downstream handlers
		c.Set("oauth_token", token)
		c.Set("user_id", token.GetUserID())

		c.Next()
	}
}

// extractBearerToken extracts a bearer token from the Authorization header
func (h *Handlers) extractBearerToken(auth string) (string, error) {
	if auth == "" {
		return "", errors.ErrInvalidAccessToken
	}

	const prefix = "Bearer "
	if len(auth) < len(prefix) {
		return "", errors.ErrInvalidAccessToken
	}

	if auth[0:len(prefix)] != prefix {
		return "", errors.ErrInvalidAccessToken
	}

	return auth[len(prefix):], nil
}

// handleError formats and returns OAuth2 errors
func (h *Handlers) handleError(c *gin.Context, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"
	description := err.Error()

	// Just return a generic error response
	c.JSON(status, gin.H{
		"error":             message,
		"error_description": description,
	})
}
