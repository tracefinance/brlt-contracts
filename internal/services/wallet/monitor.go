package wallet

import (
	"context"
	"math/big"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/transaction"
	"vault0/internal/errors"
	"vault0/internal/logger"
	txService "vault0/internal/services/transaction"
	"vault0/internal/types"
)

type WalletMonitor interface {

	// StartWalletMonitoring initializes monitoring for all non-deleted wallets
	// It fetches all non-deleted wallets and sets up monitoring for each wallet's address
	// using both transaction monitoring and transaction history services.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	StartWalletMonitoring(ctx context.Context) error

	// StopWalletMonitoring stops monitoring for all wallets.
	// It instructs both transaction monitoring and transaction history services
	// to stop monitoring all addresses.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//
	// Returns:
	//   - error: Any error that occurred during the operation
	StopWalletMonitoring(ctx context.Context) error
}

type walletMonitorService struct {
	log               logger.Logger
	repository        Repository
	blockchainFactory blockchain.Factory
	balanceService    BalanceService
	txMonitor         txService.MonitorService
	txHistory         txService.HistoryService
	txFactory         transaction.Factory
}

func NewWalletMonitorService(
	log logger.Logger,
	repository Repository,
	blockchainFactory blockchain.Factory,
	balanceService BalanceService,
	txMonitor txService.MonitorService,
	txHistory txService.HistoryService,
	txFactory transaction.Factory,
) WalletMonitor {
	return &walletMonitorService{log, repository, blockchainFactory, balanceService, txMonitor, txHistory, txFactory}
}

// StartWalletMonitoring initializes monitoring for all non-deleted wallets
func (s *walletMonitorService) StartWalletMonitoring(ctx context.Context) error {
	s.log.Info("Starting wallet monitoring...")

	// Get all non-deleted wallets (limit=0 to get all wallets)
	walletPage, err := s.repository.List(ctx, 0, "")
	if err != nil {
		s.log.Error("Failed to list wallets for monitoring",
			logger.Error(err))
		return errors.NewOperationFailedError("list wallets for monitoring", err)
	}

	if len(walletPage.Items) == 0 {
		s.log.Info("No wallets found for monitoring")
		return nil
	}

	s.log.Info("Found wallets to monitor",
		logger.Int("count", len(walletPage.Items)))

	var monitorErrors []error
	for _, wallet := range walletPage.Items {
		// Create a types.Address object - passing directly rather than a pointer
		address, err := types.NewAddress(wallet.ChainType, wallet.Address)
		if err != nil {
			s.log.Error("Failed to create address object for wallet",
				logger.Error(err),
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address))
			monitorErrors = append(monitorErrors, err)
			continue
		}

		// Monitor in transaction monitor service
		if err := s.txMonitor.MonitorAddress(*address); err != nil {
			s.log.Error("Failed to set up transaction monitoring for wallet",
				logger.Error(err),
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address))
			monitorErrors = append(monitorErrors, err)
			continue
		}

		// Monitor in transaction history service
		// Convert LastBlockNumber (int64) to *big.Int
		startBlockNumber := big.NewInt(wallet.LastBlockNumber + 1)

		if err := s.txHistory.MonitorAddress(*address, startBlockNumber); err != nil {
			s.log.Error("Failed to set up transaction history monitoring for wallet",
				logger.Error(err),
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.String("chain_type", string(wallet.ChainType)),
				logger.Int64("from_block", wallet.LastBlockNumber+1))
			monitorErrors = append(monitorErrors, err)
			continue
		}

		s.log.Info("Started monitoring wallet",
			logger.Int64("wallet_id", wallet.ID),
			logger.String("address", wallet.Address),
			logger.String("chain_type", string(wallet.ChainType)),
			logger.Int64("from_block", wallet.LastBlockNumber+1))
	}

	if len(monitorErrors) > 0 {
		s.log.Warn("Completed wallet monitoring setup with some errors",
			logger.Int("total_wallets", len(walletPage.Items)),
			logger.Int("error_count", len(monitorErrors)))
		// We don't return an error here since we want to continue monitoring the wallets that succeeded
	}

	s.log.Info("Successfully set up monitoring for all wallets",
		logger.Int("wallet_count", len(walletPage.Items)))

	// Start subscription to transaction events
	go s.subscribeToTransactionEvents(ctx)

	// Start subscription to history events
	go s.subscribeToHistoryEvents(ctx)

	return nil
}

