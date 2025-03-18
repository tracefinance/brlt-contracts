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
		tokenRoutes.GET("/:id", h.verifyToken)
		tokenRoutes.DELETE("/:id", h.deleteToken)
	}
}

// listTokens handles GET /tokens
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

// verifyToken handles GET /tokens/:id
func (h *Handler) verifyToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	token, err := h.service.VerifyToken(c.Request.Context(), id)
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

// deleteToken handles DELETE /tokens/:id
func (h *Handler) deleteToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	if err := h.service.DeleteToken(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
