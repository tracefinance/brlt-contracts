package transaction

import (
	"context"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Service defines the transaction service interface
type Service interface {
	// GetTransaction retrieves a transaction by its chain type and hash
	GetTransaction(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error)

	// GetTransactionsByWallet retrieves transactions for a specific wallet
	GetTransactionsByWallet(ctx context.Context, walletID string, limit, offset int) (*types.Page[*Transaction], error)

	// GetTransactionsByAddress retrieves transactions for a specific blockchain address
	GetTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error)

	// SyncTransactions fetches and stores transactions for a wallet
	SyncTransactions(ctx context.Context, walletID string) (int, error)

	// SyncTransactionsByAddress fetches and stores transactions for an address
	SyncTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error)

	// CountTransactions counts transactions for a specific wallet
	CountTransactions(ctx context.Context, walletID string) (int, error)

	// SubscribeToWalletEvents starts listening for wallet events and processing transactions.
	// This should be called after the service is initialized.
	SubscribeToWalletEvents(ctx context.Context)

	// UnsubscribeFromWalletEvents stops listening for wallet events.
	// This should be called when shutting down the service.
	UnsubscribeFromWalletEvents()
}

// transactionService implements the Service interface
type transactionService struct {
	config               *config.Config
	logger               logger.Logger
	repository           Repository
	walletService        wallet.Service
	blockExplorerFactory blockexplorer.Factory
	chains               *types.Chains
	syncMutex            sync.Mutex
	eventCtx             context.Context
	eventCancel          context.CancelFunc
}

// NewService creates a new transaction service
func NewService(
	config *config.Config,
	logger logger.Logger,
	repository Repository,
	walletService wallet.Service,
	blockExplorerFactory blockexplorer.Factory,
	chains *types.Chains,
) Service {
	return &transactionService{
		config:               config,
		logger:               logger,
		repository:           repository,
		walletService:        walletService,
		blockExplorerFactory: blockExplorerFactory,
		chains:               chains,
	}
}

