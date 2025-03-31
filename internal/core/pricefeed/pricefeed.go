package pricefeed

import (
	"context"
	"time"
)

// TokenPriceData represents the price data for a single token fetched from a provider.
type TokenPriceData struct {
	ID           string    `json:"id"`                   // Provider's unique ID (e.g., "bitcoin")
	Symbol       string    `json:"symbol"`               // Token symbol (e.g., "BTC")
	Name         string    `json:"name"`                 // Token name (e.g., "Bitcoin")
	Rank         int       `json:"rank,string"`          // Market cap rank
	PriceUSD     float64   `json:"priceUsd,string"`      // Price in USD
	Supply       float64   `json:"supply,string"`        // Circulating supply
	MarketCapUSD float64   `json:"marketCapUsd,string"`  // Market cap in USD
	VolumeUSD24h float64   `json:"volumeUsd24Hr,string"` // Trading volume in last 24h
	UpdatedAt    time.Time `json:"-"`                    // Timestamp when the data was fetched/updated locally
}

// PriceFeed defines the interface for fetching token price data
// from external sources like CoinCap, CoinGecko, etc.
type PriceFeed interface {
	// GetTokenPrices fetches the latest price data for multiple tokens.
	// It typically fetches the top N tokens based on the provider's default sorting (e.g., market cap).
	// Implementations should handle provider-specific pagination or limits internally if necessary,
	// based on the configuration provided during initialization (e.g., config.PriceFeedConfig.Limit).
	//
	// Returns:
	//   - A slice of TokenPriceData
	//   - An error if the request fails, e.g., ErrPriceFeedRequestFailed, ErrInvalidPriceFeedResponse.
	GetTokenPrices(ctx context.Context) ([]*TokenPriceData, error)
}
