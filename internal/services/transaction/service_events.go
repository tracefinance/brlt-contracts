package transaction

import (
	"context"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"vault0/internal/core/tokenstore"
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

		// Start a goroutine for each chain to subscribe to ERC20 transfers
		go s.subscribeToERC20Transfers(s.eventCtx, chain)
	}

	// Subscribe to token events to dynamically update ERC20 token subscriptions
	go s.subscribeToTokenEvents(s.eventCtx)

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

			// Process the transaction to set token symbol and create Transaction model
			transaction := s.processTransaction(ctx, tx, walletID)

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

// subscribeToERC20Transfers subscribes to ERC20 token transfer events on a specific chain
func (s *transactionService) subscribeToERC20Transfers(ctx context.Context, chain types.Chain) {
	// Get blockchain client for the chain type
	client, err := s.blockchainRegistry.GetBlockchain(chain.Type)
	if err != nil {
		s.log.Error("Failed to get blockchain client for ERC20 subscription",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Get tokens from the token store for this chain
	tokenPage, err := s.tokenStore.ListTokensByChain(ctx, chain.Type, 0, 0)
	if err != nil {
		s.log.Error("Failed to get tokens for ERC20 subscription",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Filter out native tokens and collect token addresses
	var tokenAddresses []string
	for _, token := range tokenPage.Items {
		// Skip native tokens (they have empty contract addresses)
		if token.Address == "" || token.IsNative() {
			continue
		}
		tokenAddresses = append(tokenAddresses, token.Address)
	}

	// If no token addresses found, log and exit
	if len(tokenAddresses) == 0 {
		s.log.Info("No ERC20 tokens found in token store, skipping subscription",
			logger.String("chain_type", string(chain.Type)))
		return
	}

	s.log.Info("Starting ERC20 transfer event subscription",
		logger.String("chain_type", string(chain.Type)),
		logger.Int("token_count", len(tokenAddresses)))

	// Subscribe to Transfer events for specific token contracts
	logCh, errCh, err := client.SubscribeContractLogs(
		ctx,
		tokenAddresses,
		types.ERC20TransferEventSignature,
		nil, // No specific args filter, we'll check if from/to match our wallets
		0,   // Start from recent blocks
	)

	if err != nil {
		s.log.Error("Failed to subscribe to ERC20 transfers",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Process the logs
	for {
		select {
		case <-ctx.Done():
			s.log.Info("ERC20 subscription stopped",
				logger.String("chain_type", string(chain.Type)))
			return
		case err := <-errCh:
			s.log.Warn("ERC20 subscription error",
				logger.String("chain_type", string(chain.Type)),
				logger.Error(err))
		case log := <-logCh:
			s.processERC20TransferLog(ctx, chain, log)
		}
	}
}

// processERC20TransferLog processes an ERC20 Transfer event log
func (s *transactionService) processERC20TransferLog(ctx context.Context, chain types.Chain, log types.Log) {
	// Check if we have enough topics (event signature + from + to)
	if len(log.Topics) < 3 {
		s.log.Warn("Invalid ERC20 transfer log format",
			logger.String("tx_hash", log.TransactionHash))
		return
	}

	// Extract from and to addresses from the topics
	// Topics[0] is the event signature hash
	// Topics[1] is the indexed 'from' address
	// Topics[2] is the indexed 'to' address
	fromAddr := common.HexToAddress(log.Topics[1]).Hex()
	toAddr := common.HexToAddress(log.Topics[2]).Hex()

	// Normalize addresses to ensure proper comparison
	fromAddr = strings.ToLower(fromAddr)
	toAddr = strings.ToLower(toAddr)

	// Get wallets for this chain to check if transfer involves any of our wallets
	wallets, err := s.getWalletsByChainType(ctx, chain.Type)
	if err != nil {
		s.log.Error("Failed to get wallets for ERC20 transfer processing",
			logger.String("chain_type", string(chain.Type)),
			logger.Error(err))
		return
	}

	// Check if either from or to address matches any of our wallets
	var matchedWallet *wallet.Wallet
	for _, w := range wallets {
		normalizedWalletAddr := strings.ToLower(w.Address)
		if normalizedWalletAddr == fromAddr || normalizedWalletAddr == toAddr {
			matchedWallet = w
			break
		}
	}

	// If this transfer doesn't involve any of our wallets, ignore it
	if matchedWallet == nil {
		return
	}

	// Create a new transaction directly from the log data
	tokenAddress := log.Address

	// Parse the transfer amount from log data
	var value *big.Int
	if len(log.Data) > 0 {
		value = new(big.Int).SetBytes(log.Data)
	} else {
		value = big.NewInt(0)
	}

	// Create a transaction record
	tx := &types.Transaction{
		Chain:        chain.Type,
		Hash:         log.TransactionHash,
		From:         fromAddr,
		To:           toAddr,
		Value:        value,
		Type:         types.TransactionTypeERC20,
		TokenAddress: tokenAddress,
		Status:       types.TransactionStatusSuccess, // ERC20 transfer logs occur only for successful transfers
		BlockNumber:  log.BlockNumber,
	}

	// Get token details from token store
	token, err := s.tokenStore.GetToken(ctx, tokenAddress, chain.Type)
	if err == nil && token != nil {
		tx.TokenSymbol = token.Symbol
	} else {
		s.log.Warn("Token not found in token store for ERC20 transfer",
			logger.String("token_address", tokenAddress),
			logger.String("chain", string(chain.Type)))
	}

	// Process the transaction to set additional details and create Transaction model
	transaction := s.processTransaction(ctx, tx, matchedWallet.ID)

	// Check if transaction already exists
	exists, err := s.repository.Exists(ctx, transaction.Hash)
	if err != nil {
		s.log.Error("Failed to check if transaction exists",
			logger.String("tx_hash", tx.Hash),
			logger.Int64("wallet_id", matchedWallet.ID),
			logger.Error(err))
		return
	}

	// If transaction exists, get it from the database
	if exists {
		transaction, err = s.repository.GetByTxHash(ctx, transaction.Hash)
		if err != nil {
			s.log.Error("Failed to get existing transaction",
				logger.String("tx_hash", tx.Hash),
				logger.Int64("wallet_id", matchedWallet.ID),
				logger.Error(err))
			return
		}
	} else {
		// Otherwise, save the new transaction
		err = s.repository.Create(ctx, transaction)
		if err != nil {
			s.log.Error("Failed to save transaction",
				logger.String("tx_hash", tx.Hash),
				logger.Int64("wallet_id", matchedWallet.ID),
				logger.Error(err))
			return
		}
	}

	// Emit transaction event
	s.emitTransactionEvent(&TransactionEvent{
		WalletID:    matchedWallet.ID,
		Transaction: transaction,
		BlockNumber: log.BlockNumber.Int64(),
		EventType:   EventTypeTransactionDetected,
	})

	// Update last block number for the wallet if this block is newer
	if log.BlockNumber != nil && log.BlockNumber.Int64() > matchedWallet.LastBlockNumber {
		if err := s.walletService.UpdateLastBlockNumber(ctx, matchedWallet.ChainType, matchedWallet.Address, log.BlockNumber.Int64()); err != nil {
			s.log.Error("Failed to update last block number after ERC20 transfer",
				logger.Int64("wallet_id", matchedWallet.ID),
				logger.String("address", matchedWallet.Address),
				logger.Error(err))
		}
	}
}

// subscribeToTokenEvents listens for token event notifications and updates ERC20 subscriptions
func (s *transactionService) subscribeToTokenEvents(ctx context.Context) {
	tokenEvents := s.tokenStore.TokenEvents()
	s.log.Info("Started token events subscription")

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Token events subscription stopped")
			return
		case event, ok := <-tokenEvents:
			if !ok {
				// Channel was closed
				s.log.Info("Token events channel closed")
				return
			}

			// Only react to token added events
			if event.EventType == tokenstore.TokenEventAdded && event.Token != nil {
				// Skip native tokens
				if event.Token.IsNative() {
					continue
				}

				s.log.Info("New token added, updating ERC20 subscription",
					logger.String("symbol", event.Token.Symbol),
					logger.String("address", event.Token.Address),
					logger.String("chain", string(event.Token.ChainType)))

				// For simplicity, restart the entire ERC20 subscription for this chain
				// A more optimized approach would be to add this token to existing subscriptions
				if chain, exists := s.chains.Chains[event.Token.ChainType]; exists {
					// Create a new context for the restarted subscription
					tokenCtx, cancel := context.WithCancel(ctx)

					// Start new subscription
					go func(chain types.Chain) {
						s.subscribeToERC20Transfers(tokenCtx, chain)
					}(chain)

					// Cancel any previous subscription for this chain after a short delay
					// This ensures we have the new subscription running before canceling the old one
					go func() {
						time.Sleep(5 * time.Second)
						cancel()
					}()
				}
			}
		}
	}
}
