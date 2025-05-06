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
		startBlockNumber := big.NewInt(wallet.LastBlockNumber)

		if err := s.txHistory.MonitorAddress(*address, startBlockNumber); err != nil {
			s.log.Error("Failed to set up transaction history monitoring for wallet",
				logger.Error(err),
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address))
			monitorErrors = append(monitorErrors, err)
			continue
		}

		s.log.Info("Started monitoring wallet",
			logger.Int64("wallet_id", wallet.ID),
			logger.String("address", wallet.Address),
			logger.String("chain_type", string(wallet.ChainType)),
			logger.Int64("from_block", wallet.LastBlockNumber))
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

			// Process transaction event - all events can be cast to Transaction
			if tx, ok := eventData.(*types.Transaction); ok {
				s.processTransactionEvent(ctx, tx)
			} else {
				s.log.Warn("Received unknown event type from transaction events channel")
				continue
			}
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

		case tx, ok := <-historyEvents:
			if !ok {
				s.log.Warn("History events channel closed, stopping subscription")
				return
			}

			if tx == nil {
				s.log.Warn("Received nil transaction from history events channel")
				continue
			}

			// Process history event
			s.processHistoryEvent(ctx, tx)
		}
	}
}

// processTransactionEvent updates wallet and token balances based on a transaction event
func (s *walletMonitorService) processTransactionEvent(ctx context.Context, tx *types.Transaction) {
	if tx == nil {
		return
	}

	s.log.Debug("Processing transaction event",
		logger.String("tx_hash", tx.Hash),
		logger.String("from", tx.From),
		logger.String("to", tx.To))

	// Check if this transaction affects any of our wallets
	if !tx.Metadata.Contains(types.WalletIDMetadaKey) {
		return // Not relevant to our wallets
	}

	walletID, ok := tx.Metadata.GetInt64(types.WalletIDMetadaKey)
	if !ok {
		s.log.Error("Invalid or missing wallet ID in transaction metadata",
			logger.String("tx_hash", tx.Hash),
			logger.Any("metadata_wallet_id_key", types.WalletIDMetadaKey))
		return
	}

	mapper, err := s.txFactory.NewDecoder(tx.ChainType)
	if err != nil {
		s.log.Error("Failed to create transaction mapper",
			logger.Error(err))
		return
	}

	// Update balances based on transaction type using balanceService
	switch tx.Type {
	case types.TransactionTypeNative:
		// Update native token balance using balanceService
		if err := s.balanceService.UpdateWalletBalance(ctx, tx); err != nil {
			s.log.Error("Failed to update native wallet balance",
				logger.Error(err),
				logger.Int64("wallet_id", walletID),
				logger.String("tx_hash", tx.Hash))
		}
	case types.TransactionTypeERC20Transfer:
		// For ERC20 transfers, we need to convert the transaction to an ERC20Transfer
		erc20Transfer, err := mapper.DecodeERC20Transfer(ctx, tx)
		if err != nil {
			s.log.Error("Failed to convert transaction to ERC20Transfer",
				logger.Error(err),
				logger.String("tx_hash", tx.Hash))
			return
		}

		if err := s.balanceService.UpdateTokenBalance(ctx, erc20Transfer); err != nil {
			s.log.Error("Failed to update token balance",
				logger.Error(err),
				logger.Int64("wallet_id", walletID),
				logger.String("tx_hash", tx.Hash),
				logger.String("token_address", erc20Transfer.TokenAddress))
		}
	default:
		// For other transaction types, just update native balance
		if err := s.balanceService.UpdateWalletBalance(ctx, tx); err != nil {
			s.log.Error("Failed to update native wallet balance",
				logger.Error(err),
				logger.Int64("wallet_id", walletID),
				logger.String("tx_hash", tx.Hash))
		}
	}
}

// processHistoryEvent updates wallet balances and LastBlockNumber based on a history event
func (s *walletMonitorService) processHistoryEvent(ctx context.Context, tx *types.Transaction) {
	if tx == nil {
		return
	}

	s.log.Debug("Processing history event",
		logger.String("tx_hash", tx.Hash),
		logger.String("from", tx.From),
		logger.String("to", tx.To),
		logger.Int64("block_number", tx.BlockNumber.Int64()))

	// Check if this transaction affects any of our wallets
	if !tx.Metadata.Contains(types.WalletIDMetadaKey) {
		return // Not relevant to our wallets
	}

	walletID, ok := tx.Metadata.GetInt64(types.WalletIDMetadaKey)
	if !ok {
		s.log.Error("Invalid or missing wallet ID in transaction metadata",
			logger.String("tx_hash", tx.Hash),
			logger.Any("metadata_wallet_id_key", types.WalletIDMetadaKey))
		return
	}

	mapper, err := s.txFactory.NewDecoder(tx.ChainType)
	if err != nil {
		s.log.Error("Failed to create transaction mapper",
			logger.Error(err))
		return
	}

	// Update balances based on transaction type using balanceService
	switch tx.Type {
	case types.TransactionTypeNative:
		// Update native token balance using balanceService
		if err := s.balanceService.UpdateWalletBalance(ctx, tx); err != nil {
			s.log.Error("Failed to update native wallet balance",
				logger.Error(err),
				logger.Int64("wallet_id", walletID),
				logger.String("tx_hash", tx.Hash))
		}
	case types.TransactionTypeERC20Transfer:
		// For ERC20 transfers, we need to convert the transaction to an ERC20Transfer
		erc20Transfer, err := mapper.DecodeERC20Transfer(ctx, tx)
		if err != nil {
			s.log.Error("Failed to convert transaction to ERC20Transfer",
				logger.Error(err),
				logger.String("tx_hash", tx.Hash))
			return
		}

		if err := s.balanceService.UpdateTokenBalance(ctx, erc20Transfer); err != nil {
			s.log.Error("Failed to update token balance",
				logger.Error(err),
				logger.Int64("wallet_id", walletID),
				logger.String("tx_hash", tx.Hash),
				logger.String("token_address", erc20Transfer.TokenAddress))
		}
	default:
		// For other transaction types, just update native balance
		if err := s.balanceService.UpdateWalletBalance(ctx, tx); err != nil {
			s.log.Error("Failed to update native wallet balance",
				logger.Error(err),
				logger.Int64("wallet_id", walletID),
				logger.String("tx_hash", tx.Hash))
		}
	}

	// Update LastBlockNumber if this transaction's block is newer
	// We need to get the wallet to know its current LastBlockNumber
	wallet, err := s.repository.GetByID(ctx, walletID)
	if err != nil {
		s.log.Error("Failed to get wallet by ID for block number update",
			logger.Error(err),
			logger.Int64("wallet_id", walletID),
			logger.String("tx_hash", tx.Hash))
		return
	}

	if tx.BlockNumber != nil && tx.BlockNumber.Int64() > wallet.LastBlockNumber {
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
