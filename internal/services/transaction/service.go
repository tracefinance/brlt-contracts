package transaction

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/logger"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Common service errors
var (
	ErrInvalidInput = errors.New("invalid input")
)

// Service defines the transaction service interface
type Service interface {
	// GetTransaction retrieves a transaction by its chain type and hash
	GetTransaction(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error)

	// GetTransactionsByWallet retrieves transactions for a specific wallet
	GetTransactionsByWallet(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error)

	// GetTransactionsByAddress retrieves transactions for a specific blockchain address
	GetTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) ([]*Transaction, error)

	// SyncTransactions fetches and stores transactions for a wallet
	SyncTransactions(ctx context.Context, walletID string) (int, error)

	// SyncTransactionsByAddress fetches and stores transactions for an address
	SyncTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error)

	// CountTransactions counts transactions for a specific wallet
	CountTransactions(ctx context.Context, walletID string) (int, error)
}

// transactionService implements the Service interface
type transactionService struct {
	config               *config.Config
	logger               logger.Logger
	repository           Repository
	walletService        wallet.Service
	blockExplorerFactory blockexplorer.Factory
	chains               types.Chains
	syncMutex            sync.Mutex
}

// NewService creates a new transaction service
func NewService(
	config *config.Config,
	logger logger.Logger,
	repository Repository,
	walletService wallet.Service,
	blockExplorerFactory blockexplorer.Factory,
	chains types.Chains,
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
		return nil, fmt.Errorf("%w: chain type is required", ErrInvalidInput)
	}
	if hash == "" {
		return nil, fmt.Errorf("%w: hash is required", ErrInvalidInput)
	}

	// Try to get from database first
	tx, err := s.repository.Get(ctx, chainType, hash)
	if err == nil {
		return tx, nil
	}

	// If not found in database, fetch from blockchain explorer
	if errors.Is(err, ErrTransactionNotFound) {
		// Get explorer for the chain
		explorer, err := s.blockExplorerFactory.GetExplorer(chainType)
		if err != nil {
			return nil, fmt.Errorf("failed to get explorer: %w", err)
		}
		defer explorer.Close()

		// Fetch transaction from explorer
		txs, err := explorer.GetTransactionsByHash(ctx, []string{hash})
		if err != nil {
			return nil, fmt.Errorf("failed to get transaction from explorer: %w", err)
		}

		if len(txs) == 0 {
			return nil, ErrTransactionNotFound
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
func (s *transactionService) GetTransactionsByWallet(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
	// Validate input
	if walletID == "" {
		return nil, fmt.Errorf("%w: wallet ID is required", ErrInvalidInput)
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
func (s *transactionService) GetTransactionsByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) ([]*Transaction, error) {
	// Validate input
	if chainType == "" {
		return nil, fmt.Errorf("%w: chain type is required", ErrInvalidInput)
	}
	if address == "" {
		return nil, fmt.Errorf("%w: address is required", ErrInvalidInput)
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
		return 0, fmt.Errorf("%w: wallet ID is required", ErrInvalidInput)
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
		return 0, fmt.Errorf("%w: chain type is required", ErrInvalidInput)
	}
	if address == "" {
		return 0, fmt.Errorf("%w: address is required", ErrInvalidInput)
	}

	// Get explorer for the chain
	explorer, err := s.blockExplorerFactory.GetExplorer(chainType)
	if err != nil {
		return 0, fmt.Errorf("failed to get explorer: %w", err)
	}
	defer explorer.Close()

	// Get wallet ID if exists
	var walletID string
	wallet, err := s.walletService.Get(ctx, chainType, address)
	if err == nil {
		walletID = wallet.ID
	}

	// Prepare options for fetching transactions
	options := blockexplorer.TransactionHistoryOptions{
		StartBlock: 0,
		EndBlock:   0, // Latest block
		Page:       1,
		PageSize:   100,
		TransactionTypes: []blockexplorer.TransactionType{
			blockexplorer.TxTypeNormal,
			blockexplorer.TxTypeInternal,
			blockexplorer.TxTypeERC20,
			blockexplorer.TxTypeERC721,
		},
		SortAscending: false,
	}

	// Fetch transactions from explorer
	txs, err := explorer.GetTransactionHistory(ctx, address, options)
	if err != nil {
		return 0, fmt.Errorf("failed to get transaction history: %w", err)
	}

	// Save transactions to database
	count := 0
	for _, coreTx := range txs {
		// Check if transaction already exists
		exists, err := s.repository.Exists(ctx, chainType, coreTx.Hash)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to check transaction existence: %v, hash: %s", err, coreTx.Hash))
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
			s.logger.Warn(fmt.Sprintf("Failed to save transaction: %v, hash: %s", err, coreTx.Hash))
			continue
		}

		count++
	}

	return count, nil
}

// CountTransactions counts transactions for a specific wallet
func (s *transactionService) CountTransactions(ctx context.Context, walletID string) (int, error) {
	// Validate input
	if walletID == "" {
		return 0, fmt.Errorf("%w: wallet ID is required", ErrInvalidInput)
	}

	return s.repository.Count(ctx, walletID)
}
