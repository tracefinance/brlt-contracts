package transaction

import (
	"context"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// TransactionEvent represents an event related to a transaction
type TransactionEvent struct {
	WalletID    int64
	Transaction *Transaction
	BlockNumber int64
	EventType   string
}

// Event types for transaction events
const (
	EventTypeTransactionDetected = "TRANSACTION_DETECTED"
	EventTypeNewBlock            = "NEW_BLOCK"
)

// SubscribeToTransactionEvents starts listening for new blocks and processing transactions
func (s *transactionService) SubscribeToTransactionEvents(ctx context.Context) {
	s.eventCtx, s.eventCancel = context.WithCancel(ctx)

	// Get list of unique chain types from chains
	for _, chain := range s.chains.Chains {
		// Start a goroutine for each chain to subscribe to new blocks
		go s.subscribeToChainBlocks(s.eventCtx, chain)
	}

	// Continue listening for wallet lifecycle events
	go func() {
		eventCh := s.walletService.LifecycleEvents()
		for {
			select {
			case <-s.eventCtx.Done():
				return
			case event, ok := <-eventCh:
				if !ok {
					// Channel was closed
					return
				}

				switch event.EventType {
				case wallet.EventTypeWalletCreated:
					// When a new wallet is created, sync its historical transactions
					if err := s.handleWalletCreated(s.eventCtx, event); err != nil {
						s.log.Error("Failed to handle wallet created event",
							logger.Int64("wallet_id", event.WalletID),
							logger.Error(err))
					}

				case wallet.EventTypeWalletDeleted:
					// When a wallet is deleted, we don't need to do anything
					// The transactions will remain in the database for historical purposes
					s.log.Info("Wallet deleted",
						logger.Int64("wallet_id", event.WalletID))
				}
			}
		}
	}()
}

// UnsubscribeFromTransactionEvents stops listening for blockchain events
func (s *transactionService) UnsubscribeFromTransactionEvents() {
	if s.eventCancel != nil {
		s.eventCancel()
		s.eventCancel = nil
	}

	// Close the transaction events channel
	close(s.transactionEvents)
}

// TransactionEvents returns a channel that emits transaction events.
// These events include new transactions detected for monitored wallets.
// The channel is closed when UnsubscribeFromTransactionEvents is called.
func (s *transactionService) TransactionEvents() <-chan *TransactionEvent {
	return s.transactionEvents
}

// subscribeToChainBlocks subscribes to new blocks for a specific chain
func (s *transactionService) subscribeToChainBlocks(ctx context.Context, chain types.Chain) {
	// Get blockchain client for the chain type
	client, err := s.blockchainRegistry.GetBlockchain(chain.Type)
	if err != nil {
		s.log.Error("Failed to get blockchain client",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	s.log.Info("Starting new block subscription",
		logger.String("chain_type", string(chain.Type)))

	// Subscribe to new block headers
	blockCh, errCh, err := client.SubscribeNewHead(ctx)
	if err != nil {
		s.log.Error("Failed to subscribe to new blocks",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Process new blocks
	for {
		select {
		case <-ctx.Done():
			s.log.Info("Block subscription stopped",
				logger.String("chain_type", string(chain.Type)))
			return
		case err := <-errCh:
			s.log.Warn("Block subscription error",
				logger.String("chain_type", string(chain.Type)),
				logger.Error(err))
		case block := <-blockCh:
			s.processBlock(ctx, chain.Type, &block)
		}
	}
}

// processBlock processes a new block, looking for transactions for monitored wallets
func (s *transactionService) processBlock(ctx context.Context, chainType types.ChainType, block *types.Block) {
	// Emit block event
	s.emitTransactionEvent(&TransactionEvent{
		BlockNumber: block.Number.Int64(),
		EventType:   EventTypeNewBlock,
	})

	s.log.Debug("Processing new block",
		logger.String("chain_type", string(chainType)),
		logger.Int64("block_number", block.Number.Int64()),
		logger.String("block_hash", block.Hash),
		logger.Int("transaction_count", block.TransactionCount))

	// Get all wallets for this chain type
	wallets, err := s.getWalletsByChainType(ctx, chainType)
	if err != nil {
		s.log.Error("Failed to get wallets for chain type",
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
		return
	}

	// Create a map of wallet addresses for quick lookup
	addressToWallet := make(map[string]*wallet.Wallet)
	for _, w := range wallets {
		addressToWallet[w.Address] = w
	}

	// Process each transaction in the block
	for _, tx := range block.Transactions {
		// Check if any transaction is from or to a wallet we're monitoring
		var walletID int64

		// Check if the transaction involves any of our monitored wallets
		if tx.From != "" {
			if w, exists := addressToWallet[tx.From]; exists {
				walletID = w.ID
			}
		}

		if tx.To != "" {
			if w, exists := addressToWallet[tx.To]; exists {
				walletID = w.ID
			}
		}
		// If this transaction involves one of our wallets, save it
		if walletID > 0 {
			// Set the timestamp to the block timestamp
			tx.Timestamp = block.Timestamp.Unix()

			// Convert core transaction to service transaction
			transaction := FromCoreTransaction(tx, walletID)

			// Check if transaction already exists
			exists, err := s.repository.Exists(ctx, transaction.Hash)
			if err != nil {
				s.log.Error("Failed to check if transaction exists",
					logger.String("tx_hash", tx.Hash),
					logger.Int64("wallet_id", walletID),
					logger.Error(err))
				continue
			}

			// If transaction exists, get it from the database
			if exists {
				transaction, err = s.repository.GetByTxHash(ctx, transaction.Hash)
				if err != nil {
					s.log.Error("Failed to get existing transaction",
						logger.String("tx_hash", tx.Hash),
						logger.Int64("wallet_id", walletID),
						logger.Error(err))
					continue
				}
			} else {
				// Otherwise, save the new transaction
				err = s.repository.Create(ctx, transaction)
				if err != nil {
					s.log.Error("Failed to save transaction",
						logger.String("tx_hash", tx.Hash),
						logger.Int64("wallet_id", walletID),
						logger.Error(err))
					continue
				}
			}

			// Emit transaction event
			s.emitTransactionEvent(&TransactionEvent{
				WalletID:    walletID,
				Transaction: transaction,
				BlockNumber: block.Number.Int64(),
				EventType:   EventTypeTransactionDetected,
			})

			// Update last block number for the wallet
			wallet := addressToWallet[tx.From]
			if wallet == nil {
				wallet = addressToWallet[tx.To]
			}

			if wallet != nil && block.Number.Int64() > wallet.LastBlockNumber {
				if err := s.walletService.UpdateLastBlockNumber(ctx, wallet.ChainType, wallet.Address, block.Number.Int64()); err != nil {
					s.log.Error("Failed to update last block number",
						logger.Int64("wallet_id", wallet.ID),
						logger.String("address", wallet.Address),
						logger.Error(err))
				}
			}
		}
	}
}

// getWalletsByChainType retrieves all wallets for a specific chain type
func (s *transactionService) getWalletsByChainType(ctx context.Context, chainType types.ChainType) ([]*wallet.Wallet, error) {
	// Get all wallets
	walletPage, err := s.walletService.List(ctx, 0, 0)
	if err != nil {
		return nil, errors.NewOperationFailedError("list wallets", err)
	}

	// Filter wallets by chain type
	var wallets []*wallet.Wallet
	for _, wallet := range walletPage.Items {
		if wallet.ChainType == chainType {
			wallets = append(wallets, wallet)
		}
	}

	return wallets, nil
}

// emitTransactionEvent sends a transaction event to the transaction events channel
func (s *transactionService) emitTransactionEvent(event *TransactionEvent) {
	select {
	case s.transactionEvents <- event:
		s.log.Debug("Emitted transaction event",
			logger.String("event_type", event.EventType),
			logger.Int64("block_number", event.BlockNumber))
	default:
		s.log.Debug("Transaction events channel is full, dropping event",
			logger.String("event_type", event.EventType),
			logger.Int64("block_number", event.BlockNumber))
	}
}

// handleWalletCreated handles the wallet created event by syncing historical transactions
func (s *transactionService) handleWalletCreated(ctx context.Context, event *wallet.LifecycleEvent) error {
	// Sync historical transactions for the new wallet
	count, err := s.SyncTransactionsByAddress(ctx, event.ChainType, event.Address)
	if err != nil {
		return errors.NewTransactionSyncFailedError("sync_history", err)
	}

	s.log.Info("Synced historical transactions for new wallet",
		logger.Int64("wallet_id", event.WalletID),
		logger.String("address", event.Address),
		logger.Int("transaction_count", count))

	return nil
}
