package blockchain

import (
	"context"
	"fmt"
	"time"
	"vault0/internal/core/blockchain"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Service defines the blockchain service interface
type Service interface {
	// Activate activates a blockchain by creating a wallet
	Activate(ctx context.Context, chainType types.ChainType) error
	// Deactivate deactivates a blockchain
	Deactivate(ctx context.Context, chainType types.ChainType) error
	// GetBlockchain retrieves blockchain information by chain type
	GetBlockchain(ctx context.Context, chainType types.ChainType) (*Blockchain, error)
	// ListActive lists all active blockchains
	ListActive(ctx context.Context) ([]*Blockchain, error)
}

type service struct {
	repository    Repository
	walletService wallet.Service
	registry      blockchain.Registry
}

// NewService creates a new blockchain service
func NewService(repository Repository, walletSvc wallet.Service, registry blockchain.Registry) Service {
	return &service{
		repository:    repository,
		walletService: walletSvc,
		registry:      registry,
	}
}

func (s *service) Activate(ctx context.Context, chainType types.ChainType) error {
	// Check if blockchain already exists
	existing, err := s.repository.FindByChainType(ctx, chainType)
	if err == nil && existing != nil && existing.IsActive() {
		return fmt.Errorf("blockchain %s is already active", chainType)
	}

	// Create wallet for the blockchain
	walletInfo, err := s.walletService.CreateWallet(ctx, chainType, fmt.Sprintf("%s-wallet", chainType), map[string]string{
		"type":  "blockchain",
		"chain": string(chainType),
	})
	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	// Get chain information
	chain, err := s.registry.GetBlockchain(chainType)
	if err != nil {
		return fmt.Errorf("failed to create blockchain client: %w", err)
	}

	// Create blockchain record
	blockchain := &Blockchain{
		ChainType: chainType,
		ChainID:   chain.Chain().ID,
		WalletID:  walletInfo.ID,
		CreatedAt: time.Now(),
	}

	if err := s.repository.Create(ctx, blockchain); err != nil {
		return fmt.Errorf("failed to create blockchain record: %w", err)
	}

	return nil
}

func (s *service) Deactivate(ctx context.Context, chainType types.ChainType) error {
	// Update blockchain record
	if err := s.repository.Deactivate(ctx, chainType); err != nil {
		return fmt.Errorf("failed to deactivate blockchain: %w", err)
	}

	return nil
}

func (s *service) GetBlockchain(ctx context.Context, chainType types.ChainType) (*Blockchain, error) {
	return s.repository.FindByChainType(ctx, chainType)
}

func (s *service) ListActive(ctx context.Context) ([]*Blockchain, error) {
	return s.repository.FindActive(ctx)
}
