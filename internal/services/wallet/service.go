package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/keystore"
	coreWallet "vault0/internal/core/wallet"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Common service errors
var (
	ErrWalletNotFound = errors.New("wallet not found")
	ErrInvalidInput   = errors.New("invalid input")
)

// Service defines the wallet service interface
type Service interface {
	// CreateWallet creates a new wallet with a key and derives its address
	CreateWallet(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error)

	// UpdateWallet updates a wallet's name and tags
	UpdateWallet(ctx context.Context, id, name string, tags map[string]string) (*Wallet, error)

	// DeleteWallet soft-deletes a wallet
	DeleteWallet(ctx context.Context, id string) error

	// GetWallet retrieves a wallet by its ID
	GetWallet(ctx context.Context, id string) (*Wallet, error)

	// ListWallets retrieves a list of wallets
	ListWallets(ctx context.Context, limit, offset int) ([]*Wallet, error)

	// SubscribeToEvents subscribes to events for all active wallets
	SubscribeToEvents(ctx context.Context) error

	// UnsubscribeFromEvents unsubscribes from all event subscriptions
	UnsubscribeFromEvents()
}

// walletService implements the Service interface
type walletService struct {
	config             *config.Config
	logger             logger.Logger
	repository         Repository
	keystore           keystore.KeyStore
	walletFactory      coreWallet.Factory
	blockchainRegistry blockchain.Registry
	chains             types.Chains
	subscribers        map[string]context.CancelFunc
	mu                 sync.RWMutex
}

// NewService creates a new wallet service
func NewService(
	config *config.Config,
	logger logger.Logger,
	repository Repository,
	keyStore keystore.KeyStore,
	walletFactory coreWallet.Factory,
	blockchainRegistry blockchain.Registry,
	chains types.Chains,
) Service {
	return &walletService{
		config:             config,
		logger:             logger,
		repository:         repository,
		keystore:           keyStore,
		walletFactory:      walletFactory,
		blockchainRegistry: blockchainRegistry,
		chains:             chains,
		subscribers:        make(map[string]context.CancelFunc),
		mu:                 sync.RWMutex{},
	}
}

// CreateWallet creates a new wallet with a key and derives its address
func (s *walletService) CreateWallet(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error) {
	// Validate inputs
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}

	// Get chain information from chains
	chain, exists := s.chains[chainType]
	if !exists {
		return nil, fmt.Errorf("unsupported chain type: %s", chainType)
	}

	// Create the key in the keystore using the chain's specified key type and curve
	key, err := s.keystore.Create(ctx, name, chain.KeyType, chain.Curve, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to create key: %w", err)
	}

	// Create a wallet instance using the factory
	w, err := s.walletFactory.NewWallet(ctx, chainType, key.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Derive the wallet address
	address, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	// Store the wallet info in the database
	wallet := &Wallet{
		KeyID:     key.ID,
		ChainType: chain.Type,
		Address:   address,
		Name:      name,
		Tags:      tags,
	}

	if err := s.repository.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	// Subscribe to wallet events
	if err := s.subscribeToWallet(wallet); err != nil {
		s.logger.Error("Failed to subscribe to wallet events",
			logger.String("wallet_id", wallet.ID),
			logger.String("address", wallet.Address),
			logger.Error(err))
	}

	return wallet, nil
}

// UpdateWallet updates a wallet's name and tags
func (s *walletService) UpdateWallet(ctx context.Context, id, name string, tags map[string]string) (*Wallet, error) {
	// Validate inputs
	if id == "" {
		return nil, fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}

	// Get the current wallet
	wallet, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	// Update the wallet
	wallet.Name = name
	wallet.Tags = tags

	if err := s.repository.Update(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return wallet, nil
}

// DeleteWallet soft-deletes a wallet
func (s *walletService) DeleteWallet(ctx context.Context, id string) error {
	// Unsubscribe from wallet events first
	s.unsubscribeFromWallet(id)

	// Then delete the wallet
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	return nil
}

// GetWallet retrieves a wallet by its ID
func (s *walletService) GetWallet(ctx context.Context, id string) (*Wallet, error) {
	// Validate inputs
	if id == "" {
		return nil, fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}

	// Get the wallet
	wallet, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	return wallet, nil
}

// ListWallets retrieves a list of wallets
func (s *walletService) ListWallets(ctx context.Context, limit, offset int) ([]*Wallet, error) {
	// Set default pagination values if not provided
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get the wallets
	wallets, err := s.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list wallets: %w", err)
	}

	return wallets, nil
}

// SubscribeToEvents subscribes to events for all active wallets
func (s *walletService) SubscribeToEvents(ctx context.Context) error {
	// Get all non-deleted wallets
	wallets, err := s.repository.List(ctx, 0, 0)
	if err != nil {
		return fmt.Errorf("failed to list wallets: %w", err)
	}

	// Subscribe to each wallet
	for _, wallet := range wallets {
		if err := s.subscribeToWallet(wallet); err != nil {
			s.logger.Error("Failed to subscribe to wallet events",
				logger.String("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.Error(err))
		}
	}

	return nil
}

// UnsubscribeFromEvents unsubscribes from all event subscriptions
func (s *walletService) UnsubscribeFromEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel all subscribers
	for walletID, cancel := range s.subscribers {
		cancel()
		delete(s.subscribers, walletID)
	}
}

func (s *walletService) subscribeToWallet(wallet *Wallet) error {
	// Create blockchain client
	client, err := s.blockchainRegistry.GetBlockchain(wallet.ChainType)
	if err != nil {
		return fmt.Errorf("failed to create blockchain client: %w", err)
	}

	// Create context with cancellation
	subscriptionCtx, cancel := context.WithCancel(context.Background())

	// Store cancel function
	s.mu.Lock()
	s.subscribers[wallet.ID] = cancel
	s.mu.Unlock()

	// Subscribe to events
	go func() {
		defer cancel()

		// Subscribe to events for the wallet address
		logCh, errCh, err := client.SubscribeToEvents(subscriptionCtx, []string{wallet.Address}, nil)
		if err != nil {
			s.logger.Error("Failed to subscribe to events",
				logger.String("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.Error(err))
			return
		}

		for {
			select {
			case <-subscriptionCtx.Done():
				return
			case err := <-errCh:
				s.logger.Error("Event subscription error",
					logger.String("wallet_id", wallet.ID),
					logger.String("address", wallet.Address),
					logger.Error(err))
				return
			case log := <-logCh:
				s.logger.Info("Received transaction event",
					logger.String("wallet_id", wallet.ID),
					logger.String("address", wallet.Address),
					logger.String("tx_hash", log.TransactionHash))
			}
		}
	}()

	return nil
}

func (s *walletService) unsubscribeFromWallet(walletID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, exists := s.subscribers[walletID]; exists {
		cancel()
		delete(s.subscribers, walletID)
	}
}
