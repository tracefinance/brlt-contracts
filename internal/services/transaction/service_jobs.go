package transaction

import (
	"context"
	"time"

	"vault0/internal/logger"
	"vault0/internal/types"
)

// StartWalletTransactionPolling starts a background scheduler that periodically polls for transactions from all active wallets
func (s *transactionService) StartWalletTransactionPolling(ctx context.Context) {
	// Get interval from config with fallback to default
	interval := 60 // Default to 1 minute if not specified
	if s.config.TransactionPollingInterval > 0 {
		interval = s.config.TransactionPollingInterval
	}

	s.pollingCtx, s.pollingCancel = context.WithCancel(ctx)

	s.log.Info("Starting transaction polling scheduler",
		logger.Int("interval_seconds", interval))

	// Start the scheduler goroutine
	go func() {
		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-s.pollingCtx.Done():
				s.log.Info("Transaction polling scheduler stopped")
				return
			case <-ticker.C:
				s.pollTransactionsForAllWallets(s.pollingCtx)
			}
		}
	}()
}

// StopWalletTransactionPolling stops the transaction polling scheduler
func (s *transactionService) StopWalletTransactionPolling() {
	if s.pollingCancel != nil {
		s.pollingCancel()
		s.pollingCancel = nil
		s.log.Info("Transaction polling scheduler stopped")
	}
}

// pollTransactionsForAllWallets fetches all active wallets and syncs their transactions
func (s *transactionService) pollTransactionsForAllWallets(ctx context.Context) {
	s.log.Info("Running scheduled transaction poll for all wallets")

	// Get all wallets
	walletPage, err := s.walletService.List(ctx, 0, 0)
	if err != nil {
		s.log.Error("Failed to list wallets for transaction polling",
			logger.Error(err))
		return
	}

	var totalSynced int
	for _, wallet := range walletPage.Items {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return
		}

		// Sync transactions for this wallet
		count, err := s.SyncTransactions(ctx, wallet.ID)
		if err != nil {
			s.log.Warn("Failed to sync transactions for wallet during polling",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.String("chain_type", string(wallet.ChainType)),
				logger.Error(err))
			continue
		}

		if count > 0 {
			s.log.Info("Synced transactions during polling",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.String("chain_type", string(wallet.ChainType)),
				logger.Int("transaction_count", count))
		}

		totalSynced += count
	}

	s.log.Info("Completed transaction polling cycle",
		logger.Int("total_wallets", len(walletPage.Items)),
		logger.Int("total_transactions_synced", totalSynced))
}

// StartPendingTransactionPolling starts a background scheduler that periodically polls for pending or mined transactions
func (s *transactionService) StartPendingTransactionPolling(ctx context.Context) {
	// Get interval from config with fallback to default
	interval := 60 // Default to 1 minute if not specified
	if s.config.PendingTransactionPollingInterval > 0 {
		interval = s.config.PendingTransactionPollingInterval
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

	// Prevent concurrent syncs
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	updatedCount := 0

	// Process each status
	for _, status := range statuses {
		// Check if context is cancelled
		if ctx.Err() != nil {
			return updatedCount, ctx.Err()
		}

		// Get transactions with the specified status across all chains
		filter := NewFilter().
			WithStatus(status).
			WithPagination(0, 0) // No pagination limit

		page, err := s.repository.List(ctx, filter)
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

				// Create transaction object with updated data
				updatedTransaction := FromCoreTransaction(updatedTx, tx.WalletID)

				// Preserve original metadata
				updatedTransaction.ID = tx.ID
				updatedTransaction.CreatedAt = tx.CreatedAt
				updatedTransaction.Timestamp = tx.Timestamp

				// Update in database
				err = s.repository.Update(ctx, updatedTransaction)
				if err != nil {
					s.log.Error("Failed to update transaction",
						logger.String("tx_hash", tx.Hash),
						logger.Error(err))
					continue
				}

				// Emit transaction event for the status change
				s.emitTransactionEvent(&TransactionEvent{
					WalletID:    tx.WalletID,
					Transaction: updatedTransaction,
					BlockNumber: updatedTx.BlockNumber.Int64(),
					EventType:   EventTypeTransactionDetected,
				})

				updatedCount++
			}
		}
	}

	s.log.Info("Completed pending transaction polling cycle",
		logger.Int("total_transactions_updated", updatedCount))

	return updatedCount, nil
}
