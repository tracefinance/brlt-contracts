package pricefeed

import (
	"strings"

	"vault0/internal/config"
	"vault0/internal/errors"
	"vault0/internal/logger"
)

// NewPriceFeed creates a PriceFeedProvider instance based on the configuration.
func NewPriceFeed(cfg *config.Config, log logger.Logger) (PriceFeed, error) {
	providerName := strings.ToLower(strings.TrimSpace(cfg.PriceFeed.Provider))
	log = log.With(logger.String("provider_name", providerName))

	switch providerName {
	case "coincap":
		log.Info("Initializing CoinCap price feed provider")
		provider, err := NewCoinCapPriceFeed(cfg.PriceFeed, log)
		if err != nil {
			log.Error("Failed to initialize CoinCap provider", logger.Error(err))
			// Return the original configuration error from NewCoinCapProvider
			return nil, err
		}
		return provider, nil
	// Add cases for other providers like "coingecko" here in the future
	default:
		log.Warn("Unsupported price feed provider configured")
		return nil, errors.NewPriceFeedProviderNotSupported(cfg.PriceFeed.Provider)
	}
}
