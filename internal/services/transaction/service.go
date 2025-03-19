package transaction

import (
	"context"
	"math/big"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
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

// Service defines the transaction service interface
type Service interface {
	// GetTransaction retrieves a transaction by its hash
	GetTransaction(ctx context.Context, hash string) (*Transaction, error)

	// GetTransactionsByWallet retrieves transactions for a specific wallet
	GetTransactionsByWallet(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error)

	// GetTransactionsByAddress retrieves transactions for a specific blockchain address
	GetTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error)

	// SyncTransactions fetches and stores transactions for a wallet
	SyncTransactions(ctx context.Context, walletID int64) (int, error)

	// SyncTransactionsByAddress fetches and stores transactions for an address
	SyncTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error)

	// SubscribeTransactionEvents starts listening for new blocks and processing transactions.
	// This method:
	// 1. Subscribes to new block headers for all supported chains
	// 2. Processes transactions in those blocks against active wallets
	// 3. Saves transactions in the database and emits transaction events
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the subscription
	SubscribeTransactionEvents(ctx context.Context)

	// UnsubscribeFromTransactionEvents stops listening for blockchain events.
	// This should be called when shutting down the service.
	UnsubscribeFromTransactionEvents()

	// TransactionEvents returns a channel that emits transaction events.
	// These events include new transactions detected for monitored wallets.
	// The channel is closed when UnsubscribeFromTransactionEvents is called.
	TransactionEvents() <-chan *TransactionEvent
}

// transactionService implements the Service interface
type transactionService struct {
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	walletService        wallet.Service
	blockExplorerFactory blockexplorer.Factory
	blockchainRegistry   blockchain.Registry
	chains               *types.Chains
	syncMutex            sync.Mutex
	eventCtx             context.Context
	eventCancel          context.CancelFunc
	transactionEvents    chan *TransactionEvent
}

// NewService creates a new transaction service
func NewService(
	config *config.Config,
	log logger.Logger,
	repository Repository,
	walletService wallet.Service,
	blockExplorerFactory blockexplorer.Factory,
	blockchainRegistry blockchain.Registry,
	chains *types.Chains,
) Service {
	const channelBufferSize = 100
	return &transactionService{
		config:               config,
		log:                  log,
		repository:           repository,
		walletService:        walletService,
		blockExplorerFactory: blockExplorerFactory,
		blockchainRegistry:   blockchainRegistry,
		chains:               chains,
		transactionEvents:    make(chan *TransactionEvent, channelBufferSize),
	}
}

// GetTransaction retrieves a transaction by its hash
func (s *transactionService) GetTransaction(ctx context.Context, hash string) (*Transaction, error) {
	if hash == "" {
		return nil, errors.NewInvalidInputError("Hash is required", "hash", "")
	}

	// Get transaction directly from repository
	return s.repository.GetByTxHash(ctx, hash)
}

// GetTransactionsByWallet retrieves transactions for a specific wallet
func (s *transactionService) GetTransactionsByWallet(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error) {
	// Validate input
	if walletID <= 0 {
		return nil, errors.NewInvalidInputError("Wallet ID is required", "wallet_id", "")
	}

	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get transactions from database
	return s.repository.ListByWalletID(ctx, walletID, limit, offset)
}

// GetTransactionsByAddress retrieves transactions for a specific blockchain address
func (s *transactionService) GetTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error) {
	// Validate input
	if chainType == "" {
		return nil, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return nil, errors.NewInvalidInputError("Address is required", "address", "")
	}

	// Set default values
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get transactions from database
	return s.repository.ListByWalletAddress(ctx, chainType, address, limit, offset)
}

// SyncTransactions fetches and stores transactions for a wallet
func (s *transactionService) SyncTransactions(ctx context.Context, walletID int64) (int, error) {
	// Prevent concurrent syncs
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	// Validate input
	if walletID <= 0 {
		return 0, errors.NewInvalidInputError("Wallet ID is required", "wallet_id", "")
	}

	// Get wallet from wallet service by ID
	wallet, err := s.walletService.GetByID(ctx, walletID)
	if err != nil {
		return 0, err
	}

	// Sync transactions for the wallet's address
	return s.SyncTransactionsByAddress(ctx, wallet.ChainType, wallet.Address)
}

