package wallet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"vault0/internal/config"
	"vault0/internal/keystore"
	"vault0/internal/types"
	coreWallet "vault0/internal/wallet"
)

// Common service errors
var (
	ErrWalletNotFound = errors.New("wallet not found")
	ErrInvalidInput   = errors.New("invalid input")
)

// Service defines the wallet service interface
type Service interface {
	// CreateWallet creates a new wallet
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

// WalletFactory defines the interface for wallet creation
// This matches the signature of *Factory.NewWallet from wallet package
type WalletFactory interface {
	NewWallet(ctx context.Context, chainType types.ChainType, keyID string) (coreWallet.Wallet, error)
}

// DefaultService is the default implementation of the wallet service
type DefaultService struct {
	repository    Repository
	walletFactory WalletFactory
	config        *config.Config
	keystore      keystore.KeyStore
}

// NewService creates a new wallet service
func NewService(repository Repository, keyStore keystore.KeyStore, config *config.Config) Service {
	// Create the wallet factory using the provided keystore and config
	walletFactory := coreWallet.NewFactory(keyStore, config)

	return &DefaultService{
		repository:    repository,
		walletFactory: walletFactory,
		config:        config,
		keystore:      keyStore,
	}
}

// CreateWallet creates a new wallet
func (s *DefaultService) CreateWallet(ctx context.Context, chainType types.ChainType, name string, tags map[string]string) (*Wallet, error) {
	// Validate inputs
	if name == "" {
		return nil, fmt.Errorf("%w: name cannot be empty", ErrInvalidInput)
	}

	// Create the wallet in the wallet module
	w, err := s.walletFactory.NewWallet(ctx, chainType, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Create the wallet in the wallet module
	walletInfo, err := w.Create(ctx, name, tags)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}

	// Store the wallet info in the database
	wallet := &Wallet{
		KeyID:     walletInfo.KeyID,
		ChainType: walletInfo.ChainType,
		Address:   walletInfo.Address,
		Name:      name,
		Tags:      tags,
	}

	if err := s.repository.Create(ctx, wallet); err != nil {
		return nil, fmt.Errorf("failed to store wallet: %w", err)
	}

	return wallet, nil
}

// UpdateWallet updates a wallet's name and tags
func (s *DefaultService) UpdateWallet(ctx context.Context, id, name string, tags map[string]string) (*Wallet, error) {
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
func (s *DefaultService) DeleteWallet(ctx context.Context, id string) error {
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
func (s *DefaultService) GetWallet(ctx context.Context, id string) (*Wallet, error) {
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
func (s *DefaultService) ListWallets(ctx context.Context, limit, offset int) ([]*Wallet, error) {
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
