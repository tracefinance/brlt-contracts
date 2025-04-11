package transaction

import (
	"context"
	"strings"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
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

	// StartPendingTransactionPolling starts a background scheduler that periodically polls
	// for pending or mined transactions to update their status.
	//
	// Parameters:
	//   - ctx: Context for the operation, used to cancel the polling
	StartPendingTransactionPolling(ctx context.Context)

	// StopPendingTransactionPolling stops the pending transaction polling scheduler
	StopPendingTransactionPolling()

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

	// TransactionEvents returns a channel that emits raw blockchain transactions.
	// These events include all transactions detected on monitored chains.
	// The channel is closed when UnsubscribeFromTransactionEvents is called.
	TransactionEvents() <-chan *types.Transaction

	// MonitorAddress adds an address to the list of addresses whose transactions should be emitted.
	MonitorAddress(ctx context.Context, addr *types.Address) error

	// UnmonitoredAddress removes an address from the monitoring list.
	UnmonitoredAddress(ctx context.Context, addr *types.Address) error
}

// transactionService implements the Service interface
type transactionService struct {
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	tokenStore           tokenstore.TokenStore
	blockExplorerFactory blockexplorer.Factory
	blockchainRegistry   blockchain.Registry
	chains               *types.Chains
	eventCtx             context.Context
	eventCancel          context.CancelFunc
	pendingPollingCtx    context.Context
	pendingPollingCancel context.CancelFunc
	transactionEvents    chan *types.Transaction
	// In-memory store for addresses to monitor
	monitoredAddresses map[types.ChainType]map[string]struct{}
	addressMutex       sync.RWMutex
}

// NewService creates a new transaction service
func NewService(
	config *config.Config,
	log logger.Logger,
	repository Repository,
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
		tokenStore:           tokenStore,
		blockExplorerFactory: blockExplorerFactory,
		blockchainRegistry:   blockchainRegistry,
		chains:               chains,
		transactionEvents:    make(chan *types.Transaction, channelBufferSize),
		// Initialize the monitored addresses map
		monitoredAddresses: make(map[types.ChainType]map[string]struct{}),
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

// processTransaction resolves the token symbol and creates a Transaction model
func (s *transactionService) processTransaction(ctx context.Context, coreTx *types.Transaction) *Transaction {
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
			token, err := s.tokenStore.GetToken(ctx, coreTx.TokenAddress)
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

	// Convert to service transaction model, WalletID is now 0 as it's not linked here
	return FromCoreTransaction(coreTx, 0)
}

// TransactionEvents returns a channel that emits raw blockchain transactions.
// These events include all transactions detected on monitored chains.
// The channel is closed when UnsubscribeFromTransactionEvents is called.
func (s *transactionService) TransactionEvents() <-chan *types.Transaction {
	return s.transactionEvents
}

// MonitorAddress adds an address to the in-memory monitoring list
func (s *transactionService) MonitorAddress(ctx context.Context, addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup

	s.addressMutex.Lock()
	defer s.addressMutex.Unlock()

	if _, ok := s.monitoredAddresses[addr.ChainType]; !ok {
		s.monitoredAddresses[addr.ChainType] = make(map[string]struct{})
	}

	if _, exists := s.monitoredAddresses[addr.ChainType][normalizedAddr]; !exists {
		s.monitoredAddresses[addr.ChainType][normalizedAddr] = struct{}{}
		s.log.Info("Added address to monitoring list",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(addr.ChainType)))
	} else {
		s.log.Debug("Address already monitored",
			logger.String("address", addr.Address),
			logger.String("chain_type", string(addr.ChainType)))
	}

	return nil
}

// UnmonitoredAddress removes an address from the in-memory monitoring list
func (s *transactionService) UnmonitoredAddress(ctx context.Context, addr *types.Address) error {
	if addr == nil {
		return errors.NewInvalidInputError("Address cannot be nil", "address", nil)
	}
	// We don't strictly need validation here, but it's good practice
	if err := addr.Validate(); err != nil {
		return err
	}

	normalizedAddr := strings.ToLower(addr.Address) // Normalize for consistent lookup

	s.addressMutex.Lock()
	defer s.addressMutex.Unlock()

	if chainMap, ok := s.monitoredAddresses[addr.ChainType]; ok {
		if _, exists := chainMap[normalizedAddr]; exists {
			delete(chainMap, normalizedAddr)
			s.log.Info("Removed address from monitoring list",
				logger.String("address", addr.Address),
				logger.String("chain_type", string(addr.ChainType)))
			// Clean up the chain map if it becomes empty
			if len(chainMap) == 0 {
				delete(s.monitoredAddresses, addr.ChainType)
			}
		} else {
			s.log.Debug("Address not found in monitoring list for removal",
				logger.String("address", addr.Address),
				logger.String("chain_type", string(addr.ChainType)))
		}
	} else {
		s.log.Debug("Chain type not found in monitoring list for removal",
			logger.String("chain_type", string(addr.ChainType)))
	}

	return nil
}
