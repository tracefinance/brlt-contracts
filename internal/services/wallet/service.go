package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/core/keystore"
	coreWallet "vault0/internal/core/wallet"
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
}

// WalletService implements the Service interface
type WalletService struct {
	config        *config.Config
	repository    Repository
	walletFactory coreWallet.Factory
	chainFactory  types.ChainFactory
	keystore      keystore.KeyStore
}

// NewService creates a new wallet service
func NewService(config *config.Config, repository Repository, keyStore keystore.KeyStore, chainFactory types.ChainFactory, walletFactory coreWallet.Factory) Service {
	return &WalletService{
		config:        config,
		repository:    repository,
		walletFactory: walletFactory,
		chainFactory:  chainFactory,
		keystore:      keyStore,
	}
}

// CreateWallet creates a new wallet with a key and derives its address
func (s *WalletService) CreateWallet(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error) {
	// Validate inputs
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}

	// Get chain information first to determine the appropriate key type and curve
	chain, err := s.chainFactory.NewChain(chainType)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain information: %w", err)
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

	return wallet, nil
}

// UpdateWallet updates a wallet's name and tags
func (s *WalletService) UpdateWallet(ctx context.Context, id, name string, tags map[string]string) (*Wallet, error) {
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
func (s *WalletService) DeleteWallet(ctx context.Context, id string) error {
	// Validate inputs
	if id == "" {
		return fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
	}

	// Delete the wallet
	err := s.repository.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWalletNotFound
		}
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	return nil
}

// GetWallet retrieves a wallet by its ID
func (s *WalletService) GetWallet(ctx context.Context, id string) (*Wallet, error) {
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
func (s *WalletService) ListWallets(ctx context.Context, limit, offset int) ([]*Wallet, error) {
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
