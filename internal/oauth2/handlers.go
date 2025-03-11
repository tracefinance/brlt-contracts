package oauth2

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"golang.org/x/crypto/bcrypt"
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

// LoginRequest represents the login request structure
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginHandler authenticates a user with email and password
// and returns an OAuth2 token if authentication is successful
func (h *Handlers) LoginHandler(c *gin.Context) {
	var loginReq LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "error_description": "Invalid request format"})
		return
	}

	// Authenticate user
	userID, err := h.authenticateUser(c, loginReq.Email, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid_grant", "error_description": "Invalid credentials"})
		return
	}

	// Create a new token
	clientID := "default-client" // Use a default client or get from request
	token, err := h.createToken(c, userID, clientID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server_error", "error_description": "Failed to generate token"})
		return
	}

	// Return the token
	c.JSON(http.StatusOK, gin.H{
		"access_token":  token.GetAccess(),
		"token_type":    "Bearer",
		"refresh_token": token.GetRefresh(),
		"expires_in":    int64(token.GetAccessExpiresIn().Seconds()),
		"scope":         token.GetScope(),
	})
}

// authenticateUser validates the email and password against the database
func (h *Handlers) authenticateUser(c context.Context, email, password string) (string, error) {
	// Query the user from the database
	query := "SELECT id, password_hash FROM users WHERE email = ?"
	rows, err := h.service.db.ExecuteQueryContext(c, query, email)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	if !rows.Next() {
		return "", errors.ErrInvalidGrant
	}

	var (
		userID       int
		passwordHash string
	)

	err = rows.Scan(&userID, &passwordHash)
	if err != nil {
		return "", err
	}

	// Verify the password
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	if err != nil {
		return "", errors.ErrInvalidGrant
	}

	// Return the user ID as a string
	return strconv.Itoa(userID), nil
}

// createToken creates and stores a new token for the user
func (h *Handlers) createToken(ctx context.Context, userID, clientID string) (oauth2.TokenInfo, error) {
	// Create a new token
	token := &models.Token{
		ClientID:         clientID,
		UserID:           userID,
		Scope:            "all",
		AccessCreateAt:   time.Now(),
		AccessExpiresIn:  h.service.config.AccessTokenExp,
		RefreshCreateAt:  time.Now(),
		RefreshExpiresIn: h.service.config.RefreshTokenExp,
	}

	// Generate access token (a unique string)
	td := time.Now().UTC().Unix()
	token.Access = strconv.FormatInt(td, 10) + "-" + userID

	// Generate refresh token (a unique string)
	token.Refresh = strconv.FormatInt(td, 10) + "-" + userID + "-r"

	// Create a token store and store the token
	tokenStore, err := NewTokenStore(h.service.db)
	if err != nil {
		return nil, err
	}

	err = tokenStore.Create(ctx, token)
	if err != nil {
		return nil, err
	}

	return token, nil
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
