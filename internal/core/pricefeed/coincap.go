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

// CoinCapPriceData represents the price data structure returned by CoinCap API.
type CoinCapPriceData struct {
	ID           string  `json:"id"`
	Symbol       string  `json:"symbol"`
	Name         string  `json:"name"`
	Rank         int     `json:"rank,string"`
	PriceUSD     float64 `json:"priceUsd,string"`
	Supply       float64 `json:"supply,string"`
	MarketCapUSD float64 `json:"marketCapUsd,string"`
	VolumeUSD24h float64 `json:"volumeUsd24Hr,string"`
}

// coinCapResponse represents the top-level structure of the CoinCap API response.
type coinCapResponse struct {
	Data      []*CoinCapPriceData `json:"data"`
	Timestamp int64               `json:"timestamp"`
}

// CoinCapPriceFeed implements the PriceFeedProvider interface for the CoinCap API.
type CoinCapPriceFeed struct {
	httpClient *http.Client
	apiURL     string
	apiKey     string
	limit      int
	log        logger.Logger
}

// NewCoinCapPriceFeed creates a new CoinCapProvider instance.
func NewCoinCapPriceFeed(cfg config.PriceFeedConfig, log logger.Logger) (*CoinCapPriceFeed, error) {
	if cfg.APIURL == "" {
		return nil, errors.NewConfigurationError("CoinCap API URL is required")
	}
	if cfg.Limit <= 0 {
		cfg.Limit = 100 // Default limit
	}

	return &CoinCapPriceFeed{
		httpClient: &http.Client{Timeout: 10 * time.Second}, // Add a timeout
		apiURL:     cfg.APIURL,
		apiKey:     cfg.APIKey,
		limit:      cfg.Limit,
		log:        log.With(logger.String("provider", "coincap")),
	}, nil
}

// mapToTokenPriceData converts CoinCapPriceData to TokenPriceData abstraction
func (p *CoinCapPriceFeed) mapToTokenPriceData(data *CoinCapPriceData, updatedAt time.Time) *TokenPriceData {
	return &TokenPriceData{
		ID:           data.ID,
		Symbol:       data.Symbol,
		Name:         data.Name,
		Rank:         data.Rank,
		PriceUSD:     data.PriceUSD,
		Supply:       data.Supply,
		MarketCapUSD: data.MarketCapUSD,
		VolumeUSD24h: data.VolumeUSD24h,
		UpdatedAt:    updatedAt,
	}
}

// GetTokenPrices fetches token prices from the CoinCap API.
func (p *CoinCapPriceFeed) GetTokenPrices(ctx context.Context) ([]*TokenPriceData, error) {
	// Build the URL with query parameters
	urlVal, err := url.Parse(p.apiURL)
	if err != nil {
		p.log.Error("Failed to parse CoinCap API URL", logger.Error(err), logger.String("url", p.apiURL))
		// Return a specific configuration error if URL parsing fails
		return nil, errors.NewConfigurationError(fmt.Sprintf("invalid CoinCap API URL: %s", p.apiURL))
	}

	query := urlVal.Query()
	query.Set("limit", strconv.Itoa(p.limit))
	if p.apiKey != "" {
		query.Set("apiKey", p.apiKey)
	}
	urlVal.RawQuery = query.Encode()
	fullURL := urlVal.String()

	p.log.Debug("Fetching token prices from CoinCap", logger.String("url", fullURL))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		p.log.Error("Failed to create CoinCap request", logger.Error(err))
		return nil, errors.NewPriceFeedRequestFailed(err, "failed to create HTTP request")
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		p.log.Error("Failed to execute CoinCap request", logger.Error(err))
		return nil, errors.NewPriceFeedRequestFailed(err, "failed to execute HTTP request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.log.Error("CoinCap API returned non-OK status", logger.Int("status_code", resp.StatusCode))
		// You might want to read the body here for more error details if available
		return nil, errors.NewPriceFeedRequestFailed(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "API returned non-OK status")
	}

	var apiResponse coinCapResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		p.log.Error("Failed to decode CoinCap API response", logger.Error(err))
		return nil, errors.NewInvalidPriceFeedResponse(err, "failed to decode JSON response")
	}

	// Set the UpdatedAt timestamp for each token and deduplicate by symbol
	now := time.Now()
	uniqueTokens := make(map[string]*TokenPriceData)
	for _, token := range apiResponse.Data {
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

	p.log.Info("Successfully fetched unique token prices from CoinCap",
		logger.Int("total_count", len(apiResponse.Data)),
		logger.Int("unique_count", len(result)))
	return result, nil
}
