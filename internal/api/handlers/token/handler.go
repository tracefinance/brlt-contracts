package token

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
	"vault0/internal/api/utils"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/token"
	"vault0/internal/types"
)

// Handler manages token-related API endpoints
type Handler struct {
	service token.Service
	logger  logger.Logger
}

// NewHandler creates a new token handler
func NewHandler(service token.Service, logger logger.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

// SetupRoutes configures the token API routes
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	errorHandler := middleares.NewErrorHandler(nil)

	tokenRoutes := router.Group("/tokens")
	tokenRoutes.Use(errorHandler.Middleware())
	tokenRoutes.GET("", h.listTokens)
	tokenRoutes.POST("", h.addToken)
	tokenRoutes.GET("/verify/:address", h.verifyToken)
	tokenRoutes.GET("/:chainType/:address", h.getToken)
	tokenRoutes.DELETE("/:address", h.deleteToken)

}

// listTokens handles GET /tokens
// @Summary List tokens
// @Description Get a paginated list of tokens with optional filtering
// @Tags tokens
// @Produce json
// @Param chain_type query string false "Filter by chain type (ethereum, polygon, etc.)"
// @Param token_type query string false "Filter by token type (erc20, erc721, etc.)"
// @Param next_token query string false "Token for pagination (empty for first page)"
// @Param limit query int false "Number of items to return (default: 10)" default(10)
// @Success 200 {object} utils.PagedResponse[TokenResponse]
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /tokens [get]
func (h *Handler) listTokens(c *gin.Context) {
	var req ListTokensRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errors.NewInvalidParameterError("query", "invalid query parameters format or value"))
		return
	}

	// Set default limit if not provided
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Build filter
	filter := token.TokenFilter{}

	if req.ChainType != "" {
		chainType := types.ChainType(req.ChainType)
		filter.ChainType = &chainType
	}

	if req.TokenType != "" {
		tokenType := types.TokenType(req.TokenType)
		filter.TokenType = &tokenType
	}

	// Get paginated tokens using token-based pagination
	tokens, err := h.service.ListTokens(c.Request.Context(), filter, limit, req.NextToken)
	if err != nil {
		c.Error(err)
		return
	}

	// Convert to response using the helper function
	response := utils.NewPagedResponse(tokens, TokenToResponse)

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
		Address:   token.Address,
		ChainType: token.ChainType,
		Symbol:    token.Symbol,
		Decimals:  token.Decimals,
		Type:      token.Type,
	}

	c.JSON(http.StatusOK, response)
}

// getToken handles GET /tokens/:chainType/:address
// @Summary Get token details
// @Description Get a token by its chain type and address
// @Tags tokens
// @Produce json
// @Param chainType path string true "Chain type (ethereum, polygon, base)"
// @Param address path string true "Token address or 'native'"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Token not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /tokens/{chainType}/{address} [get]
func (h *Handler) getToken(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.Error(errors.NewInvalidParameterError("address", "cannot be empty"))
		return
	}

	chainTypeStr := c.Param("chainType")
	if chainTypeStr == "" {
		c.Error(errors.NewMissingParameterError("chainType"))
		return
	}

	chainType := types.ChainType(chainTypeStr)

	token, err := h.service.GetTokenByChainAndAddress(c.Request.Context(), chainType, address)
	if err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := TokenResponse{
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
