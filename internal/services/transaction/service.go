package transaction

import (
	"context"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/tokenstore"
	coreTx "vault0/internal/core/transaction" // Alias core transaction package
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

	// GetMappedTransactionByHash retrieves a transaction by hash and attempts to map it
	// to its specific type (e.g., ERC20Transfer, MultiSigWithdrawalRequest) using the Mapper.
	// Returns the specific typed transaction (as any) or the original *Transaction if unmappable.
	GetMappedTransactionByHash(ctx context.Context, hash string) (any, error)

	// GetERC20TransferByHash retrieves an ERC20 transfer transaction by hash.
	GetERC20TransferByHash(ctx context.Context, hash string) (*types.ERC20Transfer, error)

	// GetMultiSigWithdrawalRequestByHash retrieves a MultiSig withdrawal request transaction by hash.
	GetMultiSigWithdrawalRequestByHash(ctx context.Context, hash string) (*types.MultiSigWithdrawalRequest, error)
}

// transactionService implements the Service interface
type transactionService struct {
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	tokenStore           tokenstore.TokenStore // Keep for potential future use or direct lookups
	blockExplorerFactory blockexplorer.Factory
	chains               *types.Chains
	mapper               coreTx.Mapper // Added Mapper dependency
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
	mapper coreTx.Mapper, // Added mapper parameter
) Service {
	return &transactionService{
		config:               config,
		log:                  log,
		repository:           repository,
		tokenStore:           tokenStore,
		blockExplorerFactory: blockExplorerFactory,
		chains:               chains,
		mapper:               mapper, // Inject mapper
	}
}

// GetTransactionByHash retrieves a transaction by its hash
func (s *transactionService) GetTransactionByHash(ctx context.Context, hash string) (*Transaction, error) {
	if hash == "" {
		return nil, errors.NewMissingParameterError("hash")
	}

	// Get transaction directly from repository
	return s.repository.GetByTxHash(ctx, hash)
}

// FilterTransactions retrieves transactions based on the provided filter criteria
func (s *transactionService) FilterTransactions(ctx context.Context, filter *Filter, limit int, nextToken string) (*types.Page[*Transaction], error) {
	if limit <= 0 {
		limit = 10
	}

	return s.repository.List(ctx, filter, limit, nextToken)
}

// CreateWalletTransaction creates and saves a transaction received from the core layer
func (s *transactionService) CreateWalletTransaction(ctx context.Context, walletID int64, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewMissingParameterError("transaction")
	}
	if walletID <= 0 {
		return errors.NewInvalidInputError("Wallet ID must be positive", "wallet_id", walletID)
	}

	// Token resolution logic is REMOVED. The service now stores the generic transaction.
	// Mapping to specific types (like ERC20Transfer) happens on retrieval using the Mapper.

	// Convert core types.Transaction to service-layer Transaction and associate walletID
	serviceTx := FromCoreTransaction(tx, walletID)
	if serviceTx == nil {
		return errors.NewInternalError(fmt.Errorf("failed to convert core transaction %s to service model", tx.Hash))
	}

	err := s.repository.Create(ctx, serviceTx)
	if err != nil {
		s.log.Error("Failed to create wallet transaction in repository",
			logger.Error(err),
			logger.Int64("wallet_id", walletID),
			logger.String("tx_hash", tx.Hash),
		)
		return err
	}

	s.log.Info("Wallet transaction created successfully",
		logger.Int64("wallet_id", walletID),
		logger.String("tx_hash", tx.Hash))

	return nil
}

// GetMappedTransactionByHash retrieves a transaction and maps it to its specific type.
func (s *transactionService) GetMappedTransactionByHash(ctx context.Context, hash string) (any, error) {
	serviceTx, err := s.GetTransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	coreTx := serviceTx.ToCoreTransaction()
	if coreTx == nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to convert service transaction %s back to core type", hash))
	}

	mappedTx, err := s.mapper.ToTypedTransaction(ctx, coreTx)
	if err != nil {
		s.log.Warn("Failed to map transaction to specific type, returning generic transaction",
			logger.String("tx_hash", hash),
			logger.Error(err),
		)
		return coreTx, nil
	}

	return mappedTx, nil
}

// GetERC20TransferByHash retrieves an ERC20 transfer transaction by hash.
func (s *transactionService) GetERC20TransferByHash(ctx context.Context, hash string) (*types.ERC20Transfer, error) {
	mappedTx, err := s.GetMappedTransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	erc20Transfer, ok := mappedTx.(*types.ERC20Transfer)
	if !ok {
		// If it's not the expected type, it could be an unmapped tx or a different type
		s.log.Info("Transaction found but is not an ERC20 transfer", logger.String("tx_hash", hash))
		return nil, errors.NewNotFoundError(fmt.Sprintf("ERC20 transfer transaction with hash %s not found or is not an ERC20 transfer", hash))
	}

	return erc20Transfer, nil
}

// GetMultiSigWithdrawalRequestByHash retrieves a MultiSig withdrawal request transaction by hash.
func (s *transactionService) GetMultiSigWithdrawalRequestByHash(ctx context.Context, hash string) (*types.MultiSigWithdrawalRequest, error) {
	mappedTx, err := s.GetMappedTransactionByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	msWithdrawal, ok := mappedTx.(*types.MultiSigWithdrawalRequest)
	if !ok {
		s.log.Info("Transaction found but is not a MultiSig withdrawal request", logger.String("tx_hash", hash))
		return nil, errors.NewNotFoundError(fmt.Sprintf("MultiSig withdrawal request transaction with hash %s not found or is not a withdrawal request", hash))
	}

	return msWithdrawal, nil
}