// SyncTransactionsByAddress fetches and stores transactions for an address
func (s *transactionService) SyncTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error) {
	// Validate input
	if chainType == "" {
		return 0, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return 0, errors.NewInvalidInputError("Address is required", "address", "")
	}

	// Get explorer for the chain
	explorer, err := s.blockExplorerFactory.GetExplorer(chainType)
	if err != nil {
		return 0, err
	}

	// Get wallet ID and last block number if exists
	wallet, err := s.walletService.GetByAddress(ctx, chainType, address)
	if err != nil {
		return 0, err
	}

	// Prepare options for fetching transactions
	options := blockexplorer.TransactionHistoryOptions{
		StartBlock: wallet.LastBlockNumber,
		EndBlock:   0, // Latest block
		Page:       1,
		PageSize:   100,
		TransactionTypes: []blockexplorer.TransactionType{
			blockexplorer.TxTypeNormal,
			blockexplorer.TxTypeInternal,
			blockexplorer.TxTypeERC20,
			blockexplorer.TxTypeERC721,
		},
		SortAscending: false, // Get newest transactions first
	}

	// Fetch transactions from explorer
	txs, err := explorer.GetTransactionHistory(ctx, address, options)
	if err != nil {
		return 0, errors.NewTransactionSyncFailedError("fetch_history", err)
	}

	// Save transactions to database
	count := 0
	var maxBlockNumber int64
	for _, coreTx := range txs.Items {
		// Update max block number if this transaction's block number is higher
		if coreTx.BlockNumber != nil {
			blockNum := coreTx.BlockNumber.Int64()
			if blockNum > maxBlockNumber {
				maxBlockNumber = blockNum
			}
		}

		// Check if transaction already exists
		exists, err := s.repository.Exists(ctx, coreTx.Hash)
		if err != nil {
			s.log.Warn("Failed to check transaction existence",
				logger.String("hash", coreTx.Hash),
				logger.Error(err))
			continue
		}

		if exists {
			continue
		}

		// Convert to service transaction
		tx := FromCoreTransaction(coreTx, wallet.ID)

		// Save to database
		err = s.repository.Create(ctx, tx)
		if err != nil {
			s.log.Warn("Failed to save transaction",
				logger.String("hash", coreTx.Hash),
				logger.Error(err))
			continue
		}

		count++
	}

	// Update wallet's last block number if we found new transactions with a higher block number
	if wallet != nil && maxBlockNumber > wallet.LastBlockNumber {
		if err := s.walletService.UpdateLastBlockNumber(ctx, wallet.ChainType, wallet.Address, maxBlockNumber); err != nil {
			s.log.Error("Failed to update wallet's last block number",
				logger.Int64("wallet_id", wallet.ID),
				logger.Int64("block_number", maxBlockNumber),
				logger.Error(err))
		}
	}

	return count, nil
}

