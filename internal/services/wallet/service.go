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

// Event types for wallet lifecycle events
const (
	EventTypeWalletCreated = "WALLET_CREATED"
	EventTypeWalletDeleted = "WALLET_DELETED"
)

// BlockchainEvent represents a blockchain event associated with a wallet
type BlockchainEvent struct {
	WalletID string
	Log      *types.Log
}

// LifecycleEvent represents a wallet lifecycle event
type LifecycleEvent struct {
	WalletID  string
	EventType string
	ChainType types.ChainType
	Address   string
}

// Service defines the wallet service interface
type Service interface {
	// Create creates a new wallet with a key and derives its address.
	// It performs the following steps:
	// 1. Creates a new key in the keystore with the specified name and tags
	// 2. Creates a wallet instance using the key
	// 3. Derives the wallet's blockchain address
	// 4. Stores the wallet information in the database
	// 5. Sets up event subscription for the wallet
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type (e.g., Ethereum, Bitcoin)
	//   - name: Human-readable name for the wallet
	//   - tags: Optional metadata key-value pairs
	//
	// Returns:
	//   - *Wallet: The created wallet information
	//   - error: ErrInvalidInput if parameters are invalid, or any other error that occurred
	Create(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error)

	// Update modifies a wallet's name and tags.
	// The wallet is identified by its chain type and address.
	// Only the name and tags can be updated; other fields are immutable.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - address: The wallet's blockchain address
	//   - name: New name for the wallet
	//   - tags: New metadata key-value pairs
	//
	// Returns:
	//   - *Wallet: The updated wallet information
	//   - error: ErrWalletNotFound if wallet doesn't exist, ErrInvalidInput for invalid parameters
	Update(ctx context.Context, chainType types.ChainType, address, name string, tags map[string]string) (*Wallet, error)

	// Delete performs a soft delete of a wallet.
	// The wallet is identified by its chain type and address.
	// This operation:
	// 1. Marks the wallet as deleted in the database
	// 2. Unsubscribes from the wallet's blockchain events
	// The wallet's data is preserved but hidden from normal operations.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - address: The wallet's blockchain address
	//
	// Returns:
	//   - error: ErrWalletNotFound if wallet doesn't exist, ErrInvalidInput for invalid parameters
	Delete(ctx context.Context, chainType types.ChainType, address string) error

	// Get retrieves a wallet's information by its chain type and address.
	// Only returns non-deleted wallets.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - address: The wallet's blockchain address
	//
	// Returns:
	//   - *Wallet: The wallet information if found
	//   - error: ErrWalletNotFound if wallet doesn't exist, ErrInvalidInput for invalid parameters
	Get(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error)

	// GetByID retrieves a wallet by its unique identifier.
	// Only returns non-deleted wallets.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The wallet's unique identifier
	//
	// Returns:
	//   - *Wallet: The wallet information if found
	//   - error: ErrWalletNotFound if wallet doesn't exist, ErrInvalidInput for invalid parameters
	GetByID(ctx context.Context, id string) (*Wallet, error)

	// List retrieves a paginated list of non-deleted wallets.
	// Results are ordered by creation date (newest first).
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - limit: Maximum number of wallets to return (default: 10 if <= 0)
	//   - offset: Number of wallets to skip (default: 0 if < 0)
	//
	// Returns:
	//   - []*Wallet: List of wallets
	//   - error: Any error that occurred during the operation
	List(ctx context.Context, limit, offset int) ([]*Wallet, error)

	// Exists checks if a non-deleted wallet exists.
	// This is a lightweight operation that doesn't return the wallet's data.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - address: The wallet's blockchain address
	//
	// Returns:
	//   - bool: true if the wallet exists and is not deleted
	//   - error: ErrInvalidInput for invalid parameters
	Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error)

	// SubscribeToBlockchainEvents sets up event subscriptions for all active wallets.
	// For each wallet, it:
	// 1. Creates a blockchain client connection
	// 2. Subscribes to relevant blockchain events
	// 3. Processes events in a background goroutine
	//
	// Parameters:
	//   - ctx: Context for the operation
	//
	// Returns:
	//   - error: Any error that occurred during subscription setup
	SubscribeToBlockchainEvents(ctx context.Context) error

	// UnsubscribeFromBlockchainEvents cancels all active wallet event subscriptions.
	// This should be called when shutting down the service or
	// when event monitoring is no longer needed.
	UnsubscribeFromBlockchainEvents()

	// BlockchainEvents returns a channel that emits blockchain-related events.
	// These are events from the blockchain like transactions, token transfers, etc.
	// The channel is closed when UnsubscribeFromEvents is called.
	BlockchainEvents() <-chan *BlockchainEvent

	// LifecycleEvents returns a channel that emits wallet lifecycle events.
	// These are events like wallet creation, deletion, etc.
	// The channel is closed when UnsubscribeFromEvents is called.
	LifecycleEvents() <-chan *LifecycleEvent
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
	subscriptions      map[string]context.CancelFunc
	mu                 sync.RWMutex
	// Separate channels for different event types
	blockchainEvents chan *BlockchainEvent // Only blockchain events
	lifecycleEvents  chan *LifecycleEvent  // Only lifecycle events
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
	const channelBuffer = 100
	return &walletService{
		config:             config,
		logger:             logger,
		repository:         repository,
		keystore:           keyStore,
		walletFactory:      walletFactory,
		blockchainRegistry: blockchainRegistry,
		chains:             chains,
		subscriptions:      make(map[string]context.CancelFunc),
		mu:                 sync.RWMutex{},
		blockchainEvents:   make(chan *BlockchainEvent, channelBuffer),
		lifecycleEvents:    make(chan *LifecycleEvent, channelBuffer),
	}
}

