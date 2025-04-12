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
	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

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

	// UpdateWalletBalance updates the native balance for a wallet based on a transaction.
	// It determines if the transaction involves the wallet as sender or receiver
	// and updates the native balance accordingly, including gas costs for senders.
	// Parameters:
	//   - ctx: Context for the operation
	//   - tx: The transaction details
	// Returns:
	//   - error: ErrWalletNotFound if wallet doesn't exist, or other processing errors
	UpdateWalletBalance(ctx context.Context, tx *types.Transaction) error

	// UpdateTokenBalance updates the token balance for a wallet based on a transaction.
	// It identifies the involved wallet (sender or receiver) and updates the specific
	// token balance. For senders, it also deducts the native currency gas cost.
	// Parameters:
	//   - ctx: Context for the operation
	//   - tx: The transaction details (must be of type ERC20)
	// Returns:
	//   - error: ErrWalletNotFound, ErrTokenNotFound, or other processing errors
	UpdateTokenBalance(ctx context.Context, tx *types.Transaction) error

	// GetWalletBalances retrieves the native and token balances for a wallet
	GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalanceData, error)

	// GetWalletBalancesByAddress retrieves the native and token balances for a wallet by its address
	GetWalletBalancesByAddress(ctx context.Context, chainType types.ChainType, address string) ([]*TokenBalanceData, error)

	// StartTransactionMonitoring starts monitoring transactions for all wallets.
	// It performs the following steps:
	// 1. Retrieves all active wallets
	// 2. Monitors each wallet's address for transactions
	// 3. Processes incoming transaction events and saves them to the database
	//
	// Parameters:
	//   - ctx: Context for the operation
	//
	// Returns:
	//   - error: Any error that occurred during setup
	StartTransactionMonitoring(ctx context.Context) error

	// StopTransactionMonitoring stops monitoring transactions.
	// This should be called when shutting down the service.
	StopTransactionMonitoring()
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
	txService          transaction.Service
	mu                 sync.RWMutex
	// Transaction monitoring fields
	monitorCtx    context.Context
	monitorCancel context.CancelFunc
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
) Service {
	return &walletService{
		config:             config,
		log:                log,
		repository:         repository,
		keystore:           keyStore,
		tokenStore:         tokenStore,
		walletFactory:      walletFactory,
		blockchainRegistry: blockchainRegistry,
		chains:             chains,
		txService:          txService,
		mu:                 sync.RWMutex{},
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

// UpdateWalletBalance updates the native balance for a wallet based on a transaction
func (s *walletService) UpdateWalletBalance(ctx context.Context, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewInvalidInputError("Transaction is required", "tx", nil)
	}

	if tx.Type != types.TransactionTypeNative {
		// This method handles native balances, call UpdateTokenBalance for gas adjustments in token transfers
		// If the transaction is incoming token transfer, we don't need to do anything here
		return nil
	}

	isOutgoing := false

	// Determine if the wallet is the sender or receiver
	wallet, err := s.repository.GetByAddress(ctx, tx.Chain, tx.From)
	if err == nil {
		isOutgoing = true
	} else {
		wallet, err = s.repository.GetByAddress(ctx, tx.Chain, tx.To)
		if err != nil {
			// Wallet involved in the transaction not found in our DB
			return errors.NewWalletNotFoundError(tx.From + " or " + tx.To)
		}
		// Wallet is the receiver, isFrom remains false
	}

	// Get the current balance
	currentBalance := wallet.Balance.ToBigInt()
	var newBalance *big.Int

	if isOutgoing {
		// Outgoing native transaction: subtract amount + gas
		gasUsed := new(big.Int).SetUint64(tx.GasUsed)
		totalSpent := new(big.Int).Add(tx.Value, new(big.Int).Mul(tx.GasPrice, gasUsed))
		newBalance = new(big.Int).Sub(currentBalance, totalSpent)
		if newBalance.Sign() < 0 {
			newBalance = big.NewInt(0) // Ensure balance doesn't go negative
		}
	} else {
		// Incoming native transaction: add amount
		newBalance = new(big.Int).Add(currentBalance, tx.Value)
	}

	// Update the wallet balance in the repository
	return s.repository.UpdateBalance(ctx, wallet.ID, newBalance)
}

// UpdateTokenBalance updates the token balance for a wallet based on a transaction
func (s *walletService) UpdateTokenBalance(ctx context.Context, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewInvalidInputError("Transaction is required", "tx", nil)
	}
	if tx.Type != types.TransactionTypeERC20 {
		return errors.NewInvalidInputError("Transaction must be ERC20 type", "tx.Type", tx.Type)
	}
	if tx.TokenAddress == "" {
		return errors.NewInvalidInputError("Token address is required for ERC20 transaction", "tx.TokenAddress", "")
	}

	isOutgoing := false

	// Determine if the wallet is the sender or receiver
	wallet, err := s.repository.GetByAddress(ctx, tx.Chain, tx.From)
	if err == nil {
		isOutgoing = true
	} else {
		wallet, err = s.repository.GetByAddress(ctx, tx.Chain, tx.To)
		if err != nil {
			// Wallet involved in the transaction not found in our DB
			return errors.NewWalletNotFoundError(tx.From + " or " + tx.To)
		}
		// Wallet is the receiver, isFrom remains false
	}

	// --- Update Token Balance ---
	// Get the current token balance
	tokenBalances, err := s.repository.GetTokenBalances(ctx, wallet.ID)
	if err != nil {
		return err
	}

	// Find the specific token balance
	var currentTokenBalance *big.Int
	found := false
	for _, tb := range tokenBalances {
		if tb.TokenAddress == tx.TokenAddress {
			currentTokenBalance = tb.Balance.ToBigInt()
			found = true
			break
		}
	}

	// If not found, start with zero balance
	if !found {
		currentTokenBalance = big.NewInt(0)
	}

	// Calculate new token balance based on transaction direction
	var newTokenBalance *big.Int
	if isOutgoing {
		// Outgoing transaction: subtract amount
		newTokenBalance = new(big.Int).Sub(currentTokenBalance, tx.Value) // ERC20 value is the token amount
		if newTokenBalance.Sign() < 0 {
			newTokenBalance = big.NewInt(0) // Prevent negative balance
		}
	} else {
		// Incoming transaction: add amount
		newTokenBalance = new(big.Int).Add(currentTokenBalance, tx.Value)
	}

	// Update the token balance in the repository
	if err := s.repository.UpdateTokenBalance(ctx, wallet.ID, tx.TokenAddress, newTokenBalance); err != nil {
		return err // Return the error if token balance update fails
	}

	// --- Update Native Balance for Gas (Sender Only) ---
	if isOutgoing {
		// Get the current native balance
		currentNativeBalance := wallet.Balance.ToBigInt()

		// Calculate gas cost
		gasUsed := new(big.Int).SetUint64(tx.GasUsed)
		gasCost := new(big.Int).Mul(tx.GasPrice, gasUsed)

		// Calculate new native balance
		newNativeBalance := new(big.Int).Sub(currentNativeBalance, gasCost)
		if newNativeBalance.Sign() < 0 {
			newNativeBalance = big.NewInt(0) // Ensure balance doesn't go negative
		}

		// Update the native balance in the repository
		if err := s.repository.UpdateBalance(ctx, wallet.ID, newNativeBalance); err != nil {
			// Log error, but don't necessarily fail the whole operation if token balance updated successfully
			s.log.Error("Failed to update sender native balance for gas during token transfer",
				logger.Int64("wallet_id", wallet.ID),
				logger.String("tx_hash", tx.Hash),
				logger.Error(err))
		}
	}

	return nil
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
	nativeToken, err := wallet.GetToken()
	if err != nil {
		return nil, err
	}

	// Add native token balance
	result = append(result, &TokenBalanceData{
		Token:     nativeToken,
		Balance:   wallet.Balance.ToBigInt(),
		UpdatedAt: wallet.UpdatedAt,
	})

	// If there are no token balances, return the native token balance
	if len(tokenBalances) == 0 {
		return result, nil
	}

	// Process each token balance
	for _, tb := range tokenBalances {
		// Get token by address
		token, err := s.tokenStore.GetToken(ctx, tb.TokenAddress)
		if err != nil {
			s.log.Warn("Could not find token",
				logger.String("token_address", tb.TokenAddress),
				logger.Error(err))
			continue
		}

		result = append(result, &TokenBalanceData{
			Token:     token,
			Balance:   tb.Balance.ToBigInt(),
			UpdatedAt: tb.UpdatedAt,
		})
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

// StartTransactionMonitoring starts monitoring transactions for all wallets
func (s *walletService) StartTransactionMonitoring(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already monitoring
	if s.monitorCtx != nil {
		s.log.Info("Transaction monitoring is already active")
		return nil
	}

	// Create a new context with cancel function for monitoring
	s.monitorCtx, s.monitorCancel = context.WithCancel(context.Background())

	// Start transaction event subscription if not already started
	s.txService.SubscribeToTransactionEvents(s.monitorCtx)

	// Get all active wallets
	wallets, err := s.repository.List(ctx, 0, 0) // Get all wallets
	if err != nil {
		s.monitorCancel()
		s.monitorCtx = nil
		s.monitorCancel = nil
		return err
	}

	// Monitor each wallet's address
	for _, wallet := range wallets.Items {
		if err := s.monitorAddress(ctx, wallet.ChainType, wallet.Address); err != nil {
			s.log.Warn("Failed to monitor wallet address",
				logger.String("address", wallet.Address),
				logger.String("chain_type", string(wallet.ChainType)),
				logger.Error(err))
		}
	}

	// Start a goroutine to process transaction events
	go s.processTransactionEvents(s.monitorCtx)

	s.log.Info("Started transaction monitoring",
		logger.Int("wallet_count", len(wallets.Items)))

	return nil
}

// StopTransactionMonitoring stops monitoring transactions
func (s *walletService) StopTransactionMonitoring() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.monitorCtx == nil {
		return // Not monitoring
	}

	// Cancel the monitoring context to stop the goroutine
	s.monitorCancel()

	// Unsubscribe from transaction events
	s.txService.UnsubscribeFromTransactionEvents()

	// Reset context and cancel function
	s.monitorCtx = nil
	s.monitorCancel = nil

	s.log.Info("Stopped transaction monitoring")
}

// monitorAddress registers a wallet address for transaction monitoring
func (s *walletService) monitorAddress(ctx context.Context, chainType types.ChainType, address string) error {
	// Register with transaction service
	addr := &types.Address{
		ChainType: chainType,
		Address:   address,
	}

	if err := s.txService.MonitorAddress(ctx, addr); err != nil {
		return err
	}

	s.log.Debug("Started monitoring address for transactions",
		logger.String("address", address),
		logger.String("chain_type", string(chainType)))

	return nil
}

// unmonitorAddress stops monitoring a wallet address for transactions
func (s *walletService) unmonitorAddress(ctx context.Context, chainType types.ChainType, address string) error {
	// Unregister with transaction service
	addr := &types.Address{
		ChainType: chainType,
		Address:   address,
	}

	if err := s.txService.UnmonitoredAddress(ctx, addr); err != nil {
		return err
	}

	s.log.Debug("Stopped monitoring address for transactions",
		logger.String("address", address),
		logger.String("chain_type", string(chainType)))

	return nil
}

// processTransactionEvents listens for transaction events and processes them
func (s *walletService) processTransactionEvents(ctx context.Context) {
	// Get the transaction events channel
	txEventsChan := s.txService.TransactionEvents()

	for {
		select {
		case <-ctx.Done():
			// Context cancelled, stop processing
			return

		case tx, ok := <-txEventsChan:
			if !ok {
				// Channel closed, stop processing
				s.log.Warn("Transaction events channel closed")
				return
			}

			// Process the transaction
			s.handleTransaction(ctx, tx)
		}
	}
}

// handleTransaction processes a single transaction
func (s *walletService) handleTransaction(ctx context.Context, tx *types.Transaction) {
	if tx == nil {
		return
	}

	// Check if the transaction already exists by hash
	existingTx, err := s.txService.GetTransaction(ctx, tx.Hash)

	if err != nil || existingTx == nil {
		// Transaction doesn't exist yet, log and update balances
		s.log.Info("Processing new transaction",
			logger.String("tx_hash", tx.Hash),
			logger.String("chain", string(tx.Chain)),
			logger.String("from", tx.From),
			logger.String("to", tx.To),
			logger.String("type", string(tx.Type)))

		// Update wallet balances if the transaction was successful
		if tx.Status == types.TransactionStatusSuccess {
			// Handle different transaction types
			if tx.Type == types.TransactionTypeNative {
				if err := s.UpdateWalletBalance(ctx, tx); err != nil {
					s.log.Error("Failed to update native balance from transaction",
						logger.String("tx_hash", tx.Hash),
						logger.Error(err))
				}
			} else if tx.Type == types.TransactionTypeERC20 {
				if err := s.UpdateTokenBalance(ctx, tx); err != nil {
					s.log.Error("Failed to update token balance from transaction",
						logger.String("tx_hash", tx.Hash),
						logger.String("token_address", tx.TokenAddress),
						logger.Error(err))
				}
			}
		}
	} else {
		s.log.Debug("Transaction already exists in database",
			logger.String("tx_hash", tx.Hash))
	}
}
