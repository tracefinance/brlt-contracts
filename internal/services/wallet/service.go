package wallet

import (
	"context"
	"math/big"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/keystore"
	"vault0/internal/core/tokenstore"
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

	// LifecycleEvents returns a channel that emits wallet lifecycle events.
	// These are events like wallet creation, deletion, etc.
	// The channel is closed when UnsubscribeFromEvents is called.
	LifecycleEvents() <-chan *LifecycleEvent

	// UpdateWalletBalance updates the native balance for a wallet identified by chain type and address
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - address: The wallet's blockchain address
	//   - balance: The new balance value
	// Returns:
	//   - error: ErrWalletNotFound if wallet doesn't exist, ErrInvalidInput for invalid parameters
	UpdateWalletBalance(ctx context.Context, chainType types.ChainType, address string, balance *big.Float) error

	// UpdateTokenBalance updates a token balance for a wallet
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - walletAddress: The wallet's blockchain address
	//   - tokenAddress: The token's contract address
	//   - balance: The new token balance value
	// Returns:
	//   - error: ErrWalletNotFound or ErrTokenNotFound if wallet or token doesn't exist,
	//            ErrInvalidInput for invalid parameters
	UpdateTokenBalance(ctx context.Context, chainType types.ChainType, walletAddress, tokenAddress string, balance *big.Float) error

	// GetWalletBalances retrieves the native and token balances for a wallet
	GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalanceData, error)

	// GetWalletBalancesByAddress retrieves the native and token balances for a wallet by its address
	GetWalletBalancesByAddress(ctx context.Context, chainType types.ChainType, address string) ([]*TokenBalanceData, error)
}

// walletService implements the Service interface
type walletService struct {
	config             *config.Config
	log                logger.Logger
	repository         Repository
	keystore           keystore.KeyStore
	tokenStore         tokenstore.TokenStore
	walletFactory      coreWallet.Factory
	blockchainRegistry blockchain.Registry
	chains             *types.Chains
	mu                 sync.RWMutex
	lifecycleEvents    chan *LifecycleEvent
}

// NewService creates a new wallet service
func NewService(
	config *config.Config,
	log logger.Logger,
	repository Repository,
	keyStore keystore.KeyStore,
	tokenStore tokenstore.TokenStore,
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
		tokenStore:         tokenStore,
		walletFactory:      walletFactory,
		blockchainRegistry: blockchainRegistry,
		chains:             chains,
		mu:                 sync.RWMutex{},
		lifecycleEvents:    make(chan *LifecycleEvent, channelBuffer),
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
		return nil, err
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
		return nil, err
	}

	return wallets, nil
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

// LifecycleEvents returns the lifecycle events channel
func (s *walletService) LifecycleEvents() <-chan *LifecycleEvent {
	return s.lifecycleEvents
}

// UpdateWalletBalance updates the native balance for a wallet
func (s *walletService) UpdateWalletBalance(ctx context.Context, chainType types.ChainType, address string, balance *big.Float) error {
	if chainType == "" {
		return errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if address == "" {
		return errors.NewInvalidInputError("Address is required", "address", "")
	}
	if balance.Sign() < 0 {
		return errors.NewInvalidInputError("Balance cannot be negative", "balance", balance.String())
	}

	// Validate chain type and address
	chain, err := s.chains.Get(chainType)
	if err != nil {
		return err
	}

	if !chain.IsValidAddress(address) {
		return errors.NewInvalidAddressError(address)
	}

	// Get wallet by address
	wallet, err := s.repository.GetByAddress(ctx, chainType, address)
	if err != nil {
		return err
	}

	// Update the wallet balance
	return s.repository.UpdateBalance(ctx, wallet.ID, balance)
}

// UpdateTokenBalance updates a token balance for a wallet
func (s *walletService) UpdateTokenBalance(ctx context.Context, chainType types.ChainType, walletAddress, tokenAddress string, balance *big.Float) error {
	if chainType == "" {
		return errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if walletAddress == "" {
		return errors.NewInvalidInputError("Wallet address is required", "wallet_address", "")
	}
	if tokenAddress == "" {
		return errors.NewInvalidInputError("Token address is required", "token_address", "")
	}
	if balance.Sign() < 0 {
		return errors.NewInvalidInputError("Balance cannot be negative", "balance", balance.String())
	}

	// Validate chain type and addresses
	chain, err := s.chains.Get(chainType)
	if err != nil {
		return err
	}

	if !chain.IsValidAddress(walletAddress) {
		return errors.NewInvalidAddressError(walletAddress)
	}

	if !chain.IsValidAddress(tokenAddress) {
		return errors.NewInvalidAddressError(tokenAddress)
	}

	// Get wallet by address
	wallet, err := s.repository.GetByAddress(ctx, chainType, walletAddress)
	if err != nil {
		return err
	}

	// Get token by address and chain type
	token, err := s.tokenStore.GetToken(ctx, tokenAddress, chainType)
	if err != nil {
		return err
	}

	// Update the token balance
	return s.repository.UpdateTokenBalance(ctx, wallet.ID, token.ID, balance)
}

// GetWalletBalances retrieves the native and token balances for a wallet
func (s *walletService) GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalanceData, error) {
	if id == 0 {
		return nil, errors.NewInvalidInputError("ID is required", "id", "0")
	}

	// Get the wallet to access its native balance and chain type
	wallet, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Get token balances
	tokenBalances, err := s.repository.GetTokenBalances(ctx, id)
	if err != nil {
		return nil, err
	}

	// Create result slice with enough capacity for native token + all token balances
	result := make([]*TokenBalanceData, 0, len(tokenBalances)+1)

	// Get the native token for this chain
	nativeToken, err := s.tokenStore.GetNativeToken(ctx, wallet.ChainType)
	if err != nil {
		return nil, err
	} else {
		// Add native token balance
		result = append(result, &TokenBalanceData{
			Token:     nativeToken,
			Balance:   wallet.Balance,
			UpdatedAt: wallet.UpdatedAt,
		})
	}

	// If there are no token balances, return the native token balance
	if len(tokenBalances) == 0 {
		return result, nil
	}

	// If there are token balances, fetch all tokens at once
	// Extract token IDs
	tokenIDs := make([]int64, len(tokenBalances))
	for i, tb := range tokenBalances {
		tokenIDs[i] = tb.TokenID
	}

	// Fetch all tokens in a single call
	tokensPage, err := s.tokenStore.ListTokensByIDs(ctx, tokenIDs, 0, 0) // No pagination limit
	if err != nil {
		return nil, err
	}

	// Create a map for quick token lookup by ID
	tokenMap := make(map[int64]*types.Token, len(tokensPage.Items))
	for i := range tokensPage.Items {
		tokenMap[tokensPage.Items[i].ID] = &tokensPage.Items[i]
	}

	// Add token balances
	for _, tb := range tokenBalances {
		if token, ok := tokenMap[tb.TokenID]; ok {
			result = append(result, &TokenBalanceData{
				Token:     token,
				Balance:   tb.Balance,
				UpdatedAt: tb.UpdatedAt,
			})
		} else {
			s.log.Warn("Token not found in results",
				logger.Int64("token_id", tb.TokenID))
		}
	}

	return result, nil
}

// GetWalletBalancesByAddress retrieves the native and token balances for a wallet by its address
func (s *walletService) GetWalletBalancesByAddress(ctx context.Context, chainType types.ChainType, address string) ([]*TokenBalanceData, error) {
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

	return s.GetWalletBalances(ctx, wallet.ID)
}
