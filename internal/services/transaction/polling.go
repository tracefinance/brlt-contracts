package transaction

import (
	"context"
	"time"
	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/logger"
	"vault0/internal/types"
)

type PoolingService interface {
	// StartPendingTransactionPolling starts a background scheduler that periodically polls
	// for pending or mined transactions to update their status.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the polling
	StartPendingTransactionPolling(ctx context.Context)

	// StopPendingTransactionPolling stops the pending transaction polling scheduler
	StopPendingTransactionPolling()
}

// NewPoolingService creates a new transaction polling service
func NewPoolingService(
	config *config.Config,
	log logger.Logger,
	repository Repository,
	blockExplorerFactory blockexplorer.Factory,
) PoolingService {
	return &txPoolingService{
		config:               config,
		log:                  log,
		repository:           repository,
		blockExplorerFactory: blockExplorerFactory,
	}
}

type txPoolingService struct {
	// Polling lifecycle management
	pendingPollingCtx    context.Context
	pendingPollingCancel context.CancelFunc

	// Dependencies
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	blockExplorerFactory blockexplorer.Factory
}

// StartPendingTransactionPolling starts a background scheduler that periodically polls for pending or mined transactions
func (s *txPoolingService) StartPendingTransactionPolling(ctx context.Context) {
	// Get interval from config with fallback to default
	interval := 60 // Default to 1 minute if not specified
	if s.config.Transaction.TransactionUpdateInterval > 0 {
		interval = s.config.Transaction.TransactionUpdateInterval
	}

	s.pendingPollingCtx, s.pendingPollingCancel = context.WithCancel(ctx)

	s.log.Info("Starting pending transaction polling scheduler",
		logger.Int("interval_seconds", interval))

	// Start the scheduler goroutine
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.pendingPollingCtx.Done():
				s.log.Info("Pending transaction polling scheduler stopped")
				return
			case <-ticker.C:
				// Call the method with no parameters
				if _, err := s.pollPendingOrMinedTransactions(s.pendingPollingCtx); err != nil {
					s.log.Error("Error in pending transaction polling", logger.Error(err))
				}
			}
		}
	}()
}

// StopPendingTransactionPolling stops the pending transaction polling scheduler
func (s *txPoolingService) StopPendingTransactionPolling() {
	if s.pendingPollingCancel != nil {
		s.pendingPollingCancel()
		s.pendingPollingCancel = nil
		s.log.Info("Pending transaction polling scheduler stopped")
	}
}

// pollPendingOrMinedTransactions polls for pending or mined transactions and attempts to update their status
func (s *txPoolingService) pollPendingOrMinedTransactions(ctx context.Context) (int, error) {
	// Default statuses to look for
	statusesToCheck := []types.TransactionStatus{
		types.TransactionStatusPending,
		types.TransactionStatusMined,
	}

	s.log.Info("Running scheduled poll for pending and mined transactions",
		logger.Any("statuses", statusesToCheck))

	updatedCount := 0

	// Process each status
	for _, status := range statusesToCheck {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return updatedCount, ctx.Err()
		}

		// Get transactions with the specified status across all chains
		statusValue := status // Create a copy for pointer
		filter := &Filter{
			Status: &statusValue,
		}

		page, err := s.repository.List(ctx, filter, 0, "") // 0 limit means no pagination
		if err != nil {
			s.log.Error("Failed to get transactions by status",
				logger.String("status", string(status)),
				logger.Error(err))
			continue
		}

		if len(page.Items) == 0 {
			s.log.Info("No transactions found with status",
				logger.String("status", string(status)))
			continue
		}

		s.log.Info("Found transactions with status to check",
			logger.String("status", string(status)),
			logger.Int("count", len(page.Items)))

		// Group transactions by chain type for more efficient explorer usage
		txsByChain := make(map[types.ChainType][]*Transaction)
		for _, tx := range page.Items {
			txsByChain[tx.Chain] = append(txsByChain[tx.Chain], tx)
		}

		// Process transactions by chain type
		for chainType, transactions := range txsByChain {
			// Get explorer for the chain
			explorer, err := s.blockExplorerFactory.NewExplorer(chainType)
			if err != nil {
				s.log.Error("Failed to get explorer",
					logger.String("chain_type", string(chainType)),
					logger.Error(err))
				continue
			}

			s.log.Info("Processing transactions",
				logger.String("chain_type", string(chainType)),
				logger.String("status", string(status)),
				logger.Int("count", len(transactions)))

			// Process each transaction individually
			for _, tx := range transactions {
				updatedCoreTx, err := explorer.GetTransactionByHash(ctx, tx.Hash)
				if err != nil {
					s.log.Error("Failed to fetch transaction update",
						logger.String("tx_hash", tx.Hash),
						logger.String("chain_type", string(chainType)),
						logger.Error(err))
					continue
				}

				// Check if status has changed
				originalStatus := types.TransactionStatus(tx.Status)
				updatedStatus := updatedCoreTx.Status

				if originalStatus == updatedStatus {
					s.log.Debug("Transaction status unchanged",
						logger.String("tx_hash", tx.Hash),
						logger.String("status", string(updatedStatus)))
					continue
				}

				s.log.Info("Transaction status changed",
					logger.String("tx_hash", tx.Hash),
					logger.String("old_status", string(originalStatus)),
					logger.String("new_status", string(updatedStatus)))

				// Update transaction status using the repository method
				if err := s.repository.UpdateTransactionStatus(ctx, tx.Hash, updatedStatus); err != nil {
					// Error logging handled by repository or DB layer, but we log context here
					s.log.Error("Failed to update transaction status via repository",
						logger.String("tx_hash", tx.Hash),
						logger.String("new_status", string(updatedStatus)),
						logger.Error(err))
					continue
				}

				updatedCount++
			}
		}
	}

	s.log.Info("Completed pending transaction polling cycle",
		logger.Int("total_transactions_updated", updatedCount))

	return updatedCount, nil
}
