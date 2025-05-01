package tokenprice

import (
	"context"
	"time"

	"vault0/internal/config"
	"vault0/internal/core/pricefeed"
	"vault0/internal/errors"
	"vault0/internal/logger"
)

type PricePoolingService interface {
	// RefreshTokenPrices fetches the latest token prices from the configured
	// external provider and updates the local database.
	//
	// Returns:
	//   - The number of tokens updated/inserted.
	//   - An error if fetching or database update fails (e.g., ErrPriceFeedUpdateFailed).
	RefreshTokenPrices(ctx context.Context) (int64, error)

	// StartPricePolling starts a background scheduler that periodically refreshes
	// token prices at an interval specified in the configuration.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the job
	StartPricePolling(ctx context.Context)

	// StopPricePolling stops the price update scheduler
	StopPricePolling()
}

type pollingService struct {
	repository Repository
	provider   pricefeed.PriceFeed
	log        logger.Logger
	config     *config.Config

	jobCtx    context.Context
	jobCancel context.CancelFunc
}

func NewPollingService(repo Repository, provider pricefeed.PriceFeed, log logger.Logger, cfg *config.Config) PricePoolingService {
	return &pollingService{
		repository: repo,
		provider:   provider,
		log:        log,
		config:     cfg,
	}
}

// StartPricePolling starts a background scheduler that periodically refreshes token prices
func (s *pollingService) StartPricePolling(ctx context.Context) {
	// Get interval from config with fallback to default
	interval := 300 // Default to 5 minutes if not specified
	if s.config.PriceFeed.RefreshInterval > 0 {
		interval = s.config.PriceFeed.RefreshInterval
	}

	s.jobCtx, s.jobCancel = context.WithCancel(ctx)

	s.log.Info("Starting token price update scheduler",
		logger.Int("interval_seconds", interval))

	// Start the scheduler goroutine
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// Immediately run once at startup
		if err := s.updatePrices(s.jobCtx); err != nil {
			s.log.Error("Initial token price update failed", logger.Error(err))
		}

		for {
			select {
			case <-s.jobCtx.Done():
				s.log.Info("Token price update scheduler stopped")
				return
			case <-ticker.C:
				if err := s.updatePrices(s.jobCtx); err != nil {
					s.log.Error("Scheduled token price update failed", logger.Error(err))
				}
			}
		}
	}()
}

// StopPricePolling stops the price update scheduler
func (s *pollingService) StopPricePolling() {
	if s.jobCancel != nil {
		s.jobCancel()
		s.jobCancel = nil
		s.log.Info("Token price update scheduler stopped")
	}
}

// updatePrices is a helper method that calls RefreshTokenPrices and logs the results
func (s *pollingService) updatePrices(ctx context.Context) error {
	s.log.Info("Running scheduled token price update")

	// Call the existing RefreshTokenPrices method
	count, err := s.RefreshTokenPrices(ctx)
	if err != nil {
		s.log.Error("Failed to refresh token prices", logger.Error(err))
		return err
	}

	s.log.Info("Completed token price update",
		logger.Int64("tokens_updated", count))

	return nil
}

// RefreshTokenPrices implements the Service interface.
func (s *pollingService) RefreshTokenPrices(ctx context.Context) (int64, error) {
	s.log.Info("Refreshing token prices from provider")

	providerData, err := s.provider.GetTokenPrices(ctx)
	if err != nil {
		s.log.Error("Failed to fetch token prices", logger.Error(err))
		return 0, errors.NewPriceFeedUpdateFailed(err, "failed to fetch data from provider")
	}

	if len(providerData) == 0 {
		s.log.Info("No token price data received from provider")
		return 0, nil
	}

	var pricesToUpsert []*TokenPrice
	for _, data := range providerData {
		price, err := convertProviderDataToTokenPrice(data)
		if err != nil {
			s.log.Warn("Failed to convert price data",
				logger.String("symbol", data.Symbol),
				logger.Error(err))
			continue
		}
		pricesToUpsert = append(pricesToUpsert, price)
	}

	if len(pricesToUpsert) == 0 {
		s.log.Warn("No valid token prices to update")
		return 0, nil
	}

	affected, err := s.repository.UpsertMany(ctx, pricesToUpsert)
	if err != nil {
		s.log.Error("Failed to update token prices", logger.Error(err))
		return 0, err
	}

	s.log.Info("Successfully refreshed token prices",
		logger.Int("total_fetched", len(providerData)),
		logger.Int("total_valid", len(pricesToUpsert)),
		logger.Int64("rows_affected", affected))

	return affected, nil
}
