package wallet

import (
	"context"
	"fmt"
	"time"

	"vault0/internal/core/blockexplorer"
	"vault0/internal/errors"
	"vault0/internal/logger"
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
	explorer, err := s.blockExplorerFactory.NewExplorer(wallet.ChainType)
	if err != nil {
		return 0, err
	}

	// Define base options for transaction history query
	baseOptions := blockexplorer.TransactionHistoryOptions{
		StartBlock:    wallet.LastBlockNumber + 1, // Start from the next block after last synced
		EndBlock:      0,                          // Up to the latest block
		SortAscending: true,                       // Oldest to newest
		Limit:         100,                        // Get 100 transactions per request
	}

	s.log.Info("Syncing wallet transaction history",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("address", wallet.Address),
		logger.String("chain_type", string(wallet.ChainType)),
		logger.Int64("start_block", baseOptions.StartBlock))

	totalSynced := 0
	highestBlockNumber := wallet.LastBlockNumber

	// First, sync normal (native) transactions
	totalSynced += s.syncNativeTransactions(ctx, explorer, wallet, baseOptions, &highestBlockNumber)

	// Then, sync ERC20 token transactions
	totalSynced += s.syncERC20Transactions(ctx, explorer, wallet, baseOptions, &highestBlockNumber)

	// Only update the last block number if we've found a higher block
	if highestBlockNumber > wallet.LastBlockNumber {
		if err := s.repository.UpdateBlockNumber(ctx, wallet.ID, highestBlockNumber); err != nil {
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

// syncNativeTransactions fetches and processes native transactions
func (s *walletService) syncNativeTransactions(
	ctx context.Context,
	explorer blockexplorer.BlockExplorer,
	wallet *Wallet,
	baseOptions blockexplorer.TransactionHistoryOptions,
	highestBlockNumber *int64,
) int {
	// Create options specific to native transactions
	options := baseOptions
	options.TransactionType = blockexplorer.TxTypeNormal

	s.log.Info("Syncing native transactions",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("address", wallet.Address),
		logger.String("transaction_type", string(blockexplorer.TxTypeNormal)))

	totalSynced := 0
	nextToken := ""
	hasMore := true

	// Loop until we have no more transactions to fetch
	for hasMore && totalSynced < 10000 { // Limit to 10000 transactions per type as a safety measure
		// Check if context has been canceled
		if ctx.Err() != nil {
			return totalSynced
		}

		// Fetch transaction history from the explorer
		txPage, err := explorer.GetTransactionHistory(ctx, wallet.Address, options, nextToken)
		if err != nil {
			s.log.Error("Failed to fetch native transaction history page",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.String("transaction_type", string(blockexplorer.TxTypeNormal)),
				logger.Error(err))
			return totalSynced
		}

		if len(txPage.Items) == 0 {
			break
		}

		for _, item := range txPage.Items {
			// Explicitly assert the type from the block explorer response
			rawTx, ok := item.(*types.Transaction)
			if !ok || rawTx == nil {
				s.log.Warn("Received non-transaction item or nil transaction from block explorer",
					logger.Int64("wallet_id", wallet.ID),
					logger.String("address", wallet.Address),
					logger.String("transaction_type", string(blockexplorer.TxTypeNormal)),
					logger.Any("item_type", fmt.Sprintf("%T", item)))
				continue
			}

			// Process the transaction using the shared method
			s.processTransaction(ctx, rawTx, wallet, highestBlockNumber, &totalSynced)
		}

		// Check if we need to fetch more pages
		if txPage.NextToken != "" {
			nextToken = txPage.NextToken
			hasMore = true
		} else {
			hasMore = false
		}
	}

	s.log.Info("Completed syncing native transactions",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("address", wallet.Address),
		logger.String("transaction_type", string(blockexplorer.TxTypeNormal)),
		logger.Int("transactions_synced", totalSynced))

	return totalSynced
}

// syncERC20Transactions fetches and processes ERC20 token transactions
func (s *walletService) syncERC20Transactions(
	ctx context.Context,
	explorer blockexplorer.BlockExplorer,
	wallet *Wallet,
	baseOptions blockexplorer.TransactionHistoryOptions,
	highestBlockNumber *int64,
) int {
	// Create options specific to ERC20 transactions
	options := baseOptions
	options.TransactionType = blockexplorer.TxTypeERC20

	s.log.Info("Syncing ERC20 transactions",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("address", wallet.Address),
		logger.String("transaction_type", string(blockexplorer.TxTypeERC20)))

	totalSynced := 0
	nextToken := ""
	hasMore := true

	// Loop until we have no more transactions to fetch
	for hasMore && totalSynced < 10000 { // Limit to 10000 transactions per type as a safety measure
		// Check if context has been canceled
		if ctx.Err() != nil {
			return totalSynced
		}

		// Fetch transaction history from the explorer
		txPage, err := explorer.GetTransactionHistory(ctx, wallet.Address, options, nextToken)
		if err != nil {
			s.log.Error("Failed to fetch ERC20 transaction history page",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.String("transaction_type", string(blockexplorer.TxTypeERC20)),
				logger.Error(err))
			return totalSynced
		}

		if len(txPage.Items) == 0 {
			break
		}

		for _, item := range txPage.Items {
			if item == nil {
				s.log.Warn("Received nil item from block explorer",
					logger.Int64("wallet_id", wallet.ID),
					logger.String("address", wallet.Address),
					logger.String("transaction_type", string(blockexplorer.TxTypeERC20)))
				continue
			}

			// Handle ERC20TxHistoryEntry case
			if erc20Entry, ok := item.(*blockexplorer.ERC20TxHistoryEntry); ok {
				s.processERC20HistoryEntry(ctx, erc20Entry, wallet, highestBlockNumber, &totalSynced)
				continue
			}

			// Handle basic Transaction case
			if rawTx, ok := item.(*types.Transaction); ok {
				s.processTransaction(ctx, rawTx, wallet, highestBlockNumber, &totalSynced)
				continue
			}

			// If we reach here, we have an unexpected type
			s.log.Warn("Received unexpected item type from block explorer",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.String("transaction_type", string(blockexplorer.TxTypeERC20)),
				logger.Any("item_type", fmt.Sprintf("%T", item)))
		}

		// Check if we need to fetch more pages
		if txPage.NextToken != "" {
			nextToken = txPage.NextToken
			hasMore = true
		} else {
			hasMore = false
		}
	}

	s.log.Info("Completed syncing ERC20 transactions",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("address", wallet.Address),
		logger.String("transaction_type", string(blockexplorer.TxTypeERC20)),
		logger.Int("transactions_synced", totalSynced))

	return totalSynced
}

// processERC20HistoryEntry processes an ERC20TxHistoryEntry and updates token balances
func (s *walletService) processERC20HistoryEntry(
	ctx context.Context,
	erc20Entry *blockexplorer.ERC20TxHistoryEntry,
	wallet *Wallet,
	highestBlockNumber *int64,
	totalSynced *int,
) {
	// Convert to ERC20Transfer
	erc20Transfer := erc20Entry.ToErc20Transfer()

	// Update highest block number if needed
	if erc20Transfer.BlockNumber != nil && erc20Transfer.BlockNumber.Int64() > *highestBlockNumber {
		*highestBlockNumber = erc20Transfer.BlockNumber.Int64()
	}

	// Check if transaction already exists
	existingTx, _ := s.txService.GetTransactionByHash(ctx, erc20Transfer.Hash)
	if existingTx != nil {
		return
	}

	// Save the transaction
	err := s.txService.CreateWalletTransaction(ctx, wallet.ID, &erc20Transfer.Transaction)
	if err != nil {
		s.log.Error("Failed to save ERC20 transaction",
			logger.String("tx_hash", erc20Transfer.Hash),
			logger.Error(err))
		return
	}

	// Increment counter
	*totalSynced++

	// Skip balance update for failed transactions
	if erc20Transfer.Status != types.TransactionStatusSuccess {
		return
	}

	// Update token balance
	if err := s.UpdateTokenBalance(ctx, erc20Transfer); err != nil {
		s.log.Error("Failed to update token balance from ERC20 transfer",
			logger.String("tx_hash", erc20Transfer.Hash),
			logger.String("token_address", erc20Transfer.TokenAddress),
			logger.Error(err))
	}
}

// processTransaction processes a Transaction that might be an ERC20 transfer
func (s *walletService) processTransaction(
	ctx context.Context,
	rawTx *types.Transaction,
	wallet *Wallet,
	highestBlockNumber *int64,
	totalSynced *int,
) {
	// Update highest block number if needed
	if rawTx.BlockNumber != nil && rawTx.BlockNumber.Int64() > *highestBlockNumber {
		*highestBlockNumber = rawTx.BlockNumber.Int64()
	}

	// Check if transaction already exists
	existingTx, _ := s.txService.GetTransactionByHash(ctx, rawTx.Hash)
	if existingTx != nil {
		return
	}

	// Save the transaction
	err := s.txService.CreateWalletTransaction(ctx, wallet.ID, rawTx)
	if err != nil {
		s.log.Error("Failed to save transaction",
			logger.String("tx_hash", rawTx.Hash),
			logger.Error(err))
		return
	}

	*totalSynced++

	if rawTx.Status != types.TransactionStatusSuccess {
		return
	}

	// Update balance
	if err := s.UpdateWalletBalance(ctx, rawTx); err != nil {
		s.log.Error("Failed to update wallet balance from transaction",
			logger.String("tx_hash", rawTx.Hash),
			logger.Error(err))
	}
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
	if s.config.Transaction.HistorySynchInterval > 0 {
		interval = s.config.Transaction.HistorySynchInterval
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
	wallets, err := s.repository.List(ctx, 0, "") // Get all wallets
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
