package wallet

import (
	"context"
	"time"

	"vault0/internal/core/blockexplorer"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

// HistoryService defines the interface for wallet transaction history management
type HistoryService interface {
	// SyncWallet fetches transaction history for a wallet from the blockchain explorer
	// and stores it in the transaction repository. It also updates the wallet's last
	// block number to facilitate incremental syncing.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - walletID: ID of the wallet to sync history for
	//
	// Returns:
	//   - int: Number of new transactions synced
	//   - error: ErrWalletNotFound if wallet doesn't exist, or other errors
	SyncWallet(ctx context.Context, walletID int64) (int, error)

	// SyncWalletByAddress fetches transaction history for a wallet by its address
	// and stores it in the transaction repository.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - address: The wallet's blockchain address
	//
	// Returns:
	//   - int: Number of new transactions synced
	//   - error: ErrWalletNotFound if wallet doesn't exist, or other errors
	SyncWalletByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error)

	// StartWalletHistorySyncing starts a background job that periodically syncs
	// transaction history for all wallets.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the syncing job
	//
	// Returns:
	//   - error: Any error that occurred during setup
	StartWalletHistorySyncing(ctx context.Context) error

	// StopWalletHistorySyncing stops the wallet history syncing job.
	// This should be called when shutting down the service.
	StopWalletHistorySyncing()
}

// SyncWallet fetches transaction history for a wallet and stores it in the database
func (s *walletService) SyncWallet(ctx context.Context, walletID int64) (int, error) {
	if walletID == 0 {
		return 0, errors.NewInvalidInputError("Wallet ID is required", "wallet_id", "0")
	}

	// Get the wallet by ID
	wallet, err := s.repository.GetByID(ctx, walletID)
	if err != nil {
		return 0, err
	}

	// Get the blockchain explorer for this chain type
	explorer, err := s.blockExplorerFactory.GetExplorer(wallet.ChainType)
	if err != nil {
		return 0, err
	}

	// Create options for transaction history query
	options := blockexplorer.TransactionHistoryOptions{
		StartBlock:    wallet.LastBlockNumber + 1, // Start from the next block after last synced
		EndBlock:      0,                          // Up to the latest block
		Page:          1,                          // Start with the first page
		PageSize:      10000,                      // Get 10000 transactions per page
		SortAscending: true,                       // Oldest to newest
	}

	s.log.Info("Syncing wallet transaction history",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("address", wallet.Address),
		logger.String("chain_type", string(wallet.ChainType)),
		logger.Int64("start_block", options.StartBlock))

	totalSynced := 0
	highestBlockNumber := wallet.LastBlockNumber
	hasMore := true

	// Loop until we have no more transactions to fetch
	for hasMore {
		// Check if context has been canceled
		if ctx.Err() != nil {
			return totalSynced, ctx.Err()
		}

		// Fetch transaction history from the explorer
		txPage, err := explorer.GetTransactionHistory(ctx, wallet.Address, options)
		if err != nil {
			s.log.Error("Failed to fetch transaction history",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.Int("page", options.Page),
				logger.Error(err))

			return totalSynced, err
		}

		if len(txPage.Items) == 0 {
			// No more transactions to process
			break
		}

		// Process each transaction
		for _, tx := range txPage.Items {
			// Convert core transaction to service transaction model
			serviceTransaction := transaction.FromCoreTransaction(tx, wallet.ID)

			// Update the highest block number we've seen
			if tx.BlockNumber != nil && tx.BlockNumber.Int64() > highestBlockNumber {
				highestBlockNumber = tx.BlockNumber.Int64()
			}

			// Check if transaction already exists
			exists, err := s.txService.GetTransaction(ctx, tx.Hash)
			if err == nil && exists != nil {
				// Transaction already exists, skip
				continue
			}

			// Save the transaction to the database
			err = s.txRepository.Create(ctx, serviceTransaction)
			if err != nil {
				s.log.Error("Failed to save transaction",
					logger.String("tx_hash", tx.Hash),
					logger.Error(err))
				continue
			}

			totalSynced++

			// If transaction was successful, update balances
			if tx.Status == types.TransactionStatusSuccess {
				if tx.Type == types.TransactionTypeNative {
					if err := s.UpdateWalletBalance(ctx, tx); err != nil {
						s.log.Error("Failed to update native balance from transaction",
							logger.String("tx_hash", tx.Hash),
							logger.Error(err))
					}
				} else if tx.Type == types.TransactionTypeERC20 {
					if err := s.UpdateTokenBalance(ctx, tx); err != nil {
						s.log.Error("Failed to update token balance from transaction",
							logger.String("tx_hash", tx.Hash),
							logger.String("token_address", tx.TokenAddress),
							logger.Error(err))
					}
				}
			}
		}

		// Check if we need to fetch more pages
		hasMore = txPage.HasMore
		options.Page++
	}

	// Only update the last block number if we've synced transactions and found a higher block
	if highestBlockNumber > wallet.LastBlockNumber {
		wallet.LastBlockNumber = highestBlockNumber
		if err := s.repository.Update(ctx, wallet); err != nil {
			s.log.Error("Failed to update wallet last block number",
				logger.Int64("wallet_id", wallet.ID),
				logger.Int64("last_block_number", highestBlockNumber),
				logger.Error(err))
		}
	}

	s.log.Info("Wallet transaction history sync completed",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("address", wallet.Address),
		logger.Int("transactions_synced", totalSynced),
		logger.Int64("last_block_number", highestBlockNumber))

	return totalSynced, nil
}

