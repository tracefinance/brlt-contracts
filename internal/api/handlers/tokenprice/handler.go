package tokenprice

import (
	"net/http"
	"strings"

	"vault0/internal/api/utils"
	"vault0/internal/errors"
	"vault0/internal/logger"
	tokensvc "vault0/internal/services/tokenprice"

	"github.com/gin-gonic/gin"
)

// Handler holds the dependencies for the token price API handlers.
type Handler struct {
	service tokensvc.Service
	logger  logger.Logger
}

// NewHandler creates a new token price handler instance.
func NewHandler(svc tokensvc.Service, log logger.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  log.With(logger.String("handler", "tokenprice")),
	}
}

// RegisterRoutes registers the token price API routes with the Gin engine.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	group := router.Group("/token-prices")
	{
		group.GET("", h.ListTokenPrices)
		group.GET("/:symbol", h.GetTokenPriceBySymbol)
	}
}

// ListTokenPrices godoc
// @Summary List token prices
// @Description Get a paginated list of token prices stored in the database.
// @Tags TokenPrices
// @Accept json
// @Produce json
// @Param limit query int false "Maximum number of items to return (0 for all)" default(50) minimum(1) maximum(100)
// @Param next_token query string false "Token for fetching the next page"
// @Param symbol query string false "Token symbol to filter by (can be used multiple times, e.g., ?symbol=BTC&symbol=ETH)"
// @Success 200 {object} utils.PagedResponse[TokenPriceResponse] "Paginated list of token prices"
// @Failure 400 {object} errors.Vault0Error "Invalid query parameters"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /token-prices [get]
func (h *Handler) ListTokenPrices(c *gin.Context) {
	var req ListTokenPricesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errors.NewInvalidParameterError("query", "invalid query parameters format or value"))
		return
	}

	// Use limit from request or 0 to let service apply default
	limit := 0
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Normalize symbols (convert to uppercase)
	var filter *tokensvc.TokenPriceFilter
	if len(req.Symbol) > 0 {
		symbols := make([]string, len(req.Symbol))
		for i, symbol := range req.Symbol {
			symbols[i] = strings.ToUpper(strings.TrimSpace(symbol))
		}
		filter = &tokensvc.TokenPriceFilter{
			Symbols: symbols,
		}
	}

	// Call the service with token-based pagination parameters
	page, err := h.service.ListTokenPrices(c.Request.Context(), filter, limit, req.NextToken)
	if err != nil {
		c.Error(err)
		return
	}

	// Use the generic paged response directly from utils
	c.JSON(http.StatusOK, utils.NewPagedResponse(page, mapModelToResponse))
}

// GetTokenPriceBySymbol godoc
// @Summary Get token price by symbol
// @Description Get the stored price data for a specific token symbol.
// @Tags TokenPrices
// @Accept json
// @Produce json
// @Param symbol path string true "Token Symbol (e.g., BTC)"
// @Success 200 {object} TokenPriceResponse "Token price data"
// @Failure 400 {object} errors.Vault0Error "Invalid symbol format"
// @Failure 404 {object} errors.Vault0Error "Token price not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /token-prices/{symbol} [get]
func (h *Handler) GetTokenPriceBySymbol(c *gin.Context) {
	var req GetTokenPriceRequest
	if err := c.ShouldBindUri(&req); err != nil {
		c.Error(errors.NewInvalidParameterError("symbol", "invalid symbol format in path"))
		return
	}

	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))

	// Call the service
	price, err := h.service.GetTokenPriceBySymbol(c.Request.Context(), symbol)
	if err != nil {
		c.Error(err)
		return
	}

	// Map to response DTO
	response := mapModelToResponse(price)
	c.JSON(http.StatusOK, response)
}

// mapModelToResponse converts a service layer TokenPrice model to an API response DTO.
func mapModelToResponse(model *tokensvc.TokenPrice) TokenPriceResponse {
	return TokenPriceResponse{
		Symbol:       model.Symbol,
		Rank:         model.Rank,
		PriceUSD:     model.PriceUSD,
		Supply:       model.Supply,
		MarketCapUSD: model.MarketCapUSD,
		VolumeUSD24h: model.VolumeUSD24h,
		UpdatedAt:    model.UpdatedAt,
	}
}
