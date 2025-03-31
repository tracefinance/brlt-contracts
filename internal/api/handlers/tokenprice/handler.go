package tokenprice

import (
	"net/http"
	"strings"

	stdErrors "errors"
	"vault0/internal/errors"
	"vault0/internal/logger"
	tokensvc "vault0/internal/services/tokenprice"

	"github.com/gin-gonic/gin"
)

const (
	defaultLimit  = 50
	defaultOffset = 0
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
// @Param limit query int false "Maximum number of items to return" default(50) minimum(1) maximum(100)
// @Param offset query int false "Number of items to skip" default(0) minimum(0)
// @Param symbol query string false "Token symbol to filter by (can be used multiple times, e.g., ?symbol=BTC&symbol=ETH)"
// @Success 200 {object} PagedTokenPriceResponse "Paginated list of token prices"
// @Failure 400 {object} errors.Vault0Error "Invalid query parameters"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /token-prices [get]
func (h *Handler) ListTokenPrices(c *gin.Context) {
	var req ListTokenPricesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.logger.Warn("Failed to bind list token prices query params", logger.Error(err))
		handleAPIError(c, errors.NewInvalidParameterError("query", "invalid query parameters format or value"))
		return
	}

	// Set defaults if not provided
	limit := defaultLimit
	if req.Limit != nil {
		limit = *req.Limit
	}
	offset := defaultOffset
	if req.Offset != nil {
		offset = *req.Offset
	}

	// Normalize symbols (convert to uppercase)
	var symbols []string
	if len(req.Symbol) > 0 {
		symbols = make([]string, len(req.Symbol))
		for i, symbol := range req.Symbol {
			symbols[i] = strings.ToUpper(strings.TrimSpace(symbol))
		}
	}

	h.logger.Debug("Handling list token prices request",
		logger.Int("limit", limit),
		logger.Int("offset", offset),
		logger.Int("symbols_count", len(symbols)),
	)

	// Call the service with limit, offset, and optional symbols
	servicePage, err := h.service.ListTokenPrices(c.Request.Context(), limit, offset, symbols)
	if err != nil {
		h.logger.Error("Failed to list token prices from service", logger.Error(err))
		handleAPIError(c, err)
		return
	}

	// Convert service models to response DTOs
	responseItems := make([]TokenPriceResponse, len(servicePage.Items))
	for i, item := range servicePage.Items {
		responseItems[i] = mapModelToResponse(item)
	}

	// Create the response using concrete PagedTokenPriceResponse instead of generic Page
	responsePage := PagedTokenPriceResponse{
		Items:   responseItems,
		Offset:  servicePage.Offset,
		Limit:   servicePage.Limit,
		HasMore: servicePage.HasMore,
	}

	c.JSON(http.StatusOK, responsePage)
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
		h.logger.Warn("Failed to bind get token price path param", logger.Error(err))
		handleAPIError(c, errors.NewInvalidParameterError("symbol", "invalid symbol format in path"))
		return
	}

	symbol := strings.ToUpper(strings.TrimSpace(req.Symbol))
	h.logger.Debug("Handling get token price by symbol request", logger.String("symbol", symbol))

	// Call the service
	price, err := h.service.GetTokenPriceBySymbol(c.Request.Context(), symbol)
	if err != nil {
		h.logger.Error("Failed to get token price from service", logger.String("symbol", symbol), logger.Error(err))
		handleAPIError(c, err)
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

// handleAPIError maps Vault0Error types to HTTP status codes and sends JSON response.
func handleAPIError(c *gin.Context, err error) {
	var v0Error *errors.Vault0Error
	httpStatus := http.StatusInternalServerError // Default to 500

	if stdErrors.As(err, &v0Error) {
		switch v0Error.Code {
		// 400 Bad Request
		case errors.ErrCodeInvalidInput, errors.ErrCodeInvalidParameter, errors.ErrCodeMissingParameter, errors.ErrCodeValidationError, errors.ErrCodeDataConversionFailed:
			httpStatus = http.StatusBadRequest
		// 404 Not Found
		case errors.ErrCodeNotFound, errors.ErrCodeTokenPriceNotFound:
			httpStatus = http.StatusNotFound
		// 401 Unauthorized
		case errors.ErrCodeUnauthorized, errors.ErrCodeInvalidAccessToken, errors.ErrCodeAccessTokenExpired:
			httpStatus = http.StatusUnauthorized
		// 403 Forbidden
		case errors.ErrCodeForbidden:
			httpStatus = http.StatusForbidden
		// 500 Internal Server Error (Catch-all for specific codes)
		case errors.ErrCodeDatabaseError, errors.ErrCodeOperationFailed, errors.ErrCodeConfiguration,
			errors.ErrCodePriceFeedRequestFailed, errors.ErrCodeInvalidPriceFeedResponse,
			errors.ErrCodePriceFeedUpdateFailed:
			httpStatus = http.StatusInternalServerError
		// Default case for unmapped Vault0Error codes -> 500
		default:
			httpStatus = http.StatusInternalServerError
		}
		c.AbortWithStatusJSON(httpStatus, v0Error)
	} else {
		// Non-Vault0Error, treat as internal server error
		c.AbortWithStatusJSON(httpStatus, errors.NewInternalError(err))
	}
}
