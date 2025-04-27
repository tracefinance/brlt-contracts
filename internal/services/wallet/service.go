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
	coretx "vault0/internal/core/transaction"
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

	// CreateWallet creates a new wallet with a key and derives its address.
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
	CreateWallet(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error)

	// UpdateWallet updates a wallet's name and tags by chain type and address.
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
	UpdateWallet(ctx context.Context, chainType types.ChainType, address, name string, tags map[string]string) (*Wallet, error)

	// DeleteWallet soft-deletes a wallet by chain type and address.
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
	DeleteWallet(ctx context.Context, chainType types.ChainType, address string) error

	// GetWalletByAddress retrieves a wallet by its chain type and address.
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
	GetWalletByAddress(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error)

	// GetWalletByID retrieves a wallet by its unique identifier.
	// Only returns non-deleted wallets.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - id: The wallet's unique identifier
	//
	// Returns:
	//   - *Wallet: The wallet information if found
	//   - error: ErrWalletNotFound if wallet doesn't exist, ErrInvalidInput for invalid parameters
	GetWalletByID(ctx context.Context, id int64) (*Wallet, error)

	// ListWallets retrieves a paginated list of non-deleted wallets
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - limit: Maximum number of wallets to return (default: 10 if <= 0)
	//   - nextToken: Token for fetching the next page of results (empty string for first page)
	//
	// Returns:
	//   - *types.Page[*Wallet]: Paginated list of wallets with nextToken
	//   - error: Any error that occurred during the operation
	ListWallets(ctx context.Context, limit int, nextToken string) (*types.Page[*Wallet], error)

	// WalletExists checks if a non-deleted wallet exists.
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
	WalletExists(ctx context.Context, chainType types.ChainType, address string) (bool, error)

	// ActivateToken creates a token balance for a wallet and token address.
	// Parameters:
	//   - ctx: Context for the operation
	//   - chainType: The blockchain network type
	//   - walletAddress: The wallet's blockchain address
	//   - tokenAddress: The token's contract address
	// Returns:
	//   - error: ErrInvalidInput for invalid parameters, or any error from the token store
	ActivateToken(ctx context.Context, chainType types.ChainType, walletAddress, tokenAddress string) error

	// FindWalletsByKeyID retrieves all non-deleted wallets associated with a specific keystore key ID.
	// This is used internally, for example, to check if a key can be safely deleted.
	//
	// Parameters:
	//   - ctx: Context for the operation
	//   - keyID: The ID of the keystore key
	//
	// Returns:
	//   - []*Wallet: A slice of wallets associated with the key ID
	//   - error: Any error that occurred during the database query
	FindWalletsByKeyID(ctx context.Context, keyID string) ([]*Wallet, error)
}

// walletService implements the Service interface
type walletService struct {
	config               *config.Config
	log                  logger.Logger
	repository           Repository
	keystore             keystore.KeyStore
	tokenStore           tokenstore.TokenStore
	walletFactory        coreWallet.Factory
	blockchainFactory    blockchain.Factory
	chains               *types.Chains
	txService            transaction.Service
	txRepository         transaction.Repository
	txMonitor            coretx.Monitor
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
	blockchainRegistry blockchain.Factory,
	chains *types.Chains,
	txService transaction.Service,
	txRepository transaction.Repository,
	txMonitor coretx.Monitor,
	blockExplorerFactory blockexplorer.Factory,
) Service {
	return &walletService{
		config:               config,
		log:                  log,
		repository:           repository,
		keystore:             keyStore,
		tokenStore:           tokenStore,
		walletFactory:        walletFactory,
		blockchainFactory:    blockchainRegistry,
		chains:               chains,
		txService:            txService,
		txRepository:         txRepository,
		txMonitor:            txMonitor,
		blockExplorerFactory: blockExplorerFactory,
		mu:                   sync.RWMutex{},
	}
}

// CreateWallet creates a new wallet with a key and derives its address
func (s *walletService) CreateWallet(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error) {
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

// UpdateWallet updates a wallet's name and tags by chain type and address
func (s *walletService) UpdateWallet(ctx context.Context, chainType types.ChainType, address, name string, tags map[string]string) (*Wallet, error) {
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

// DeleteWallet soft-deletes a wallet by chain type and address
func (s *walletService) DeleteWallet(ctx context.Context, chainType types.ChainType, address string) error {
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

// GetWalletByAddress retrieves a wallet by its chain type and address
func (s *walletService) GetWalletByAddress(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
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

// ListWallets retrieves a paginated list of non-deleted wallets
func (s *walletService) ListWallets(ctx context.Context, limit int, nextToken string) (*types.Page[*Wallet], error) {
	// Only apply default limit for negative values, pass limit=0 to get all items
	if limit < 0 {
		limit = 10
	}

	return s.repository.List(ctx, limit, nextToken)
}

// WalletExists checks if a wallet exists by its chain type and address
func (s *walletService) WalletExists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
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

// GetWalletByID retrieves a wallet by its unique identifier
func (s *walletService) GetWalletByID(ctx context.Context, id int64) (*Wallet, error) {
	if id == 0 {
		return nil, errors.NewInvalidInputError("ID is required", "id", "0")
	}

	wallet, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// ActivateToken creates a token balance for a wallet and token address
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

// FindWalletsByKeyID retrieves all non-deleted wallets associated with a specific keystore key ID.
func (s *walletService) FindWalletsByKeyID(ctx context.Context, keyID string) ([]*Wallet, error) {
	if keyID == "" {
		return nil, errors.NewInvalidInputError("Key ID cannot be empty", "key_id", "")
	}
	s.log.Debug("Retrieving wallets by key ID",
		logger.String("key_id", keyID))
	wallets, err := s.repository.GetWalletsByKeyID(ctx, keyID)
	if err != nil {
		s.log.Error("Failed to get wallets by key ID from repository",
			logger.Error(err),
			logger.String("key_id", keyID))
		return nil, err
	}
	s.log.Debug("Successfully retrieved wallets by key ID",
		logger.String("key_id", keyID),
		logger.Int("count", len(wallets)))
	return wallets, nil
}