// GetTransaction retrieves a transaction by its chain type and hash
func (s *transactionService) GetTransaction(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
	// Validate input
	if chainType == "" {
		return nil, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if hash == "" {
		return nil, errors.NewInvalidInputError("Hash is required", "hash", "")
	}

	// Try to get from database first
	tx, err := s.repository.Get(ctx, chainType, hash)
	if err == nil {
		return tx, nil
	}

	// If not found in database, fetch from blockchain explorer
	if _, ok := err.(*errors.Vault0Error); ok && err.(*errors.Vault0Error).Code == errors.ErrCodeTransactionNotFound {
		// Get explorer for the chain
		explorer, err := s.blockExplorerFactory.GetExplorer(chainType)
		if err != nil {
			return nil, err
		}
		defer explorer.Close()

		// Fetch transaction from explorer
		txs, err := explorer.GetTransactionsByHash(ctx, []string{hash})
		if err != nil {
			return nil, err
		}

		if len(txs) == 0 {
			return nil, errors.NewTransactionNotFoundError(hash)
		}

		// Convert to service transaction
		coreTx := txs[0]
		tx = &Transaction{
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
			Status:       coreTx.Status,
			Timestamp:    coreTx.Timestamp,
		}

		// Save to database (best effort)
		_ = s.repository.Create(ctx, tx)

		return tx, nil
	}

	return nil, err
}

// GetTransactionsByWallet retrieves transactions for a specific wallet
func (s *transactionService) GetTransactionsByWallet(ctx context.Context, walletID string, limit, offset int) (*types.Page[*Transaction], error) {
	// Validate input
	if walletID == "" {
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
	return s.repository.GetByWallet(ctx, walletID, limit, offset)
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
	return s.repository.GetByAddress(ctx, chainType, address, limit, offset)
}

// SyncTransactions fetches and stores transactions for a wallet
func (s *transactionService) SyncTransactions(ctx context.Context, walletID string) (int, error) {
	// Prevent concurrent syncs
	s.syncMutex.Lock()
	defer s.syncMutex.Unlock()

	// Validate input
	if walletID == "" {
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
	defer explorer.Close()

	// Get wallet ID and last block number if exists
	var walletID string
	var startBlock int64
	wallet, err := s.walletService.Get(ctx, chainType, address)
	if err == nil {
		walletID = wallet.ID
		startBlock = wallet.LastBlockNumber
	}

	// Prepare options for fetching transactions
	options := blockexplorer.TransactionHistoryOptions{
		StartBlock: startBlock,
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
		exists, err := s.repository.Exists(ctx, chainType, coreTx.Hash)
		if err != nil {
			s.logger.Warn("Failed to check transaction existence",
				logger.String("hash", coreTx.Hash),
				logger.Error(err))
			continue
		}

		if exists {
			continue
		}

		// Convert to service transaction
		tx := FromCoreTransaction(coreTx, walletID)

		// Save to database
		err = s.repository.Create(ctx, tx)
		if err != nil {
			s.logger.Warn("Failed to save transaction",
				logger.String("hash", coreTx.Hash),
				logger.Error(err))
			continue
		}

		count++
	}

	// Update wallet's last block number if we found new transactions with a higher block number
	if wallet != nil && maxBlockNumber > wallet.LastBlockNumber {
		if err := s.walletService.UpdateLastBlockNumber(ctx, wallet.ChainType, wallet.Address, maxBlockNumber); err != nil {
			s.logger.Error("Failed to update wallet's last block number",
				logger.String("wallet_id", wallet.ID),
				logger.Int64("block_number", maxBlockNumber),
				logger.Error(err))
		}
	}

	return count, nil
}

// CountTransactions counts transactions for a specific wallet
func (s *transactionService) CountTransactions(ctx context.Context, walletID string) (int, error) {
	// Validate input
	if walletID == "" {
		return 0, errors.NewInvalidInputError("Wallet ID is required", "wallet_id", "")
	}

	return s.repository.Count(ctx, walletID)
}

// OnWalletEvent processes a blockchain event for a wallet and updates the transactions table
func (s *transactionService) OnWalletEvent(ctx context.Context, walletID string, event *types.Log) error {
	// Validate input
	if walletID == "" {
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
	exists, err := s.repository.Exists(ctx, wallet.ChainType, event.TransactionHash)
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
	defer explorer.Close()

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
		Status:       coreTx.Status,
		Timestamp:    coreTx.Timestamp,
		WalletID:     walletID,
	}

	// Save to database
	if err := s.repository.Create(ctx, tx); err != nil {
		return err
	}

	s.logger.Info("New transaction processed",
		logger.String("wallet_id", walletID),
		logger.String("tx_hash", event.TransactionHash),
		logger.String("chain_type", string(wallet.ChainType)))

	return nil
}

// SubscribeToWalletEvents starts listening for wallet events
func (s *transactionService) SubscribeToWalletEvents(ctx context.Context) {
	s.eventCtx, s.eventCancel = context.WithCancel(ctx)

	// Start goroutine for blockchain events
	go func() {
		eventCh := s.walletService.BlockchainEvents()
		for {
			select {
			case <-s.eventCtx.Done():
				return
			case event, ok := <-eventCh:
				if !ok {
					// Channel was closed
					return
				}
				// Process blockchain event
				if err := s.OnWalletEvent(s.eventCtx, event.WalletID, event.Log); err != nil {
					s.logger.Error("Failed to process blockchain event",
						logger.String("wallet_id", event.WalletID),
						logger.String("tx_hash", event.Log.TransactionHash),
						logger.Error(err))
				}
			}
		}
	}()

	// Start goroutine for lifecycle events
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
						s.logger.Error("Failed to handle wallet created event",
							logger.String("wallet_id", event.WalletID),
							logger.Error(err))
					}

				case wallet.EventTypeWalletDeleted:
					// When a wallet is deleted, we don't need to do anything
					// The transactions will remain in the database for historical purposes
					s.logger.Info("Wallet deleted",
						logger.String("wallet_id", event.WalletID))
				}
			}
		}
	}()
}

// handleWalletCreated handles the wallet created event by syncing historical transactions
func (s *transactionService) handleWalletCreated(ctx context.Context, event *wallet.LifecycleEvent) error {
	// Sync historical transactions for the new wallet
	count, err := s.SyncTransactionsByAddress(ctx, event.ChainType, event.Address)
	if err != nil {
		return errors.NewTransactionSyncFailedError("sync_history", err)
	}

	s.logger.Info("Synced historical transactions for new wallet",
		logger.String("wallet_id", event.WalletID),
		logger.String("address", event.Address),
		logger.Int("transaction_count", count))

	return nil
}

// UnsubscribeFromWalletEvents stops listening for wallet events
func (s *transactionService) UnsubscribeFromWalletEvents() {
	if s.eventCancel != nil {
		s.eventCancel()
	}
}
