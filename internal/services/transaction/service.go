package transaction

import (
	"context"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Service defines the transaction service interface
type Service interface {
	MonitorService

	// GetTransactionByHash retrieves a transaction by its hash
	GetTransactionByHash(ctx context.Context, hash string) (*Transaction, error)

	// FilterTransactions retrieves transactions based on the provided filter criteria
	FilterTransactions(ctx context.Context, filter *Filter, limit int, nextToken string) (*types.Page[*Transaction], error)

	// CreateWalletTransaction creates and saves a transaction for a specific wallet
	CreateWalletTransaction(ctx context.Context, walletID int64, tx *types.Transaction) error
}

// transactionService implements the Service interface
type transactionService struct {
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	tokenStore           tokenstore.TokenStore
	blockExplorerFactory blockexplorer.Factory
	chains               *types.Chains
	pendingPollingCtx    context.Context
	pendingPollingCancel context.CancelFunc
}

// NewService creates a new transaction service
func NewService(
	config *config.Config,
	log logger.Logger,
	repository Repository,
	tokenStore tokenstore.TokenStore,
	blockExplorerFactory blockexplorer.Factory,
	chains *types.Chains,
) Service {
	return &transactionService{
		config:               config,
		log:                  log,
		repository:           repository,
		tokenStore:           tokenStore,
		blockExplorerFactory: blockExplorerFactory,
		chains:               chains,
	}
}

// GetTransactionByHash retrieves a transaction by its hash
func (s *transactionService) GetTransactionByHash(ctx context.Context, hash string) (*Transaction, error) {
	if hash == "" {
		return nil, errors.NewInvalidInputError("Hash is required", "hash", "")
	}

	// Get transaction directly from repository
	return s.repository.GetByTxHash(ctx, hash)
}

// FilterTransactions retrieves transactions based on the provided filter criteria
func (s *transactionService) FilterTransactions(ctx context.Context, filter *Filter, limit int, nextToken string) (*types.Page[*Transaction], error) {
	// Set default limit
	if limit <= 0 {
		limit = 10
	}

	// Use the repository to filter transactions
	return s.repository.List(ctx, filter, limit, nextToken)
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

// CreateWalletTransaction creates and saves a transaction for a specific wallet
func (s *transactionService) CreateWalletTransaction(ctx context.Context, walletID int64, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewInvalidInputError("Transaction is required", "transaction", nil)
	}
	if walletID <= 0 {
		return errors.NewInvalidInputError("Wallet ID is required", "wallet_id", walletID)
	}

	// Resolve token symbol and decimals
	switch tx.Type {
	case types.TransactionTypeERC20:
		if tx.TokenAddress == "" {
			return errors.NewInvalidInputError("Token address is required for ERC20 transaction", "token_address", tx.TokenAddress)
		}
		token, err := s.tokenStore.GetToken(ctx, tx.TokenAddress)
		if err != nil || token == nil {
			s.log.Warn("ERC20 token not found in token store",
				logger.String("chain", string(tx.Chain)),
				logger.String("token_address", tx.TokenAddress),
				logger.Error(err),
			)
		} else {
			tx.TokenSymbol = token.Symbol
		}
	case types.TransactionTypeNative:
		nativeToken, err := types.NewNativeToken(tx.Chain)
		if err != nil {
			s.log.Warn("Failed to resolve native token",
				logger.String("chain", string(tx.Chain)),
				logger.Error(err),
			)
			tx.TokenSymbol = "UNKNOWN"
		} else {
			tx.TokenSymbol = nativeToken.Symbol
		}
	}

	// Convert types.Transaction to service-layer Transaction and associate walletID
	serviceTx := FromCoreTransaction(tx, walletID)

	// Save to database using repository
	err := s.repository.Create(ctx, serviceTx)
	if err != nil {
		s.log.Error("Failed to create wallet transaction",
			logger.Error(err),
			logger.Int64("wallet_id", walletID),
			logger.String("tx_hash", tx.Hash),
		)
		return errors.NewOperationFailedError("create wallet transaction", err)
	}

	return nil
}
