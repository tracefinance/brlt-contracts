package wallet

import (
	"context"
	"fmt"

	"vault0/internal/core/abiutils"
	"vault0/internal/core/tokenstore"
	"vault0/internal/core/transaction"
	"vault0/internal/logger"
	"vault0/internal/types"
)

type MonitorService interface {
	// StartTransactionMonitoring starts monitoring transactions for all wallets.
	// It performs the following steps:
	// 1. Retrieves all active wallets
	// 2. Monitors each wallet's address for transactions
	// 3. Subscribes to ERC20 token contracts for Transfer events
	// 4. Processes incoming transaction events and saves them to the database
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

	// Monitor all tokens in the token store for Transfer events
	if err := s.monitorAllTokens(ctx); err != nil {
		s.log.Warn("Failed to monitor all tokens", logger.Error(err))
		// Continue even if token monitoring setup fails
	}

	// Start listening for new tokens added to the token store
	go s.listenForTokenEvents(s.monitorCtx)

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

// monitorAllTokens monitors all tokens in the token store for Transfer events
func (s *walletService) monitorAllTokens(ctx context.Context) error {
	// Get all tokens from the token store
	tokens, err := s.tokenStore.ListTokens(ctx, 0, "")
	if err != nil {
		return err
	}

	// Set up monitoring for each token's Transfer events
	for _, token := range tokens.Items {
		if err := s.monitorTokenContract(ctx, token.ChainType, token.Address); err != nil {
			s.log.Warn("Failed to monitor token contract",
				logger.String("token_address", token.Address),
				logger.String("chain_type", string(token.ChainType)),
				logger.Error(err))
			// Continue with other tokens even if one fails
		}
	}

	s.log.Info("Started monitoring all token contracts for Transfer events",
		logger.Int("token_count", len(tokens.Items)))

	return nil
}

// monitorTokenContract sets up monitoring for a specific token contract
func (s *walletService) monitorTokenContract(ctx context.Context, chainType types.ChainType, tokenAddress string) error {
	// Create the address object for the token
	addr, err := types.NewAddress(chainType, tokenAddress)
	if err != nil {
		return err
	}

	// Monitor for Transfer events on the token contract
	events := []string{string(types.ERC20TransferEventSignature)}
	if err := s.txMonitor.MonitorContractAddress(ctx, addr, events); err != nil {
		return err
	}

	s.log.Debug("Started monitoring token contract for Transfer events",
		logger.String("token_address", tokenAddress),
		logger.String("chain_type", string(chainType)))

	return nil
}

// unmonitorTokenContract stops monitoring a token contract
func (s *walletService) unmonitorTokenContract(ctx context.Context, chainType types.ChainType, tokenAddress string) error {
	// Create the address object for the token
	addr, err := types.NewAddress(chainType, tokenAddress)
	if err != nil {
		return err
	}

	// Stop monitoring the token contract
	if err := s.txMonitor.UnmonitorContractAddress(ctx, addr); err != nil {
		return err
	}

	s.log.Debug("Stopped monitoring token contract",
		logger.String("token_address", tokenAddress),
		logger.String("chain_type", string(chainType)))

	return nil
}

// listenForTokenEvents listens for new tokens added to the token store
func (s *walletService) listenForTokenEvents(ctx context.Context) {
	tokenEvents := s.tokenStore.TokenEvents()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop processing
			return

		case event, ok := <-tokenEvents:
			if !ok {
				// Channel closed
				s.log.Warn("Token events channel closed")
				return
			}

			// Process the token event
			switch event.EventType {
			case tokenstore.TokenEventAdded:
				// New token added, start monitoring it
				if event.Token != nil {
					if err := s.monitorTokenContract(ctx, event.Token.ChainType, event.Token.Address); err != nil {
						s.log.Warn("Failed to monitor new token contract",
							logger.String("token_address", event.Token.Address),
							logger.String("chain_type", string(event.Token.ChainType)),
							logger.Error(err))
					}
				}

			case tokenstore.TokenEventDeleted:
				// Token deleted, stop monitoring it
				if event.Token != nil {
					if err := s.unmonitorTokenContract(ctx, event.Token.ChainType, event.Token.Address); err != nil {
						s.log.Warn("Failed to unmonitor token contract",
							logger.String("token_address", event.Token.Address),
							logger.String("chain_type", string(event.Token.ChainType)),
							logger.Error(err))
					}
				}

			case tokenstore.TokenEventUpdated:
				// Token updated, no action needed for monitoring
				s.log.Debug("Token updated in token store",
					logger.String("token_address", event.Token.Address),
					logger.String("chain_type", string(event.Token.ChainType)))
			}
		}
	}
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

