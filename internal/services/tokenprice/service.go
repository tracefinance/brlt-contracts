package tokenprice

import (
	"context" // Import standard errors package with an alias
	"strings"

	"vault0/internal/config"
	"vault0/internal/core/pricefeed"
	"vault0/internal/errors" // Custom errors package
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Service defines the interface for token price operations.
type Service interface {
	// GetTokenPriceBySymbol retrieves the stored price data for a specific token symbol.
	//
	// Returns:
	//   - A pointer to the TokenPrice if found.
	//   - ErrTokenPriceNotFound if no price data exists for the symbol.
	//   - Other errors propagated from the repository.
	GetTokenPriceBySymbol(ctx context.Context, symbol string) (*TokenPrice, error)

	// ListTokenPrices retrieves a paginated list of token prices, optionally filtered
	//
	// Parameters:
	//   - ctx: The context for the operation
	//   - filter: Optional filtering criteria
	//   - limit: Maximum number of items to return (0 for all items)
	//   - nextToken: Token for pagination (empty string for first page)
	//
	// Returns:
	//   - A page of token prices with pagination information
	//   - An error if the database operation fails
	ListTokenPrices(ctx context.Context, filter *TokenPriceFilter, limit int, nextToken string) (*types.Page[*TokenPrice], error)
}

type service struct {
	repository Repository
	provider   pricefeed.PriceFeed
	log        logger.Logger
	config     *config.Config
}

// NewService creates a new token price service instance.
func NewService(repo Repository, provider pricefeed.PriceFeed, log logger.Logger, cfg *config.Config) Service {
	return &service{
		repository: repo,
		provider:   provider,
		log:        log.With(logger.String("service", "tokenprice")),
		config:     cfg,
	}
}

// GetTokenPriceBySymbol implements the Service interface.
func (s *service) GetTokenPriceBySymbol(ctx context.Context, symbol string) (*TokenPrice, error) {
	symbol = strings.ToUpper(strings.TrimSpace(symbol))
	if symbol == "" {
		return nil, errors.NewInvalidInputError("Token symbol cannot be empty", "symbol", "")
	}

	s.log.Debug("Getting token price by symbol", logger.String("symbol", symbol))
	price, err := s.repository.GetBySymbol(ctx, symbol)
	if err != nil {
		s.log.Error("Failed to get token price by symbol", logger.String("symbol", symbol), logger.Error(err))
		return nil, err
	}
	return price, nil
}

// ListTokenPrices retrieves a paginated list of token prices, optionally filtered
//
// Parameters:
//   - ctx: The context for the operation
//   - filter: Optional filtering criteria
//   - limit: Maximum number of items to return (0 for all items)
//   - nextToken: Token for pagination (empty string for first page)
//
// Returns:
//   - A page of token prices with pagination information
//   - An error if the database operation fails
func (s *service) ListTokenPrices(ctx context.Context, filter *TokenPriceFilter, limit int, nextToken string) (*types.Page[*TokenPrice], error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 10
	}

	// Use the repository's List method with the filter
	page, err := s.repository.List(ctx, filter, limit, nextToken)
	if err != nil {
		return nil, err
	}

	return page, nil
}

// convertProviderDataToTokenPrice converts the data structure from the price feed provider
// to the service's internal TokenPrice model, returning a DataConversionFailed error on failure.
func convertProviderDataToTokenPrice(data *pricefeed.TokenPriceData) (*TokenPrice, error) {
	symbol := strings.ToUpper(strings.TrimSpace(data.Symbol))
	if symbol == "" {
		return nil, errors.NewDataConversionFailed(nil, "token symbol is empty", map[string]any{"provider_id": data.ID})
	}

	return &TokenPrice{
		Symbol:       symbol,
		Rank:         data.Rank,
		PriceUSD:     data.PriceUSD,
		Supply:       data.Supply,
		MarketCapUSD: data.MarketCapUSD,
		VolumeUSD24h: data.VolumeUSD24h,
		UpdatedAt:    data.UpdatedAt,
	}, nil
}
