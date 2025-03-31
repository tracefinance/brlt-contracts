package pricefeed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"
)

// coinCapResponse represents the top-level structure of the CoinCap API response.
type coinCapResponse struct {
	Data      []*TokenPriceData `json:"data"`
	Timestamp int64             `json:"timestamp"`
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

	// Set the UpdatedAt timestamp for each token
	now := time.Now()
	for _, token := range apiResponse.Data {
		token.UpdatedAt = now
	}

	p.log.Info("Successfully fetched token prices from CoinCap", logger.Int("count", len(apiResponse.Data)))
	return apiResponse.Data, nil
}
