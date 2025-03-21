package token

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
	"vault0/internal/errors"
	"vault0/internal/services/token"
	"vault0/internal/types"
)

// Handler manages token-related API endpoints
type Handler struct {
	service token.Service
}

// NewHandler creates a new token handler
func NewHandler(service token.Service) *Handler {
	return &Handler{service: service}
}

// SetupRoutes configures the token API routes
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	errorHandler := middleares.NewErrorHandler(nil)

	tokenRoutes := router.Group("/tokens")
	tokenRoutes.Use(errorHandler.Middleware())
	{
		tokenRoutes.GET("", h.listTokens)
		tokenRoutes.POST("", h.addToken)
		tokenRoutes.GET("/:address", h.verifyToken)
		tokenRoutes.DELETE("/:address", h.deleteToken)
	}
}

// listTokens handles GET /tokens
// @Summary List tokens
// @Description Get a paginated list of tokens with optional filtering
// @Tags tokens
// @Produce json
// @Param chain_type query string false "Filter by chain type (ethereum, polygon, etc.)"
// @Param token_type query string false "Filter by token type (erc20, erc721, etc.)"
// @Param offset query int false "Number of items to skip (default: 0)" default(0)
// @Param limit query int false "Number of items to return (default: 10)" default(10)
// @Success 200 {object} TokenListResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /tokens [get]
func (h *Handler) listTokens(c *gin.Context) {
	// Parse query parameters
	chainTypeStr := c.Query("chain_type")
	tokenTypeStr := c.Query("token_type")
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("offset", "must be a valid integer"))
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("limit", "must be a valid integer"))
		return
	}

	// Build filter
	filter := token.TokenFilter{}

	if chainTypeStr != "" {
		chainType := types.ChainType(chainTypeStr)
		filter.ChainType = &chainType
	}

	if tokenTypeStr != "" {
		tokenType := types.TokenType(tokenTypeStr)
		filter.TokenType = &tokenType
	}

	// Get paginated tokens (directly passing offset and limit)
	tokens, err := h.service.ListTokens(c.Request.Context(), filter, offset, limit)
	if err != nil {
		c.Error(errors.NewOperationFailedError("list tokens", err))
		return
	}

	// Build response
	response := TokenListResponse{
		Items: make([]TokenResponse, len(tokens.Items)),
		Total: int64(len(tokens.Items)),
	}

	for i, t := range tokens.Items {
		response.Items[i] = TokenResponse{
			ID:        t.ID,
			Address:   t.Address,
			ChainType: t.ChainType,
			Symbol:    t.Symbol,
			Decimals:  t.Decimals,
			Type:      t.Type,
		}
	}

	c.JSON(http.StatusOK, response)
}

// addToken handles POST /tokens
// @Summary Add a new token
// @Description Add a new token to the system
// @Tags tokens
// @Accept json
// @Produce json
// @Param token body AddTokenRequest true "Token details"
// @Success 201 {object} TokenResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 409 {object} errors.Vault0Error "Token already exists"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /tokens [post]
func (h *Handler) addToken(c *gin.Context) {
	var req AddTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Let middleware handle binding errors
		c.Error(err)
		return
	}

	// Convert request to token
	token := &types.Token{
		Address:   req.Address,
		ChainType: req.ChainType,
		Symbol:    req.Symbol,
		Decimals:  req.Decimals,
		Type:      req.Type,
	}

	// Add token
	if err := h.service.AddToken(c.Request.Context(), token); err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := TokenResponse{
		ID:        token.ID,
		Address:   token.Address,
		ChainType: token.ChainType,
		Symbol:    token.Symbol,
		Decimals:  token.Decimals,
		Type:      token.Type,
	}

	c.JSON(http.StatusCreated, response)
}

// verifyToken handles GET /tokens/:address
// @Summary Verify token
// @Description Verify a token by its address and return its details
// @Tags tokens
// @Produce json
// @Param address path string true "Token address"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Token not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /tokens/{address} [get]
func (h *Handler) verifyToken(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.Error(errors.NewInvalidParameterError("address", "cannot be empty"))
		return
	}

	token, err := h.service.VerifyToken(c.Request.Context(), address)
	if err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := TokenResponse{
		ID:        token.ID,
		Address:   token.Address,
		ChainType: token.ChainType,
		Symbol:    token.Symbol,
		Decimals:  token.Decimals,
		Type:      token.Type,
	}

	c.JSON(http.StatusOK, response)
}

// deleteToken handles DELETE /tokens/:address
// @Summary Delete token
// @Description Delete a token by its address
// @Tags tokens
// @Param address path string true "Token address"
// @Success 204 "No Content"
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Token not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /tokens/{address} [delete]
func (h *Handler) deleteToken(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.Error(errors.NewInvalidParameterError("address", "cannot be empty"))
		return
	}

	if err := h.service.DeleteToken(c.Request.Context(), address); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
