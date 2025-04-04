package pricefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"time"

	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"
)

// CoinPaprikaPriceData represents the price data structure returned by CoinPaprika API.
type CoinPaprikaPriceData struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Symbol      string  `json:"symbol"`
	Rank        int     `json:"rank"`
	TotalSupply float64 `json:"total_supply"`
	Quotes      struct {
		USD struct {
			Price     float64 `json:"price"`
			Volume24h float64 `json:"volume_24h"`
			MarketCap float64 `json:"market_cap"`
		} `json:"USD"`
	} `json:"quotes"`
}

// CoinPaprikaPriceFeed implements the PriceFeed interface for the CoinPaprika API.
type CoinPaprikaPriceFeed struct {
	httpClient *http.Client
	apiURL     string
	limit      int
	log        logger.Logger
}

// NewCoinPaprikaPriceFeed creates a new CoinPaprikaPriceFeed instance.
func NewCoinPaprikaPriceFeed(cfg config.PriceFeedConfig, log logger.Logger) (*CoinPaprikaPriceFeed, error) {
	if cfg.APIURL == "" {
		return nil, errors.NewConfigurationError("CoinPaprika API URL is required")
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 100 // Default limit
	}

	return &CoinPaprikaPriceFeed{
		httpClient: &http.Client{Timeout: 10 * time.Second}, // Add a timeout
		apiURL:     cfg.APIURL,
		limit:      cfg.Limit,
		log:        log.With(logger.String("provider", "coinpaprika")),
	}, nil
}

// mapToTokenPriceData converts CoinPaprikaPriceData to TokenPriceData abstraction
func (c *CoinPaprikaPriceFeed) mapToTokenPriceData(data *CoinPaprikaPriceData, updatedAt time.Time) *TokenPriceData {
	return &TokenPriceData{
		ID:           data.ID,
		Symbol:       data.Symbol,
		Name:         data.Name,
		Rank:         data.Rank,
		PriceUSD:     data.Quotes.USD.Price,
		Supply:       data.TotalSupply,
		MarketCapUSD: data.Quotes.USD.MarketCap,
		VolumeUSD24h: data.Quotes.USD.Volume24h,
		UpdatedAt:    updatedAt,
	}
}

// GetTokenPrices fetches token prices from the CoinPaprika API.
func (p *CoinPaprikaPriceFeed) GetTokenPrices(ctx context.Context) ([]*TokenPriceData, error) {
	// Build the URL with query parameters
	urlVal, err := url.Parse(p.apiURL)
	if err != nil {
		p.log.Error("Failed to parse CoinPaprika API URL", logger.Error(err), logger.String("url", p.apiURL))
		return nil, errors.NewConfigurationError(fmt.Sprintf("invalid CoinPaprika API URL: %s", p.apiURL))
	}

	query := urlVal.Query()
	query.Set("limit", strconv.Itoa(p.limit))
	urlVal.RawQuery = query.Encode()
	fullURL := urlVal.String()

	p.log.Debug("Fetching token prices from CoinPaprika", logger.String("url", fullURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		p.log.Error("Failed to create CoinPaprika request", logger.Error(err))
		return nil, errors.NewPriceFeedRequestFailed(err, "failed to create HTTP request")
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.log.Error("Failed to execute CoinPaprika request", logger.Error(err))
		return nil, errors.NewPriceFeedRequestFailed(err, "failed to execute HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.log.Error("CoinPaprika API returned non-OK status", logger.Int("status_code", resp.StatusCode))
		return nil, errors.NewPriceFeedRequestFailed(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "API returned non-OK status")
	}

	var apiResponse []*CoinPaprikaPriceData
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		p.log.Error("Failed to decode CoinPaprika API response", logger.Error(err))
		return nil, errors.NewInvalidPriceFeedResponse(err, "failed to decode JSON response")
	}

	// Set the UpdatedAt timestamp for each token and deduplicate by symbol
	now := time.Now()
	uniqueTokens := make(map[string]*TokenPriceData)
	for _, token := range apiResponse {
		// Map provider-specific data to the abstracted TokenPriceData
		abstractedToken := p.mapToTokenPriceData(token, now)
		// Keep only the first occurrence of each symbol (which should be the highest ranked one)
		if _, exists := uniqueTokens[abstractedToken.Symbol]; !exists {
			uniqueTokens[abstractedToken.Symbol] = abstractedToken
		}
	}

	// Convert map back to slice
	result := make([]*TokenPriceData, 0, len(uniqueTokens))
	for _, token := range uniqueTokens {
		result = append(result, token)
	}

	// Sort by rank
	sort.Slice(result, func(i, j int) bool {
		return result[i].Rank < result[j].Rank
	})

	p.log.Info("Successfully fetched unique token prices from CoinPaprika",
		logger.Int("total_count", len(apiResponse)),
		logger.Int("unique_count", len(result)))
	return result, nil
}