// processTransactionEvent updates wallet and token balances based on a transaction event
// and optionally updates the last block number if updateBlockNumber is true
func (s *walletMonitorService) processTransactionEvent(ctx context.Context, event *txService.TransactionEvent, updateBlockNumber bool) {
	if event == nil || event.Transaction == nil {
		return
	}

	tx := event.Transaction.GetTransaction()
	if tx == nil {
		return
	}

	// Early return if transaction doesn't contain wallet ID metadata
	if !tx.Metadata.Contains(types.WalletIDMetadaKey) {
		return // Not relevant to our wallets
	}

	s.log.Debug("Processing transaction event",
		logger.String("tx_hash", tx.Hash),
		logger.String("from", tx.From),
		logger.String("to", tx.To),
		logger.Bool("is_new", event.IsNew),
		logger.Bool("update_block_number", updateBlockNumber))

	walletID, ok := tx.Metadata.GetInt64(types.WalletIDMetadaKey)
	if !ok {
		s.log.Error("Invalid or missing wallet ID in transaction metadata",
			logger.String("tx_hash", tx.Hash),
			logger.Any("metadata_wallet_id_key", types.WalletIDMetadaKey))
		return
	}

	// Skip balance update for existing transactions
	if !event.IsNew {
		s.log.Debug("Skipping balance update for existing transaction",
			logger.String("tx_hash", tx.Hash),
			logger.Int64("wallet_id", walletID))
	} else {
		// Only update balances for new transactions
		mapper, err := s.txFactory.NewDecoder(tx.ChainType)
		if err != nil {
			s.log.Error("Failed to create transaction mapper",
				logger.Error(err))
			return
		}

		decodedTx, err := mapper.DecodeTransaction(ctx, tx)
		if err != nil {
			s.log.Error("Failed to decode transaction",
				logger.Error(err),
				logger.String("tx_hash", tx.Hash))
			return
		}

		// Update balances based on transaction type
		switch decodedTx.GetType() {
		case types.TransactionTypeERC20Transfer:
			erc20Transfer, ok := decodedTx.(*types.ERC20Transfer)
			if !ok {
				s.log.Error("Failed to convert transaction to ERC20Transfer",
					logger.String("tx_hash", tx.Hash))
				return
			}

			// Update token balance
			err := s.balanceService.UpdateTokenBalance(ctx, erc20Transfer)
			if err != nil {
				s.log.Error("Failed to update token balance",
					logger.Error(err),
					logger.Int64("wallet_id", walletID),
					logger.String("tx_hash", tx.Hash),
					logger.String("token_address", erc20Transfer.TokenAddress))
			}

		default:
			// Update native balance
			err := s.balanceService.UpdateWalletBalance(ctx, tx)
			if err != nil {
				s.log.Error("Failed to update native wallet balance",
					logger.Error(err),
					logger.Int64("wallet_id", walletID),
					logger.String("tx_hash", tx.Hash))
			}
		}
	}

	// Update LastBlockNumber if requested and the transaction has a block number
	if updateBlockNumber && tx.BlockNumber != nil {
		wallet, err := s.repository.GetByID(ctx, walletID)
		if err != nil {
			s.log.Error("Failed to get wallet by ID for block number update",
				logger.Error(err),
				logger.Int64("wallet_id", walletID),
				logger.String("tx_hash", tx.Hash))
			return
		}

		if tx.BlockNumber.Int64() > wallet.LastBlockNumber {
			err := s.repository.UpdateBlockNumber(ctx, wallet.ID, tx.BlockNumber.Int64())
			if err != nil {
				s.log.Error("Failed to update wallet last block number",
					logger.Error(err),
					logger.Int64("wallet_id", wallet.ID),
					logger.Int64("block_number", tx.BlockNumber.Int64()))
				return
			}

			s.log.Debug("Updated wallet last block number",
				logger.Int64("wallet_id", wallet.ID),
				logger.Int64("old_block", wallet.LastBlockNumber),
				logger.Int64("new_block", tx.BlockNumber.Int64()))
		}
	}
}

