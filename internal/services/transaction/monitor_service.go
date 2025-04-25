package transaction

import (
	"context"
	"time"

	"vault0/internal/logger"
	"vault0/internal/types"
)

type MonitorService interface {
	// StartPendingTransactionPolling starts a background scheduler that periodically polls
	// for pending or mined transactions to update their status.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the polling
	StartPendingTransactionPolling(ctx context.Context)

	// StopPendingTransactionPolling stops the pending transaction polling scheduler
	StopPendingTransactionPolling()
}

// StartPendingTransactionPolling starts a background scheduler that periodically polls for pending or mined transactions
func (s *transactionService) StartPendingTransactionPolling(ctx context.Context) {
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

// updateTransactionStatus updates a transaction with new data from the explorer
func (s *transactionService) updateTransactionStatus(
	ctx context.Context,
	originalTx *Transaction,
	updatedTxData *types.Transaction,
) (*Transaction, error) {
	// Process the transaction to set token symbol and create Transaction model
	updatedTransaction := s.processTransaction(ctx, updatedTxData)

	// Preserve original metadata
	updatedTransaction.ID = originalTx.ID
	updatedTransaction.CreatedAt = originalTx.CreatedAt
	updatedTransaction.Timestamp = originalTx.Timestamp

	// Update in database
	err := s.repository.Update(ctx, updatedTransaction)
	if err != nil {
		s.log.Error("Failed to update transaction",
			logger.String("tx_hash", originalTx.Hash),
			logger.Error(err))
		return nil, err
	}

	return updatedTransaction, nil
}

// StopPendingTransactionPolling stops the pending transaction polling scheduler
func (s *transactionService) StopPendingTransactionPolling() {
	if s.pendingPollingCancel != nil {
		s.pendingPollingCancel()
		s.pendingPollingCancel = nil
		s.log.Info("Pending transaction polling scheduler stopped")
	}
}

// pollPendingOrMinedTransactions polls for pending or mined transactions and attempts to update their status
func (s *transactionService) pollPendingOrMinedTransactions(ctx context.Context) (int, error) {
	// Default statuses to look for
	statuses := []string{
		string(types.TransactionStatusPending),
		string(types.TransactionStatusMined),
	}

	s.log.Info("Running scheduled poll for pending and mined transactions",
		logger.Any("statuses", statuses))

	updatedCount := 0

	// Process each status
	for _, status := range statuses {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return updatedCount, ctx.Err()
		}

		// Get transactions with the specified status across all chains
		statusValue := status // Create a copy to use its address
		filter := &Filter{
			Status: &statusValue,
		}

		page, err := s.repository.List(ctx, filter, 0, "") // 0 limit means no pagination
		if err != nil {
			s.log.Error("Failed to get transactions by status",
				logger.String("status", status),
				logger.Error(err))
			continue
		}

		if len(page.Items) == 0 {
			s.log.Info("No transactions found with status",
				logger.String("status", status))
			continue
		}

		s.log.Info("Found transactions with status to check",
			logger.String("status", status),
			logger.Int("count", len(page.Items)))

		// Group transactions by chain type for more efficient explorer usage
		txsByChain := make(map[types.ChainType][]*Transaction)
		for _, tx := range page.Items {
			txsByChain[tx.ChainType] = append(txsByChain[tx.ChainType], tx)
		}

		// Process transactions by chain type
		for chainType, transactions := range txsByChain {
			// Get explorer for the chain
			explorer, err := s.blockExplorerFactory.GetExplorer(chainType)
			if err != nil {
				s.log.Error("Failed to get explorer",
					logger.String("chain_type", string(chainType)),
					logger.Error(err))
				continue
			}

			s.log.Info("Processing transactions",
				logger.String("chain_type", string(chainType)),
				logger.String("status", status),
				logger.Int("count", len(transactions)))

			// Process each transaction individually
			for _, tx := range transactions {
				// Fetch updated transaction details from explorer
				updatedTx, err := explorer.GetTransactionByHash(ctx, tx.Hash)
				if err != nil {
					s.log.Error("Failed to fetch transaction update",
						logger.String("tx_hash", tx.Hash),
						logger.String("chain_type", string(chainType)),
						logger.Error(err))
					continue
				}

				// Check if status has changed
				originalStatus := types.TransactionStatus(tx.Status)
				updatedStatus := updatedTx.Status

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

				// Update transaction with new status and data
				if _, err := s.updateTransactionStatus(ctx, tx, updatedTx); err != nil {
					// Error already logged in updateTransactionStatus
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
