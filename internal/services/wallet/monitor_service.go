package wallet

import (
	"context"
	"vault0/internal/logger"
	"vault0/internal/types"
)

type MonitorService interface {
	// StartTransactionMonitoring starts monitoring transactions for all wallets.
	// It performs the following steps:
	// 1. Retrieves all active wallets
	// 2. Monitors each wallet's address for transactions
	// 3. Processes incoming transaction events and saves them to the database
	//
	// Parameters:
	//   - ctx: Context for the operation
	//
	// Returns:
	//   - error: Any error that occurred during setup
	StartTransactionMonitoring(ctx context.Context) error

	// StopTransactionMonitoring stops monitoring transactions.
	// This should be called when shutting down the service.
	StopTransactionMonitoring()
}

// StartTransactionMonitoring starts monitoring transactions for all wallets
func (s *walletService) StartTransactionMonitoring(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already monitoring
	if s.monitorCtx != nil {
		s.log.Info("Transaction monitoring is already active")
		return nil
	}

	// Create a new context with cancel function for monitoring
	s.monitorCtx, s.monitorCancel = context.WithCancel(context.Background())

	// Start transaction event subscription if not already started
	s.txMonitor.SubscribeToTransactionEvents(s.monitorCtx)

	// Get all active wallets
	wallets, err := s.repository.List(ctx, 0, "") // Get all wallets
	if err != nil {
		s.monitorCancel()
		s.monitorCtx = nil
		s.monitorCancel = nil
		return err
	}

	// Monitor each wallet's address
	for _, wallet := range wallets.Items {
		if err := s.monitorAddress(ctx, wallet.ChainType, wallet.Address); err != nil {
			s.log.Warn("Failed to monitor wallet address",
				logger.String("address", wallet.Address),
				logger.String("chain_type", string(wallet.ChainType)),
				logger.Error(err))
		}
	}

	// Start a goroutine to process transaction events
	go s.processTransactionEvents(s.monitorCtx)

	s.log.Info("Started transaction monitoring",
		logger.Int("wallet_count", len(wallets.Items)))

	return nil
}

// StopTransactionMonitoring stops monitoring transactions
func (s *walletService) StopTransactionMonitoring() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.monitorCtx == nil {
		return // Not monitoring
	}

	// Cancel the monitoring context to stop the goroutine
	s.monitorCancel()

	// Reset context and cancel function
	s.monitorCtx = nil
	s.monitorCancel = nil

	s.log.Info("Stopped transaction monitoring")
}

// monitorAddress registers a wallet address for transaction monitoring
func (s *walletService) monitorAddress(ctx context.Context, chainType types.ChainType, address string) error {
	// Register with transaction service
	addr, err := types.NewAddress(chainType, address)
	if err != nil {
		return err
	}

	if err := s.txMonitor.MonitorAddress(ctx, addr); err != nil {
		return err
	}

	s.log.Debug("Started monitoring address for transactions",
		logger.String("address", address),
		logger.String("chain_type", string(chainType)))

	return nil
}

// unmonitorAddress stops monitoring a wallet address for transactions
func (s *walletService) unmonitorAddress(ctx context.Context, chainType types.ChainType, address string) error {
	// Unregister with transaction service
	addr, err := types.NewAddress(chainType, address)
	if err != nil {
		return err
	}

	if err := s.txMonitor.UnmonitorAddress(ctx, addr); err != nil {
		return err
	}

	s.log.Debug("Stopped monitoring address for transactions",
		logger.String("address", address),
		logger.String("chain_type", string(chainType)))

	return nil
}

// processTransactionEvents listens for transaction events and processes them
func (s *walletService) processTransactionEvents(ctx context.Context) {
	// Get the transaction events channel
	txEventsChan := s.txMonitor.TransactionEvents()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop processing
			return

		case tx, ok := <-txEventsChan:
			if !ok {
				// Channel closed, stop processing
				s.log.Warn("Transaction events channel closed")
				return
			}

			// Process the transaction
			s.handleTransaction(ctx, tx)
		}
	}
}

// handleTransaction processes a single transaction
func (s *walletService) handleTransaction(ctx context.Context, tx *types.Transaction) {
	if tx == nil {
		return
	}

	// Check if the transaction already exists by hash
	existingTx, _ := s.txService.GetTransactionByHash(ctx, tx.Hash)
	if existingTx != nil {
		s.log.Debug("Transaction already exists in database",
			logger.String("tx_hash", tx.Hash))
		return
	}

	// Transaction doesn't exist yet, log and update balances
	s.log.Info("Processing new transaction",
		logger.String("tx_hash", tx.Hash),
		logger.String("chain", string(tx.Chain)),
		logger.String("from", tx.From),
		logger.String("to", tx.To),
		logger.String("type", string(tx.Type)))

	// Find the wallet by chain and address (if possible)
	wallet, err := s.repository.GetByAddress(ctx, tx.Chain, tx.To)
	if err != nil || wallet == nil {
		// Try sender address if recipient not found
		wallet, err = s.repository.GetByAddress(ctx, tx.Chain, tx.From)
	}
	if err == nil && wallet != nil {
		// Save the transaction using the service method
		if err := s.txService.CreateWalletTransaction(ctx, wallet.ID, tx); err != nil {
			s.log.Error("Failed to create wallet transaction from monitor event",
				logger.String("tx_hash", tx.Hash),
				logger.Error(err))
		}
	}

	// Update wallet balances if the transaction was successful
	if tx.Status == types.TransactionStatusSuccess {
		switch tx.Type {
		case types.TransactionTypeNative:
			if err := s.UpdateWalletBalance(ctx, tx); err != nil {
				s.log.Error("Failed to update native balance from transaction",
					logger.String("tx_hash", tx.Hash),
					logger.Error(err))
			}
		case types.TransactionTypeERC20:
			if err := s.UpdateTokenBalance(ctx, tx); err != nil {
				s.log.Error("Failed to update token balance from transaction",
					logger.String("tx_hash", tx.Hash),
					logger.String("token_address", tx.TokenAddress),
					logger.Error(err))
			}
		}
	}
}