// SyncWalletByAddress fetches transaction history for a wallet by its address
func (s *walletService) SyncWalletByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error) {
	if chainType == "" {
		return 0, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return 0, errors.NewInvalidInputError("Address is required", "address", "")
	}

	// Get the wallet by address
	wallet, err := s.repository.GetByAddress(ctx, chainType, address)
	if err != nil {
		return 0, err
	}

	// Use the SyncWallet method with the wallet ID
	return s.SyncWallet(ctx, wallet.ID)
}

// StartWalletHistorySyncing starts a background job that periodically syncs transaction history for all wallets
func (s *walletService) StartWalletHistorySyncing(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already syncing
	if s.syncHistoryCtx != nil {
		s.log.Info("Wallet history syncing is already active")
		return nil
	}

	// Get the syncing interval from config
	interval := 60 * 10 // Default to 10 minutes if not specified
	if s.config.TransactionSynchingInterval > 0 {
		interval = s.config.TransactionSynchingInterval
	}

	// Create a new context with cancel function for syncing
	s.syncHistoryCtx, s.syncHistoryCancel = context.WithCancel(context.Background())

	// Start the syncing job in a goroutine
	go func() {
		s.log.Info("Starting wallet history syncing job",
			logger.Int("interval_seconds", interval))

		ticker := time.NewTicker(time.Duration(interval) * time.Second)
		defer ticker.Stop()

		// Run initial sync
		s.syncAllWallets(s.syncHistoryCtx)

		for {
			select {
			case <-s.syncHistoryCtx.Done():
				s.log.Info("Wallet history syncing job stopped")
				return
			case <-ticker.C:
				// Run periodic sync
				s.syncAllWallets(s.syncHistoryCtx)
			}
		}
	}()

	return nil
}

// StopWalletHistorySyncing stops the wallet history syncing job
func (s *walletService) StopWalletHistorySyncing() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.syncHistoryCtx == nil {
		return // Not syncing
	}

	// Cancel the syncing context to stop the goroutine
	s.syncHistoryCancel()

	// Reset context and cancel function
	s.syncHistoryCtx = nil
	s.syncHistoryCancel = nil

	s.log.Info("Stopped wallet history syncing job")
}

// syncAllWallets syncs transaction history for all wallets
func (s *walletService) syncAllWallets(ctx context.Context) {
	// Get all active wallets
	wallets, err := s.repository.List(ctx, 0, 0) // Get all wallets
	if err != nil {
		s.log.Error("Failed to get wallets for history sync", logger.Error(err))
		return
	}

	s.log.Info("Starting transaction history sync for all wallets",
		logger.Int("wallet_count", len(wallets.Items)))

	// Track total transactions synced
	totalTxSynced := 0

	// Process each wallet
	for _, wallet := range wallets.Items {
		// Check if context has been canceled
		if ctx.Err() != nil {
			s.log.Info("Wallet history sync stopped",
				logger.Int("transactions_synced", totalTxSynced))
			return
		}

		// Sync wallet transaction history
		txCount, err := s.SyncWallet(ctx, wallet.ID)
		if err != nil {
			s.log.Error("Failed to sync wallet history",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.String("chain_type", string(wallet.ChainType)),
				logger.Error(err))
			continue
		}

		totalTxSynced += txCount
	}

	s.log.Info("Completed transaction history sync for all wallets",
		logger.Int("wallet_count", len(wallets.Items)),
		logger.Int("total_transactions_synced", totalTxSynced))
}
