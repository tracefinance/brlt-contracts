package transaction

import (
	"context"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/tokenstore"
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
	tokenStore           tokenstore.TokenStore
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
	tokenStore tokenstore.TokenStore,
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
		tokenStore:           tokenStore,
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

		// Process the transaction to set token symbol
		tx := s.processTransaction(ctx, coreTx, wallet.ID)

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

// processTransaction resolves the token symbol and creates a Transaction model
func (s *transactionService) processTransaction(ctx context.Context, coreTx *types.Transaction, walletID int64) *Transaction {
	// Resolve token symbol if not already set
	if coreTx.TokenSymbol == "" {
		// For native transactions, use the chain's native token symbol
		if coreTx.Type == types.TransactionTypeNative {
			// Get the chain configuration
			chain, exists := s.chains.Chains[coreTx.Chain]
			if exists {
				coreTx.TokenSymbol = chain.Symbol
			} else {
				coreTx.TokenSymbol = "UNKNOWN"
			}
		}
		// For ERC20 transactions, look up the token symbol from the token store
		if coreTx.Type == types.TransactionTypeERC20 && coreTx.TokenAddress != "" {
			token, err := s.tokenStore.GetToken(ctx, coreTx.TokenAddress, coreTx.Chain)
			if err == nil && token != nil {
				coreTx.TokenSymbol = token.Symbol
			} else {
				// If token not found in store, log a warning
				s.log.Warn("Token not found in token store",
					logger.String("chain", string(coreTx.Chain)),
					logger.String("address", coreTx.TokenAddress))
			}
		}
	}

	// Convert to service transaction model
	return FromCoreTransaction(coreTx, walletID)
}

// createAddressToWalletMap creates a map of normalized wallet addresses to wallet objects
// for efficient lookup. All addresses are converted to lowercase for consistent comparison.
func (s *transactionService) createAddressToWalletMap(wallets []*wallet.Wallet) map[string]*wallet.Wallet {
	addressToWallet := make(map[string]*wallet.Wallet)
	for _, w := range wallets {
		// Normalize address by converting to lowercase
		key := types.NormalizeAddress(w.Address)
		addressToWallet[key] = w
	}
	return addressToWallet
}

// findRelevantWallet determines if a transaction involves one of our monitored wallets
// and returns the wallet if found. Addresses are normalized for consistent comparison.
func (s *transactionService) findRelevantWallet(fromAddr string, toAddr string, addressToWallet map[string]*wallet.Wallet) *wallet.Wallet {
	// Check if the transaction involves any of our monitored wallets
	if fromAddr != "" {
		key := types.NormalizeAddress(fromAddr)
		if w, exists := addressToWallet[key]; exists {
			return w
		}
	}

	if toAddr != "" {
		key := types.NormalizeAddress(toAddr)
		if w, exists := addressToWallet[key]; exists {
			return w
		}
	}

	return nil
}

// processTransactionRecord handles transaction database operations (check existence, retrieve or create)
func (s *transactionService) processTransactionRecord(ctx context.Context, tx *types.Transaction, walletID int64) (*Transaction, error) {
	// Process the transaction to set token symbol and create Transaction model
	transaction := s.processTransaction(ctx, tx, walletID)

	// Check if transaction already exists
	exists, err := s.repository.Exists(ctx, transaction.Hash)
	if err != nil {
		s.log.Error("Failed to check if transaction exists",
			logger.String("tx_hash", tx.Hash),
			logger.Int64("wallet_id", walletID),
			logger.Error(err))
		return nil, err
	}

	// If transaction exists, get it from the database
	if exists {
		transaction, err = s.repository.GetByTxHash(ctx, transaction.Hash)
		if err != nil {
			s.log.Error("Failed to get existing transaction",
				logger.String("tx_hash", tx.Hash),
				logger.Int64("wallet_id", walletID),
				logger.Error(err))
			return nil, err
		}
	} else {
		// Otherwise, save the new transaction
		err = s.repository.Create(ctx, transaction)
		if err != nil {
			s.log.Error("Failed to save transaction",
				logger.String("tx_hash", tx.Hash),
				logger.Int64("wallet_id", walletID),
				logger.Error(err))
			return nil, err
		}
	}

	return transaction, nil
}

// updateBalanceForTransaction handles wallet balance updates based on transaction type.
// This method follows a "best effort" approach, handling all errors internally.
// Errors are logged for diagnostic purposes but not returned to callers,
// allowing transaction processing to continue even if balance updates fail.
// Balance updates are considered non-critical side effects of transaction processing.
func (s *transactionService) updateBalanceForTransaction(
	ctx context.Context,
	transaction *Transaction,
	wallet *wallet.Wallet,
) {
	chainType := wallet.ChainType

	// Only update balances for successful transactions
	if transaction.Status != string(types.TransactionStatusSuccess) {
		return
	}

	// Handle based on transaction type
	if transaction.Type == string(types.TransactionTypeNative) {
		// Native token transaction
		s.updateNativeBalance(ctx, wallet, chainType)
	} else if transaction.Type == string(types.TransactionTypeERC20) && transaction.TokenAddress != "" {
		// ERC20 token transaction
		s.updateTokenBalance(ctx, wallet, chainType, transaction.TokenAddress)
	}
}

// updateNativeBalance updates the native token balance for a wallet.
// This method handles all errors internally and logs them without returning errors to callers.
// If any operation fails (blockchain client retrieval, balance fetching, or database updates),
// the error is logged but transaction processing continues.
// This design prioritizes transaction tracking over balance accuracy,
// as balances can be corrected in subsequent update attempts.
func (s *transactionService) updateNativeBalance(
	ctx context.Context,
	wallet *wallet.Wallet,
	chainType types.ChainType,
) {
	// Get blockchain client
	client, err := s.blockchainRegistry.GetBlockchain(chainType)
	if err != nil {
		s.log.Error("Failed to get blockchain client for native balance update",
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
		return
	}

	// Get the latest balance from blockchain
	balance, err := client.GetBalance(ctx, wallet.Address)
	if err != nil {
		s.log.Error("Failed to get wallet balance",
			logger.String("address", wallet.Address),
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
		return
	}

	// Update wallet balance
	if err := s.walletService.UpdateWalletBalance(ctx, chainType, wallet.Address, balance); err != nil {
		s.log.Error("Failed to update wallet balance",
			logger.String("address", wallet.Address),
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
	} else {
		s.log.Info("Updated wallet balance",
			logger.String("address", wallet.Address),
			logger.String("chain_type", string(chainType)),
			logger.String("balance", balance.String()))
	}
}

// updateTokenBalance updates a token balance for a wallet.
// Like other balance update methods, this follows a non-blocking error handling approach,
// where errors are logged but not returned to callers.
// This allows the main transaction processing flow to continue uninterrupted
// even if token balance updates cannot be completed.
// The logs contain detailed error context for diagnostic purposes.
func (s *transactionService) updateTokenBalance(
	ctx context.Context,
	wallet *wallet.Wallet,
	chainType types.ChainType,
	tokenAddress string,
) {
	// Get blockchain client
	client, err := s.blockchainRegistry.GetBlockchain(chainType)
	if err != nil {
		s.log.Error("Failed to get blockchain client for token balance update",
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
		return
	}

	// Get token info
	token, err := s.tokenStore.GetToken(ctx, tokenAddress, chainType)
	if err != nil {
		s.log.Error("Failed to get token",
			logger.String("token_address", tokenAddress),
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
		return
	}

	// Get the latest token balance from blockchain
	tokenBalance, err := client.GetTokenBalance(ctx, tokenAddress, wallet.Address)
	if err != nil {
		s.log.Error("Failed to get token balance",
			logger.String("address", wallet.Address),
			logger.String("token_address", tokenAddress),
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
		return
	}

	// Update token balance
	if err := s.walletService.UpdateTokenBalance(ctx, chainType, wallet.Address, tokenAddress, tokenBalance); err != nil {
		s.log.Error("Failed to update token balance",
			logger.String("address", wallet.Address),
			logger.String("token_address", tokenAddress),
			logger.String("chain_type", string(chainType)),
			logger.Error(err))
	} else {
		s.log.Info("Updated token balance",
			logger.String("address", wallet.Address),
			logger.String("token_address", tokenAddress),
			logger.String("token_symbol", token.Symbol),
			logger.String("chain_type", string(chainType)),
			logger.String("balance", tokenBalance.String()))
	}
}

// handleTransactionCompletion emits events and updates the last block number
func (s *transactionService) handleTransactionCompletion(
	ctx context.Context,
	wallet *wallet.Wallet,
	transaction *Transaction,
	blockNumber int64,
) {
	// Emit transaction event
	s.emitTransactionEvent(&TransactionEvent{
		WalletID:    wallet.ID,
		Transaction: transaction,
		BlockNumber: blockNumber,
		EventType:   EventTypeTransactionDetected,
	})

	// Update last block number for the wallet if this block is newer
	if blockNumber > wallet.LastBlockNumber {
		if err := s.walletService.UpdateLastBlockNumber(ctx, wallet.ChainType, wallet.Address, blockNumber); err != nil {
			s.log.Error("Failed to update last block number",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.Error(err))
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