// emitBlockchainEvent sends a blockchain event to the blockchain events channel
func (s *walletService) emitBlockchainEvent(event *BlockchainEvent) {
	select {
	case s.blockchainEvents <- event:
	default:
		s.logger.Warn("Blockchain events channel is full, dropping event",
			logger.String("wallet_id", event.WalletID))
	}
}

// emitLifecycleEvent sends a lifecycle event to the lifecycle events channel
func (s *walletService) emitLifecycleEvent(event *LifecycleEvent) {
	select {
	case s.lifecycleEvents <- event:
	default:
		s.logger.Warn("Lifecycle events channel is full, dropping event",
			logger.String("wallet_id", event.WalletID),
			logger.String("event_type", event.EventType))
	}
}

// Create creates a new wallet with a key and derives its address
func (s *walletService) Create(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error) {
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

	// Emit wallet created event
	s.emitLifecycleEvent(&LifecycleEvent{
		WalletID:  wallet.ID,
		EventType: EventTypeWalletCreated,
		ChainType: wallet.ChainType,
		Address:   wallet.Address,
	})

	return wallet, nil
}

// Update updates a wallet's name and tags by chain type and address
func (s *walletService) Update(ctx context.Context, chainType types.ChainType, address, name string, tags map[string]string) (*Wallet, error) {
	// Validate inputs
	if chainType == "" {
		return nil, fmt.Errorf("%w: chain type cannot be empty", ErrInvalidInput)
	}
	if address == "" {
		return nil, fmt.Errorf("%w: address cannot be empty", ErrInvalidInput)
	}
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}

	// Validate chain type
	chain, exists := s.chains[chainType]
	if !exists {
		return nil, fmt.Errorf("%w: unsupported chain type: %s", ErrInvalidInput, chainType)
	}

	// Validate address format
	if !chain.IsValidAddress(address) {
		return nil, fmt.Errorf("%w: invalid address format for chain type %s", ErrInvalidInput, chainType)
	}

	// Update the wallet
	wallet, err := s.repository.Update(ctx, chainType, address, name, tags)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	return wallet, nil
}

