package wallet

import (
	"context"
	"strconv"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/keystore"
	coreWallet "vault0/internal/core/wallet"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Event types for wallet lifecycle events
const (
	EventTypeWalletCreated = "WALLET_CREATED"
	EventTypeWalletDeleted = "WALLET_DELETED"
)

// BlockchainEvent represents a blockchain event associated with a wallet
type BlockchainEvent struct {
	WalletID int64
	Log      *types.Log
}

// LifecycleEvent represents a wallet lifecycle event
type LifecycleEvent struct {
	WalletID  int64
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

	// Update updates a wallet's name and tags by chain type and address.
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

	// UpdateLastBlockNumber updates the last block number for a wallet.
	// This method is used to track the last processed block from blockchain events.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - address: The wallet's blockchain address
	//   - blockNumber: The new last block number
	//
	// Returns:
	//   - error: ErrWalletNotFound if wallet doesn't exist, ErrInvalidInput for invalid parameters
	UpdateLastBlockNumber(ctx context.Context, chainType types.ChainType, address string, blockNumber int64) error

	// Delete soft-deletes a wallet by chain type and address.
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

	// GetByAddress retrieves a wallet by its chain type and address.
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
	GetByAddress(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error)

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
	GetByID(ctx context.Context, id int64) (*Wallet, error)

	// List retrieves a paginated list of non-deleted wallets
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - limit: Maximum number of wallets to return (default: 10 if <= 0)
	//   - offset: Number of wallets to skip (default: 0 if < 0)
	//
	// Returns:
	//   - *types.Page[*Wallet]: Paginated list of wallets
	//   - error: Any error that occurred during the operation
	List(ctx context.Context, limit, offset int) (*types.Page[*Wallet], error)

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
	log                logger.Logger
	repository         Repository
	keystore           keystore.KeyStore
	walletFactory      coreWallet.Factory
	blockchainRegistry blockchain.Registry
	chains             *types.Chains
	subscriptions      map[string]context.CancelFunc
	mu                 sync.RWMutex
	blockchainEvents   chan *BlockchainEvent
	lifecycleEvents    chan *LifecycleEvent
}

// NewService creates a new wallet service
func NewService(
	config *config.Config,
	log logger.Logger,
	repository Repository,
	keyStore keystore.KeyStore,
	walletFactory coreWallet.Factory,
	blockchainRegistry blockchain.Registry,
	chains *types.Chains,
) Service {
	const channelBuffer = 100
	return &walletService{
		config:             config,
		log:                log,
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
	if event == nil {
		s.log.Error("Received nil blockchain event",
			logger.String("func", "emitBlockchainEvent"))
		return
	}

	select {
	case s.blockchainEvents <- event:
	default:
		s.log.Warn("Blockchain events channel is full, dropping event",
			logger.Int64("wallet_id", event.WalletID))
	}
}

// emitLifecycleEvent sends a lifecycle event to the lifecycle events channel
func (s *walletService) emitLifecycleEvent(event *LifecycleEvent) {
	select {
	case s.lifecycleEvents <- event:
	default:
		s.log.Warn("Lifecycle events channel is full, dropping event",
			logger.Int64("wallet_id", event.WalletID),
			logger.String("event_type", event.EventType))
	}
}

// Create creates a new wallet with a key and derives its address
func (s *walletService) Create(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error) {
	if name == "" {
		return nil, errors.NewInvalidInputError("Name is required", "name", "")
	}

	chain, err := s.chains.Get(chainType)
	if err != nil {
		return nil, err
	}

	key, err := s.keystore.Create(ctx, name, chain.KeyType, chain.Curve, tags)
	if err != nil {
		return nil, err
	}

	w, err := s.walletFactory.NewWallet(ctx, chainType, key.ID)
	if err != nil {
		return nil, err
	}

	address, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, err
	}

	wallet := &Wallet{
		KeyID:     key.ID,
		ChainType: chain.Type,
		Address:   address,
		Name:      name,
		Tags:      tags,
	}

	if err := s.repository.Create(ctx, wallet); err != nil {
		return nil, errors.NewOperationFailedError("create wallet", err)
	}

	if err := s.subscribeToWallet(ctx, wallet); err != nil {
		s.log.Error("Failed to subscribe to wallet events",
			logger.Int64("wallet_id", wallet.ID),
			logger.String("address", wallet.Address),
			logger.Error(err))
	}

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
	if chainType == "" {
		return nil, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return nil, errors.NewInvalidInputError("Address is required", "address", "")
	}
	if name == "" {
		return nil, errors.NewInvalidInputError("Name is required", "name", "")
	}

	chain, err := s.chains.Get(chainType)
	if err != nil {
		return nil, err
	}

	if err := chain.ValidateAddress(address); err != nil {
		return nil, err
	}

	wallet, err := s.repository.GetByAddress(ctx, chainType, address)
	if err != nil {
		return nil, err
	}

	wallet.Name = name
	wallet.Tags = tags

	if err := s.repository.Update(ctx, wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

// UpdateLastBlockNumber updates the last block number for a wallet
func (s *walletService) UpdateLastBlockNumber(ctx context.Context, chainType types.ChainType, address string, blockNumber int64) error {
	if chainType == "" {
		return errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return errors.NewInvalidInputError("Address is required", "address", "")
	}
	if blockNumber < 0 {
		return errors.NewInvalidInputError("Block number cannot be negative", "block_number", blockNumber)
	}

	chain, err := s.chains.Get(chainType)
	if err != nil {
		return err
	}

	if !chain.IsValidAddress(address) {
		return errors.NewInvalidAddressError(address)
	}

	wallet, err := s.repository.GetByAddress(ctx, chainType, address)
	if err != nil {
		return err
	}

	wallet.LastBlockNumber = blockNumber

	if err := s.repository.Update(ctx, wallet); err != nil {
		return err
	}

	return nil
}

// Delete soft-deletes a wallet by chain type and address
func (s *walletService) Delete(ctx context.Context, chainType types.ChainType, address string) error {
	if chainType == "" {
		return errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return errors.NewInvalidInputError("Address is required", "address", "")
	}

	chain, err := s.chains.Get(chainType)
	if err != nil {
		return err
	}

	if !chain.IsValidAddress(address) {
		return errors.NewInvalidAddressError(address)
	}

	wallet, err := s.repository.GetByAddress(ctx, chainType, address)
	if err != nil {
		return err
	}

	s.unsubscribeFromWallet(wallet.ID)

	if err := s.repository.Delete(ctx, chainType, address); err != nil {
		return err
	}

	s.emitLifecycleEvent(&LifecycleEvent{
		WalletID:  wallet.ID,
		EventType: EventTypeWalletDeleted,
		ChainType: wallet.ChainType,
		Address:   wallet.Address,
	})

	return nil
}

// GetByAddress retrieves a wallet by its chain type and address
func (s *walletService) GetByAddress(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
	if chainType == "" {
		return nil, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return nil, errors.NewInvalidInputError("Address is required", "address", "")
	}

	chain, err := s.chains.Get(chainType)
	if err != nil {
		return nil, err
	}

	if !chain.IsValidAddress(address) {
		return nil, errors.NewInvalidAddressError(address)
	}

	wallet, err := s.repository.GetByAddress(ctx, chainType, address)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// List retrieves a paginated list of non-deleted wallets
func (s *walletService) List(ctx context.Context, limit, offset int) (*types.Page[*Wallet], error) {
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	wallets, err := s.repository.List(ctx, limit, offset)
	if err != nil {
		return nil, errors.NewOperationFailedError("list wallets", err)
	}

	return wallets, nil
}

// SubscribeToBlockchainEvents subscribes to events for all active wallets
func (s *walletService) SubscribeToBlockchainEvents(ctx context.Context) error {
	// Get all wallets without pagination to subscribe to all
	walletPage, err := s.repository.List(ctx, 0, 0)
	if err != nil {
		return errors.NewOperationFailedError("list wallets for subscription", err)
	}

	for _, wallet := range walletPage.Items {
		if err := s.subscribeToWallet(ctx, wallet); err != nil {
			s.log.Error("Failed to subscribe to wallet events",
				logger.Int64("wallet_id", wallet.ID),
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

	for walletID, cancel := range s.subscriptions {
		cancel()
		delete(s.subscriptions, walletID)
	}

	close(s.blockchainEvents)
	close(s.lifecycleEvents)
}

func (s *walletService) subscribeToWallet(ctx context.Context, wallet *Wallet) error {
	// Get blockchain client for the wallet's chain type
	client, err := s.blockchainRegistry.GetBlockchain(wallet.ChainType)
	if err != nil {
		return err
	}

	// Create a context with cancellation for this subscription
	subscriptionCtx, cancel := context.WithCancel(context.Background())

	// Store the cancellation function with the wallet ID as the key
	s.mu.Lock()
	walletIDStr := strconv.FormatInt(wallet.ID, 10)
	s.subscriptions[walletIDStr] = cancel
	s.mu.Unlock()

	// Start a goroutine to monitor events
	go func() {
		defer cancel()
		s.log.Info("Starting event subscription for wallet",
			logger.Int64("wallet_id", wallet.ID),
			logger.String("address", wallet.Address))

		// Subscribe to events, using wallet's LastBlockNumber as the starting point
		logCh, errCh, err := client.SubscribeContractLogs(subscriptionCtx, []string{wallet.Address}, nil, wallet.LastBlockNumber)
		if err != nil {
			s.log.Error("Failed to subscribe to events",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("address", wallet.Address),
				logger.Error(err))
			return
		}

		// Process events
		for {
			select {
			case <-subscriptionCtx.Done():
				s.log.Info("Event subscription stopped",
					logger.Int64("wallet_id", wallet.ID),
					logger.String("address", wallet.Address))
				return
			case err := <-errCh:
				s.log.Error("Event subscription error",
					logger.Int64("wallet_id", wallet.ID),
					logger.String("address", wallet.Address),
					logger.Error(err))
			case log := <-logCh:
				// Process the log
				s.emitBlockchainEvent(&BlockchainEvent{
					WalletID: wallet.ID,
					Log:      &log,
				})

				// Update the last processed block if log's block number is greater
				if log.BlockNumber != nil {
					blockNum := log.BlockNumber.Int64()
					if blockNum > wallet.LastBlockNumber {
						if err := s.UpdateLastBlockNumber(ctx, wallet.ChainType, wallet.Address, blockNum); err != nil {
							s.log.Error("Failed to update last block number",
								logger.Int64("wallet_id", wallet.ID),
								logger.String("address", wallet.Address),
								logger.Error(err))
						}
					}
				}
			}
		}
	}()

	return nil
}

func (s *walletService) unsubscribeFromWallet(walletID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	walletIDStr := strconv.FormatInt(walletID, 10)
	if cancel, exists := s.subscriptions[walletIDStr]; exists {
		cancel()
		delete(s.subscriptions, walletIDStr)
	}
}

// Exists checks if a wallet exists by its chain type and address
func (s *walletService) Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
	if chainType == "" {
		return false, errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return false, errors.NewInvalidInputError("Address is required", "address", "")
	}

	chain, err := s.chains.Get(chainType)
	if err != nil {
		return false, err
	}

	if !chain.IsValidAddress(address) {
		return false, errors.NewInvalidAddressError(address)
	}

	return s.repository.Exists(ctx, chainType, address)
}

// GetByID retrieves a wallet by its unique identifier
func (s *walletService) GetByID(ctx context.Context, id int64) (*Wallet, error) {
	if id == 0 {
		return nil, errors.NewInvalidInputError("ID is required", "id", "0")
	}

	wallet, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
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
