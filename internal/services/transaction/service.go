package transaction

import (
	"context"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/tokenstore"
	coreTx "vault0/internal/core/transaction" // Alias core transaction package
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Service defines the transaction service interface
type Service interface {
	// GetTransactionByHash retrieves a transaction by hash and attempts to map it
	// to its specific type (e.g., ERC20Transfer, MultiSigWithdrawalRequest) using the Mapper.
	// Returns the specific typed transaction as a CoreTransaction interface.
	GetTransactionByHash(ctx context.Context, hash string) (types.CoreTransaction, error)

	// FilterTransactions retrieves transactions based on filter criteria and maps each to its specific type
	// Returns a page of mapped transactions as CoreTransaction interfaces
	FilterTransactions(ctx context.Context, filter *Filter, limit int, nextToken string) (*types.Page[types.CoreTransaction], error)
}

// transactionService implements the Service interface
type transactionService struct {
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	tokenStore           tokenstore.TokenStore
	blockExplorerFactory blockexplorer.Factory
	chains               *types.Chains
	txFactory            coreTx.Factory
	blockchainFactory    blockchain.Factory
	transformer          TransformerService
}

// NewService creates a new transaction service
func NewService(
	config *config.Config,
	log logger.Logger,
	repository Repository,
	tokenStore tokenstore.TokenStore,
	blockExplorerFactory blockexplorer.Factory,
	chains *types.Chains,
	txFactory coreTx.Factory,
	blokchainFactory blockchain.Factory,
	transformer TransformerService,
) Service {
	return &transactionService{
		config:               config,
		log:                  log,
		repository:           repository,
		tokenStore:           tokenStore,
		blockExplorerFactory: blockExplorerFactory,
		chains:               chains,
		txFactory:            txFactory,
		blockchainFactory:    blokchainFactory,
		transformer:          transformer,
	}
}

// GetTransactionByHash retrieves a transaction and maps it to its specific type.
func (s *transactionService) GetTransactionByHash(ctx context.Context, hash string) (types.CoreTransaction, error) {
	serviceTx, err := s.repository.GetByHash(ctx, hash)
	if err != nil {
		return nil, err
	}

	coreTx := serviceTx.ToCoreTransaction()
	if coreTx == nil {
		return nil, errors.NewInternalError(fmt.Errorf("failed to convert service transaction %s back to core type", hash))
	}

	mapper, err := s.txFactory.NewMapper(coreTx.ChainType)
	if err != nil {
		return nil, err
	}

	mappedTx, err := mapper.ToTypedTransaction(ctx, coreTx)
	if err != nil {
		s.log.Warn("Failed to map transaction to specific type, returning generic transaction",
			logger.String("tx_hash", hash),
			logger.Error(err),
		)
		return coreTx, nil
	}

	// Ensure the mapped transaction implements CoreTransaction interface
	coreTransaction, ok := mappedTx.(types.CoreTransaction)
	if !ok {
		s.log.Warn("Mapped transaction does not implement CoreTransaction interface, returning generic transaction",
			logger.String("tx_hash", hash),
		)
		return coreTx, nil
	}

	return coreTransaction, nil
}

// FilterTransactions retrieves transactions based on filter criteria and maps each to its specific type
func (s *transactionService) FilterTransactions(ctx context.Context, filter *Filter, limit int, nextToken string) (*types.Page[types.CoreTransaction], error) {
	if limit <= 0 {
		limit = 10
	}

	// Get paginated transactions from repository
	txPage, err := s.repository.List(ctx, filter, limit, nextToken)
	if err != nil {
		return nil, err
	}

	// Create a new page with the same limit and next token
	mappedItems := make([]types.CoreTransaction, 0, len(txPage.Items))

	// Map each transaction to its specific type
	for _, tx := range txPage.Items {
		coreTx := tx.ToCoreTransaction()
		if coreTx == nil {
			s.log.Warn("Failed to convert service transaction to core type",
				logger.String("tx_hash", tx.Hash),
			)
			// Skip this transaction as we can't include non-CoreTransaction items
			continue
		}

		mapper, err := s.txFactory.NewMapper(coreTx.ChainType)
		if err != nil {
			s.log.Warn("Failed to create mapper for transaction",
				logger.String("tx_hash", tx.Hash),
				logger.String("chain_type", string(coreTx.ChainType)),
				logger.Error(err),
			)
			// Include the original coreTx which implements CoreTransaction
			mappedItems = append(mappedItems, coreTx)
			continue
		}

		mappedTx, err := mapper.ToTypedTransaction(ctx, coreTx)
		if err != nil {
			s.log.Warn("Failed to map transaction to specific type",
				logger.String("tx_hash", tx.Hash),
				logger.Error(err),
			)
			// Include the original coreTx which implements CoreTransaction
			mappedItems = append(mappedItems, coreTx)
			continue
		}

		// Ensure the mapped transaction implements CoreTransaction interface
		coreTransaction, ok := mappedTx.(types.CoreTransaction)
		if !ok {
			s.log.Warn("Mapped transaction does not implement CoreTransaction interface, returning generic transaction",
				logger.String("tx_hash", tx.Hash),
			)
			// Include the original coreTx which implements CoreTransaction
			mappedItems = append(mappedItems, coreTx)
			continue
		}

		mappedItems = append(mappedItems, coreTransaction)
	}

	// Create a new page with mapped items
	return &types.Page[types.CoreTransaction]{
		Items:     mappedItems,
		NextToken: txPage.NextToken,
		Limit:     txPage.Limit,
	}, nil
}