// OnWalletEvent processes a blockchain event for a wallet and updates the transactions table
func (s *transactionService) OnWalletEvent(ctx context.Context, walletID int64, event *types.Log) error {
	// Validate input
	if walletID <= 0 {
		return errors.NewInvalidInputError("Wallet ID is required", "wallet_id", "")
	}
	if event == nil {
		return errors.NewInvalidInputError("Event is required", "event", nil)
	}

	// Get wallet information
	wallet, err := s.walletService.GetByID(ctx, walletID)
	if err != nil {
		return err
	}

	// Check if transaction already exists
	exists, err := s.repository.Exists(ctx, event.TransactionHash)
	if err != nil {
		return err
	}

	if exists {
		// Transaction already processed
		return nil
	}

	// Get explorer for the chain
	explorer, err := s.blockExplorerFactory.GetExplorer(wallet.ChainType)
	if err != nil {
		return err
	}

	// Fetch full transaction details
	txs, err := explorer.GetTransactionsByHash(ctx, []string{event.TransactionHash})
	if err != nil {
		return errors.NewTransactionSyncFailedError("fetch_transaction", err)
	}

	if len(txs) == 0 {
		return errors.NewTransactionNotFoundError(event.TransactionHash)
	}

	// Convert to service transaction
	coreTx := txs[0]
	tx := &Transaction{
		ChainType:    coreTx.Chain,
		Hash:         coreTx.Hash,
		FromAddress:  coreTx.From,
		ToAddress:    coreTx.To,
		Value:        coreTx.Value,
		Data:         coreTx.Data,
		Nonce:        coreTx.Nonce,
		GasPrice:     coreTx.GasPrice,
		GasLimit:     coreTx.GasLimit,
		Type:         string(coreTx.Type),
		TokenAddress: coreTx.TokenAddress,
		Status:       string(coreTx.Status),
		Timestamp:    coreTx.Timestamp,
		WalletID:     walletID,
	}

	// Save to database
	if err := s.repository.Create(ctx, tx); err != nil {
		return err
	}

	// Update wallet's last block number if the transaction has a higher block number
	if event.BlockNumber != nil && event.BlockNumber.Int64() > wallet.LastBlockNumber {
		blockNumber := event.BlockNumber.Int64()
		if err := s.walletService.UpdateLastBlockNumber(ctx, wallet.ChainType, wallet.Address, blockNumber); err != nil {
			s.log.Error("Failed to update wallet's last block number",
				logger.Int64("wallet_id", wallet.ID),
				logger.Int64("block_number", blockNumber),
				logger.Error(err))
		} else {
			s.log.Info("Updated wallet's last block number",
				logger.Int64("wallet_id", wallet.ID),
				logger.Int64("block_number", blockNumber))
		}
	}

	s.log.Info("New transaction processed",
		logger.Int64("wallet_id", walletID),
		logger.String("tx_hash", event.TransactionHash),
		logger.String("chain_type", string(wallet.ChainType)))

	return nil
}

// SubscribeTransactionEvents starts listening for new blocks and processing transactions
func (s *transactionService) SubscribeTransactionEvents(ctx context.Context) {
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
			s.log.Error("Block subscription error",
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
		var isOutgoing bool

		// Check if the transaction involves any of our monitored wallets
		if tx.From != "" {
			if w, exists := addressToWallet[tx.From]; exists {
				walletID = w.ID
				isOutgoing = true
			}
		}

		if tx.To != "" {
			if w, exists := addressToWallet[tx.To]; exists {
				walletID = w.ID
				isOutgoing = false
			}
		}

		// If this transaction involves one of our wallets, save it
		if walletID > 0 {
			// Save transaction to database
			transaction, err := s.saveTransaction(ctx, tx, walletID, isOutgoing)
			if err != nil {
				s.log.Error("Failed to save transaction",
					logger.String("tx_hash", tx.Hash),
					logger.Int64("wallet_id", walletID),
					logger.Error(err))
				continue
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
		s.log.Warn("Transaction events channel is full, dropping event",
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

// saveTransaction converts a blockchain transaction to a database transaction and saves it
func (s *transactionService) saveTransaction(ctx context.Context, tx *types.Transaction, walletID int64, isOutgoing bool) (*Transaction, error) {
	// Use the existing helper function to convert core transaction to service transaction
	transaction := FromCoreTransaction(tx, walletID)

	// Check if transaction already exists
	exists, err := s.repository.Exists(ctx, transaction.Hash)
	if err != nil {
		return nil, errors.NewOperationFailedError("check transaction exists", err)
	}

	// If transaction already exists, get it from the database
	if exists {
		return s.repository.GetByTxHash(ctx, transaction.Hash)
	}

	// Otherwise, save the new transaction
	err = s.repository.Create(ctx, transaction)
	if err != nil {
		return nil, errors.NewOperationFailedError("save transaction", err)
	}

	return transaction, nil
}

// calculateTransactionFee calculates the transaction fee from gas used and gas price
func calculateTransactionFee(gasUsed uint64, gasPrice *big.Int) string {
	if gasPrice == nil {
		return "0"
	}

	// Calculate fee = gasUsed * gasPrice
	fee := new(big.Int).Mul(big.NewInt(int64(gasUsed)), gasPrice)
	return fee.String()
}