// subscribeToTransactionEvents subscribes to real-time transaction events and updates wallet balances
func (s *walletMonitorService) subscribeToTransactionEvents(ctx context.Context) {
	s.log.Info("Starting subscription to transaction events")

	// Get the transaction events channel
	transactionEvents := s.txMonitor.TransactionEvents()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Stopping transaction events subscription due to context cancellation")
			return

		case eventData, ok := <-transactionEvents:
			if !ok {
				s.log.Warn("Transaction events channel closed, stopping subscription")
				return
			}
			if eventData == nil {
				s.log.Warn("Received nil event from transaction events channel")
				continue
			}
			s.processTransactionEvent(ctx, eventData, false)
		}
	}
}

// subscribeToHistoryEvents subscribes to historical transaction events and updates wallet balances and block numbers
func (s *walletMonitorService) subscribeToHistoryEvents(ctx context.Context) {
	s.log.Info("Starting subscription to history events")

	// Get the history events channel
	historyEvents := s.txHistory.HistoryEvents()

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Stopping history events subscription due to context cancellation")
			return

		case event, ok := <-historyEvents:
			if !ok {
				s.log.Warn("History events channel closed, stopping subscription")
				return
			}

			if event == nil {
				s.log.Warn("Received nil event from history events channel")
				continue
			}

			// Process history event - update block number for history events
			s.processTransactionEvent(ctx, event, true)
		}
	}
}

// StopWalletMonitoring stops monitoring for all wallets
func (s *walletMonitorService) StopWalletMonitoring(ctx context.Context) error {
	s.log.Info("Stopping wallet monitoring...")

	// Get all non-deleted wallets (limit=0 to get all wallets)
	walletPage, err := s.repository.List(ctx, 0, "")
	if err != nil {
		s.log.Error("Failed to list wallets to stop monitoring",
			logger.Error(err))
		return errors.NewOperationFailedError("list wallets to stop monitoring", err)
	}

	if len(walletPage.Items) == 0 {
		s.log.Info("No wallets found to stop monitoring")
		return nil
	}

	var unmonitorErrors []error
	for _, wallet := range walletPage.Items {
		// Create a types.Address object
		address, err := types.NewAddress(wallet.ChainType, wallet.Address)
		if err != nil {
			s.log.Error("Failed to create address object for wallet",
				logger.Error(err),
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address))
			unmonitorErrors = append(unmonitorErrors, err)
			continue
		}

		// Unmonitor in transaction monitor service
		if err := s.txMonitor.UnmonitorAddress(*address); err != nil {
			s.log.Error("Failed to stop transaction monitoring for wallet",
				logger.Error(err),
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address))
			unmonitorErrors = append(unmonitorErrors, err)
			// Continue to try unmonitoring in history service even if this failed
		}

		// Unmonitor in transaction history service
		if err := s.txHistory.UnmonitorAddress(*address); err != nil {
			s.log.Error("Failed to stop transaction history monitoring for wallet",
				logger.Error(err),
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address))
			unmonitorErrors = append(unmonitorErrors, err)
			continue
		}

		s.log.Info("Stopped monitoring wallet",
			logger.Int64("wallet_id", wallet.ID),
			logger.String("address", wallet.Address),
			logger.String("chain_type", string(wallet.ChainType)))
	}

	if len(unmonitorErrors) > 0 {
		s.log.Warn("Completed stopping wallet monitoring with some errors",
			logger.Int("total_wallets", len(walletPage.Items)),
			logger.Int("error_count", len(unmonitorErrors)))
		// We don't return an error here since we want to continue with other operations
		return nil
	}

	s.log.Info("Successfully stopped monitoring for all wallets",
		logger.Int("wallet_count", len(walletPage.Items)))
	return nil
}
