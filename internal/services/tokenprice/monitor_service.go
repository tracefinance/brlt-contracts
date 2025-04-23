package tokenprice

import (
	"context"
	"time"

	"vault0/internal/logger"
)

type MonitorService interface {
	// StartPricePolling starts a background scheduler that periodically refreshes
	// token prices at an interval specified in the configuration.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the job
	StartPricePolling(ctx context.Context)

	// StopPricePolling stops the price update scheduler
	StopPricePolling()
}

// StartPricePolling starts a background scheduler that periodically refreshes token prices
func (s *service) StartPricePolling(ctx context.Context) {
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
func (s *service) StopPricePolling() {
	if s.jobCancel != nil {
		s.jobCancel()
		s.jobCancel = nil
		s.log.Info("Token price update scheduler stopped")
	}
}

// updatePrices is a helper method that calls RefreshTokenPrices and logs the results
func (s *service) updatePrices(ctx context.Context) error {
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
