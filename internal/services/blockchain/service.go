package blockchain

import (
	"context"
	"fmt"
	"sync"
	"time"
	"vault0/internal/core/blockchain"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Service defines the blockchain service interface
type Service interface {
	// Activate activates a blockchain by creating a wallet and starting event listening
	Activate(ctx context.Context, chainType types.ChainType) error
	// Deactivate deactivates a blockchain and stops event listening
	Deactivate(ctx context.Context, chainType types.ChainType) error
	// GetBlockchain retrieves blockchain information by chain type
	GetBlockchain(ctx context.Context, chainType types.ChainType) (*Blockchain, error)
	// ListActive lists all active blockchains
	ListActive(ctx context.Context) ([]*Blockchain, error)
	// SubscribeToEvents subscribes to events on all active blockchains
	SubscribeToEvents(ctx context.Context) error
	// UnsubscribeFromEvents unsubscribes from all event subscriptions
	UnsubscribeFromEvents()
}

type service struct {
	repository    Repository
	walletService wallet.Service
	factory       blockchain.Factory
	subscribers   map[types.ChainType]context.CancelFunc
	mu            sync.RWMutex
}

// NewService creates a new blockchain service
func NewService(repository Repository, walletSvc wallet.Service, factory blockchain.Factory) Service {
	return &service{
		repository:    repository,
		walletService: walletSvc,
		factory:       factory,
		subscribers:   make(map[types.ChainType]context.CancelFunc),
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
	chain, err := s.factory.NewBlockchain(chainType)
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

	// Subscribe to events
	if err := s.subscribeToChain(ctx, blockchain); err != nil {
		return fmt.Errorf("failed to subscribe to events: %w", err)
	}

	return nil
}

func (s *service) Deactivate(ctx context.Context, chainType types.ChainType) error {
	// Unsubscribe from events
	s.unsubscribeFromChain(chainType)

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

func (s *service) SubscribeToEvents(ctx context.Context) error {
	// Get all active blockchains
	blockchains, err := s.repository.FindActive(ctx)
	if err != nil {
		return fmt.Errorf("failed to get active blockchains: %w", err)
	}

	// Subscribe to each blockchain
	for _, blockchain := range blockchains {
		if err := s.subscribeToChain(ctx, blockchain); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", blockchain.ChainType, err)
		}
	}

	return nil
}

func (s *service) UnsubscribeFromEvents() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel all subscribers
	for chainType, cancel := range s.subscribers {
		cancel()
		delete(s.subscribers, chainType)
	}
}

func (s *service) subscribeToChain(ctx context.Context, blockchain *Blockchain) error {
	// Get wallet info
	walletInfo, err := s.walletService.GetWallet(ctx, blockchain.WalletID)
	if err != nil {
		return fmt.Errorf("failed to get wallet: %w", err)
	}

	// Create blockchain client
	client, err := s.factory.NewBlockchain(blockchain.ChainType)
	if err != nil {
		return fmt.Errorf("failed to create blockchain client: %w", err)
	}

	// Create context with cancellation
	subscriptionCtx, cancel := context.WithCancel(context.Background())

	// Store cancel function
	s.mu.Lock()
	s.subscribers[blockchain.ChainType] = cancel
	s.mu.Unlock()

	// Subscribe to events
	go func() {
		defer cancel()

		// Subscribe to events for the wallet address
		logCh, errCh, err := client.SubscribeToEvents(subscriptionCtx, []string{walletInfo.Address}, nil)
		if err != nil {
			fmt.Printf("Failed to subscribe to events for %s: %v\n", blockchain.ChainType, err)
			return
		}

		for {
			select {
			case <-subscriptionCtx.Done():
				return
			case err := <-errCh:
				fmt.Printf("Event subscription error for %s: %v\n", blockchain.ChainType, err)
				return
			case log := <-logCh:
				// Handle incoming transaction event
				fmt.Printf("Received event for %s: %+v\n", blockchain.ChainType, log)
			}
		}
	}()

	return nil
}

func (s *service) unsubscribeFromChain(chainType types.ChainType) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cancel, exists := s.subscribers[chainType]; exists {
		cancel()
		delete(s.subscribers, chainType)
	}
}