// Delete soft-deletes a wallet by chain type and address
func (s *walletService) Delete(ctx context.Context, chainType types.ChainType, address string) error {
	// Validate inputs
	if chainType == "" {
		return fmt.Errorf("%w: chain type cannot be empty", ErrInvalidInput)
	}
	if address == "" {
		return fmt.Errorf("%w: address cannot be empty", ErrInvalidInput)
	}

	// Validate chain type
	chain, exists := s.chains[chainType]
	if !exists {
		return fmt.Errorf("%w: unsupported chain type: %s", ErrInvalidInput, chainType)
	}

	// Validate address format
	if !chain.IsValidAddress(address) {
		return fmt.Errorf("%w: invalid address format for chain type %s", ErrInvalidInput, chainType)
	}

	// Get the wallet first to get its ID for unsubscribing
	wallet, err := s.repository.Get(ctx, chainType, address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWalletNotFound
		}
		return err
	}

	// Unsubscribe from wallet events first
	s.unsubscribeFromWallet(wallet.ID)

	// Then delete the wallet
	if err := s.repository.Delete(ctx, chainType, address); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWalletNotFound
		}
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	// Emit wallet deleted event
	s.emitLifecycleEvent(&LifecycleEvent{
		WalletID:  wallet.ID,
		EventType: EventTypeWalletDeleted,
		ChainType: wallet.ChainType,
		Address:   wallet.Address,
	})

	return nil
}

// Get retrieves a wallet by its chain type and address
func (s *walletService) Get(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
	// Validate inputs
	if chainType == "" {
		return nil, fmt.Errorf("%w: chain type cannot be empty", ErrInvalidInput)
	}
	if address == "" {
		return nil, fmt.Errorf("%w: address cannot be empty", ErrInvalidInput)
	}

	// Validate chain type
	chain, exists := s.chains[chainType]
	if !exists {
		return nil, fmt.Errorf("%w: unsupported chain type: %s", ErrInvalidInput, chainType)
	}

	// Validate address format
	if !chain.IsValidAddress(address) {
		return nil, fmt.Errorf("%w: invalid address format for chain type %s", ErrInvalidInput, chainType)
	}

	// Get the wallet
	wallet, err := s.repository.Get(ctx, chainType, address)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	return wallet, nil
}

// List retrieves a list of wallets
func (s *walletService) List(ctx context.Context, limit, offset int) ([]*Wallet, error) {
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

// SubscribeToBlockchainEvents subscribes to events for all active wallets
func (s *walletService) SubscribeToBlockchainEvents(ctx context.Context) error {
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

// UnsubscribeFromBlockchainEvents unsubscribes from all event subscriptions
func (s *walletService) UnsubscribeFromBlockchainEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel all subscribers
	for walletID, cancel := range s.subscriptions {
		cancel()
		delete(s.subscriptions, walletID)
	}

	// Close all event channels
	close(s.blockchainEvents)
	close(s.lifecycleEvents)
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
	s.subscriptions[wallet.ID] = cancel
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
				// Emit blockchain event
				s.emitBlockchainEvent(&BlockchainEvent{
					WalletID: wallet.ID,
					Log:      &log,
				})
			}
		}
	}()

	return nil
}

func (s *walletService) unsubscribeFromWallet(walletID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, exists := s.subscriptions[walletID]; exists {
		cancel()
		delete(s.subscriptions, walletID)
	}
}

// Exists checks if a wallet exists by its chain type and address
func (s *walletService) Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
	// Validate inputs
	if chainType == "" {
		return false, fmt.Errorf("%w: chain type cannot be empty", ErrInvalidInput)
	}
	if address == "" {
		return false, fmt.Errorf("%w: address cannot be empty", ErrInvalidInput)
	}

	// Validate chain type
	chain, exists := s.chains[chainType]
	if !exists {
		return false, fmt.Errorf("%w: unsupported chain type: %s", ErrInvalidInput, chainType)
	}

	// Validate address format
	if !chain.IsValidAddress(address) {
		return false, fmt.Errorf("%w: invalid address format for chain type %s", ErrInvalidInput, chainType)
	}

	// Check if the wallet exists
	return s.repository.Exists(ctx, chainType, address)
}

// GetByID retrieves a wallet by its unique identifier
func (s *walletService) GetByID(ctx context.Context, id string) (*Wallet, error) {
	// Validate input
	if id == "" {
		return nil, fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}

	// Get the wallet from repository
	wallet, err := s.repository.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWalletNotFound
		}
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	return wallet, nil
}

// BlockchainEvents returns the blockchain events channel
func (s *walletService) BlockchainEvents() <-chan *BlockchainEvent {
	return s.blockchainEvents
}

// LifecycleEvents returns the lifecycle events channel
func (s *walletService) LifecycleEvents() <-chan *LifecycleEvent {
	return s.lifecycleEvents
}
