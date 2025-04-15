package wallet

import (
	"context"
	"math/big"
	"sync"

	"vault0/internal/config"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/blockexplorer"
	"vault0/internal/core/keystore"
	"vault0/internal/core/tokenstore"
	coreWallet "vault0/internal/core/wallet"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

// Service defines the wallet service interface
type Service interface {
	BalanceService
	MonitorService
	HistoryService

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

	// ActivateToken creates a token balance for a wallet and token address.
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - walletAddress: The wallet's blockchain address
	//   - tokenAddress: The token's contract address
	// Returns:
	//   - error: ErrInvalidInput for invalid parameters, or any error from the token store
	ActivateToken(ctx context.Context, chainType types.ChainType, walletAddress, tokenAddress string) error
}

// walletService implements the Service interface
type walletService struct {
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	keystore             keystore.KeyStore
	tokenStore           tokenstore.TokenStore
	walletFactory        coreWallet.Factory
	blockchainRegistry   blockchain.Registry
	chains               *types.Chains
	txService            transaction.Service
	txRepository         transaction.Repository
	blockExplorerFactory blockexplorer.Factory
	mu                   sync.RWMutex

	// Transaction monitoring fields
	monitorCtx    context.Context
	monitorCancel context.CancelFunc

	// Transaction history syncing fields
	syncHistoryCtx    context.Context
	syncHistoryCancel context.CancelFunc
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
	txService transaction.Service,
	txRepository transaction.Repository,
	blockExplorerFactory blockexplorer.Factory,
) Service {
	return &walletService{
		config:               config,
		log:                  log,
		repository:           repository,
		keystore:             keyStore,
		tokenStore:           tokenStore,
		walletFactory:        walletFactory,
		blockchainRegistry:   blockchainRegistry,
		chains:               chains,
		txService:            txService,
		txRepository:         txRepository,
		blockExplorerFactory: blockExplorerFactory,
		mu:                   sync.RWMutex{},
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

	// Monitor the new wallet's address for transactions if transaction monitoring is active
	if s.monitorCtx != nil {
		s.monitorAddress(ctx, wallet.ChainType, wallet.Address)
	}

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

	// Unmonitor the wallet's address if transaction monitoring is active
	if s.monitorCtx != nil {
		s.unmonitorAddress(ctx, wallet.ChainType, wallet.Address)
	}

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

// ActivateToken creates a token balance for a wallet and token address.
func (s *walletService) ActivateToken(ctx context.Context, chainType types.ChainType, walletAddress, tokenAddress string) error {
	if chainType == "" {
		return errors.NewInvalidInputError("Chain type is required", "chain_type", "")
	}
	if walletAddress == "" {
		return errors.NewInvalidInputError("Wallet address is required", "wallet_address", "")
	}
	if tokenAddress == "" {
		return errors.NewInvalidInputError("Token address is required", "token_address", "")
	}

	// Validate and normalize wallet address
	normalizedWalletAddr, err := types.NewAddress(chainType, walletAddress)
	if err != nil {
		return err
	}
	normalizedWalletAddressStr := normalizedWalletAddr.ToChecksum()
	// Validate and normalize token address
	normalizedTokenAddr, err := types.NewAddress(chainType, tokenAddress)
	if err != nil {
		return err
	}
	normalizedTokenAddressStr := normalizedTokenAddr.ToChecksum()

	// Ensure wallet exists using the normalized wallet address
	wallet, err := s.repository.GetByAddress(ctx, chainType, normalizedWalletAddressStr)
	if err != nil {
		return err
	}

	// If the token balance already exists, log and return successfully
	exists, err := s.repository.TokenBalanceExists(ctx, wallet, normalizedTokenAddressStr)
	if err != nil {
		return err // Propagate repository error
	}

	if exists {
		s.log.Info("Token already active, skipping zero balance creation",
			logger.Int64("wallet_id", wallet.ID),
			logger.String("token_address", normalizedTokenAddressStr))
		return nil
	}

	// If the balance does not exist, proceed to create it
	s.log.Info("Activating token: creating zero balance entry",
		logger.Int64("wallet_id", wallet.ID),
		logger.String("token_address", normalizedTokenAddressStr))

	// Create token balance with initial value zero
	err = s.repository.UpdateTokenBalance(ctx, wallet, normalizedTokenAddressStr, big.NewInt(0))
	if err != nil {
		return err
	}

	return nil
}