// handleTransaction processes a single transaction received from the monitor
func (s *walletService) handleTransaction(ctx context.Context, tx *types.Transaction) {
	if tx == nil {
		return
	}

	// 1. Check if the transaction already exists by hash
	existingTx, _ := s.txService.GetTransactionByHash(ctx, tx.Hash)
	if existingTx != nil {
		// Already processed, maybe update status if needed, but likely can just return
		s.log.Debug("Transaction already processed", logger.String("tx_hash", tx.Hash))
		return
	}

	s.log.Info("Processing new transaction event",
		logger.String("tx_hash", tx.Hash),
		logger.String("chain", string(tx.Chain)),
		logger.String("type", string(tx.Type)), // Original type from monitor
	)

	abiUtils, err := abiutils.NewABIUtils(tx.Chain, s.config, s.blockExplorerFactory)
	if err != nil {
		s.log.Error("Failed to create ABI utils", logger.Error(err))
		return
	}

	txMapper, err := transaction.NewMapper(tx.Chain, s.tokenStore, s.log, abiUtils)
	if err != nil {
		s.log.Error("Failed to create transaction mapper", logger.Error(err))
		return
	}

	// 2. Attempt to map the transaction to a specific type
	mappedTx, mapErr := txMapper.ToTypedTransaction(ctx, tx)
	if mapErr != nil {
		// Log mapping error but continue processing as a generic transaction
		s.log.Warn("Failed to map transaction to specific type",
			logger.String("tx_hash", tx.Hash),
			logger.Error(mapErr))
	}

	// 3. Find associated wallet (check both 'to' and 'from')
	// Determine relevant addresses based on the mapped type or the original tx
	var relevantAddresses []string
	if mappedTx != nil {
		switch specificTx := mappedTx.(type) {
		case *types.ERC20Transfer:
			relevantAddresses = append(relevantAddresses, specificTx.From, specificTx.Recipient)
			// Potentially add specific logging for ERC20 Transfer
			s.log.Info("Mapped transaction to ERC20 Transfer",
				logger.String("tx_hash", tx.Hash),
				logger.String("token", specificTx.TokenAddress),
				logger.String("recipient", specificTx.Recipient),
				logger.String("amount", specificTx.Amount.String()))
		// Add cases for MultiSig types if they affect monitored wallets directly
		case *types.MultiSigWithdrawalRequest, *types.MultiSigSignWithdrawal, *types.MultiSigExecuteWithdrawal:
			// Add relevant addresses if needed (e.g., MultiSig contract address, recipient in execute)
			relevantAddresses = append(relevantAddresses, tx.From, tx.To) // Default to original for now
			s.log.Info("Mapped transaction to MultiSig operation",
				logger.String("tx_hash", tx.Hash),
				logger.Any("mapped_type", fmt.Sprintf("%T", specificTx)))
		default:
			// Unknown mapped type or simple contract call, use original addresses
			relevantAddresses = append(relevantAddresses, tx.From, tx.To)
		}
	} else {
		// Mapping failed or returned nil, use original addresses
		relevantAddresses = append(relevantAddresses, tx.From, tx.To)
	}

	// Find wallets associated with relevant addresses
	var associatedWallet *Wallet
	for _, addr := range relevantAddresses {
		if addr == "" { // Skip empty addresses
			continue
		}
		wallet, err := s.repository.GetByAddress(ctx, tx.Chain, addr)
		if err == nil && wallet != nil {
			associatedWallet = wallet
			break // Found the first associated wallet
		}
	}

	// 4. Save the original transaction if an associated wallet was found
	if associatedWallet != nil {
		if err := s.txService.CreateWalletTransaction(ctx, associatedWallet.ID, tx); err != nil {
			s.log.Error("Failed to create wallet transaction from monitor event",
				logger.String("tx_hash", tx.Hash),
				logger.Int64("wallet_id", associatedWallet.ID),
				logger.Error(err))
			// Decide if we should return here or still attempt balance update
		}
	} else {
		// No associated wallet found for this transaction
		s.log.Debug("Transaction does not involve any monitored wallets", logger.String("tx_hash", tx.Hash))
		return // Stop processing if no wallet is involved
	}

	// 5. Update wallet balances if the transaction was successful
	if tx.Status != types.TransactionStatusSuccess {
		s.log.Info("Transaction not successful, skipping balance update",
			logger.String("tx_hash", tx.Hash),
			logger.String("status", string(tx.Status)))
		return
	}

	// Use mapped transaction for balance updates if available and relevant
	balanceUpdated := false
	if mappedTx != nil {
		switch specificTx := mappedTx.(type) {
		case *types.ERC20Transfer:
			if err := s.UpdateTokenBalance(ctx, specificTx); err != nil {
				s.log.Error("Failed to update token balance from mapped ERC20 transfer",
					logger.String("tx_hash", tx.Hash),
					logger.Int64("wallet_id", associatedWallet.ID),
					logger.String("token_address", specificTx.TokenAddress),
					logger.Error(err))
			} else {
				balanceUpdated = true
			}
			// Add cases for other mapped types that affect balances if needed
		}
	}

	// Fallback to updating native balance if no specific balance was updated
	// and the transaction involves the associated wallet directly (sender/receiver)
	if !balanceUpdated && (tx.From == associatedWallet.Address || tx.To == associatedWallet.Address) {
		if tx.Type == types.TransactionTypeNative || mappedTx == nil { // Update native balance for native tx or unmapped tx
			if err := s.UpdateWalletBalance(ctx, tx); err != nil {
				s.log.Error("Failed to update native balance from transaction",
					logger.String("tx_hash", tx.Hash),
					logger.Int64("wallet_id", associatedWallet.ID), // Log the wallet ID found earlier
					logger.Error(err))
			}
		}
	}
}
