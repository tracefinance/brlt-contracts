package transaction

import (
	"context"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Service defines the transaction service interface
type Service interface {
	// GetTransaction retrieves a transaction by its hash
	GetTransaction(ctx context.Context, hash string) (*Transaction, error)

	// GetTransactionsByWallet retrieves transactions for a specific wallet
	GetTransactionsByWallet(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error)

	// GetTransactionsByAddress retrieves transactions for a specific blockchain address
	GetTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error)

	// FilterTransactions retrieves transactions based on the provided filter criteria
	FilterTransactions(ctx context.Context, filter *Filter) (*types.Page[*Transaction], error)

	// SyncTransactions fetches and stores transactions for a wallet
	SyncTransactions(ctx context.Context, walletID int64) (int, error)

	// SyncTransactionsByAddress fetches and stores transactions for an address
	SyncTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error)

	// StartPendingTransactionPolling starts a background scheduler that periodically polls
	// for pending or mined transactions to update their status.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the polling
	StartPendingTransactionPolling(ctx context.Context)

	// StopPendingTransactionPolling stops the pending transaction polling scheduler
	StopPendingTransactionPolling()

	// StartWalletTransactionPolling starts a background scheduler that periodically polls
	// for transactions from all active wallets.
	// This provides a fallback mechanism to ensure transactions are not missed by the event-based system.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the polling
	StartWalletTransactionPolling(ctx context.Context)

	// StopWalletTransactionPolling stops the transaction polling scheduler
	StopWalletTransactionPolling()

	// SubscribeToTransactionEvents starts listening for new blocks and processing transactions.
	// This method:
	// 1. Subscribes to new block headers for all supported chains
	// 2. Processes transactions in those blocks against active wallets
	// 3. Saves transactions in the database and emits transaction events
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the subscription
	SubscribeToTransactionEvents(ctx context.Context)

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
	pollingCtx           context.Context
	pollingCancel        context.CancelFunc
	pendingPollingCtx    context.Context
	pendingPollingCancel context.CancelFunc
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

// FilterTransactions retrieves transactions based on the provided filter criteria
func (s *transactionService) FilterTransactions(ctx context.Context, filter *Filter) (*types.Page[*Transaction], error) {
	// Validate filter
	if filter == nil {
		filter = NewFilter()
	}

	// Set default values for pagination if not provided
	if filter.Limit <= 0 {
		filter.Limit = 10
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	// Use the repository to filter transactions
	return s.repository.List(ctx, filter)
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
		SortAscending: false, // Get oldest transactions first
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
	coreTx, err := explorer.GetTransactionByHash(ctx, event.TransactionHash)
	if err != nil {
		return errors.NewTransactionSyncFailedError("fetch_transaction", err)
	}

	// Convert to service transaction
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
