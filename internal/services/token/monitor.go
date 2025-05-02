package token

import (
	"context"
	"vault0/internal/core/tokenstore"
	"vault0/internal/logger"
	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

// TokenMonitorService defines an interface for monitoring token transactions
type TokenMonitorService interface {
	// StartTokenTransactionMonitoring starts monitoring all tokens for transfer events
	// and also watches for token lifecycle events to automatically monitor new tokens
	StartTokenTransactionMonitoring(ctx context.Context) error

	// StopTokenTransactionMonitoring stops monitoring all token addresses
	// and stops watching for token lifecycle events
	StopTokenTransactionMonitoring() error
}

type tokenMonitorService struct {
	log              logger.Logger
	tokenStore       tokenstore.TokenStore
	txMonitorService transaction.MonitorService
	lifecycleCtx     context.Context
	lifecycleCancel  context.CancelFunc
	isMonitoring     bool
}

// NewTokenMonitorService creates a new TokenMonitorService
func NewTokenMonitorService(
	log logger.Logger,
	tokenStore tokenstore.TokenStore,
	txMonitorService transaction.MonitorService,
) TokenMonitorService {
	return &tokenMonitorService{
		log:              log,
		tokenStore:       tokenStore,
		txMonitorService: txMonitorService,
	}
}

// StartTokenTransactionMonitoring starts monitoring all tokens for transfer events
// and watches for token lifecycle events
func (s *tokenMonitorService) StartTokenTransactionMonitoring(ctx context.Context) error {
	s.log.Info("Starting token transaction monitoring")

	// Check if already monitoring
	if s.isMonitoring {
		s.log.Info("Token transaction monitoring is already active")
		return nil
	}

	// Create a cancellable context for lifecycle monitoring
	s.lifecycleCtx, s.lifecycleCancel = context.WithCancel(ctx)
	s.isMonitoring = true

	// Get all tokens from token store
	// Using 0 as limit to get all tokens in one query
	tokensPage, err := s.tokenStore.ListTokens(ctx, 0, "")
	if err != nil {
		s.log.Error("Failed to retrieve tokens from token store", logger.Error(err))
		return err
	}

	// Track successful monitoring count
	successCount := 0

	// Monitor each token contract address for Transfer events
	if len(tokensPage.Items) > 0 {
		for _, token := range tokensPage.Items {
			address, err := types.NewAddress(token.ChainType, token.Address)
			if err != nil {
				s.log.Error("Invalid token address",
					logger.String("token_address", token.Address),
					logger.String("token_symbol", token.Symbol),
					logger.String("chain_type", string(token.ChainType)),
					logger.Error(err),
				)
				continue
			}

			// Only monitor for the Transfer event
			events := []string{string(types.ERC20TransferEventSignature)}

			err = s.txMonitorService.MonitorContractAddress(*address, events)
			if err != nil {
				s.log.Error("Failed to monitor token contract",
					logger.String("token_address", token.Address),
					logger.String("token_symbol", token.Symbol),
					logger.String("chain_type", string(token.ChainType)),
					logger.Error(err),
				)
				continue
			}

			successCount++
			s.log.Debug("Started monitoring token contract",
				logger.String("token_address", token.Address),
				logger.String("token_symbol", token.Symbol),
				logger.String("chain_type", string(token.ChainType)),
			)
		}

		s.log.Info("Existing token monitoring started",
			logger.Int("total_tokens", len(tokensPage.Items)),
			logger.Int("successfully_monitored", successCount),
		)
	} else {
		s.log.Info("No existing tokens found to monitor")
	}

	// Start monitoring token lifecycle events
	tokenEvents := s.tokenStore.TokenEvents()
	if tokenEvents == nil {
		s.log.Error("Failed to get token events channel, cannot monitor token lifecycle")
		// Continue anyway since we've already set up monitoring for existing tokens
	} else {
		// Start the lifecycle monitoring goroutine
		go s.monitorTokenLifecycle(s.lifecycleCtx, tokenEvents)
		s.log.Info("Token lifecycle monitoring started")
	}

	s.log.Info("Token transaction monitoring started successfully")
	return nil
}

// StopTokenTransactionMonitoring stops monitoring all token addresses
// and stops watching for token lifecycle events
func (s *tokenMonitorService) StopTokenTransactionMonitoring() error {
	s.log.Info("Stopping token transaction monitoring")

	// Stop lifecycle monitoring first
	if s.isMonitoring && s.lifecycleCancel != nil {
		s.lifecycleCancel()
		s.isMonitoring = false
		s.lifecycleCtx = nil
		s.lifecycleCancel = nil
		s.log.Info("Token lifecycle monitoring stopped")
	}

	// Create a simple context for the token listing operation
	ctx := context.Background()

	// Get all tokens from token store
	tokensPage, err := s.tokenStore.ListTokens(ctx, 0, "")
	if err != nil {
		s.log.Error("Failed to retrieve tokens from token store", logger.Error(err))
		return err
	}

	if len(tokensPage.Items) == 0 {
		s.log.Info("No tokens found to unmonitor")
		return nil
	}

	// Track successful unmonitoring count
	successCount := 0

	// Unmonitor each token contract address
	for _, token := range tokensPage.Items {
		address, err := types.NewAddress(token.ChainType, token.Address)
		if err != nil {
			s.log.Error("Invalid token address",
				logger.String("token_address", token.Address),
				logger.String("token_symbol", token.Symbol),
				logger.String("chain_type", string(token.ChainType)),
				logger.Error(err),
			)
			continue
		}

		err = s.txMonitorService.UnmonitorContractAddress(*address)
		if err != nil {
			s.log.Error("Failed to unmonitor token contract",
				logger.String("token_address", token.Address),
				logger.String("token_symbol", token.Symbol),
				logger.String("chain_type", string(token.ChainType)),
				logger.Error(err),
			)
			continue
		}

		successCount++
		s.log.Debug("Stopped monitoring token contract",
			logger.String("token_address", token.Address),
			logger.String("token_symbol", token.Symbol),
			logger.String("chain_type", string(token.ChainType)),
		)
	}

	s.log.Info("Token transaction monitoring stopped",
		logger.Int("total_tokens", len(tokensPage.Items)),
		logger.Int("successfully_unmonitored", successCount),
	)

	return nil
}

// monitorTokenLifecycle handles token events and manages address monitoring accordingly
func (s *tokenMonitorService) monitorTokenLifecycle(ctx context.Context, tokenEvents <-chan tokenstore.TokenEvent) {
	s.log.Info("Starting token lifecycle event processor")

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Stopping token lifecycle event processor due to context cancellation")
			return

		case event, ok := <-tokenEvents:
			if !ok {
				s.log.Warn("Token events channel closed, stopping lifecycle processor")
				return
			}

			if event.Token == nil {
				s.log.Warn("Received token event with nil token, skipping")
				continue
			}

			// Process the token event based on its type
			switch event.EventType {
			case tokenstore.TokenEventAdded:
				s.handleTokenAdded(event.Token)

			case tokenstore.TokenEventUpdated:
				s.handleTokenUpdated(event.Token)

			case tokenstore.TokenEventDeleted:
				s.handleTokenDeleted(event.Token)

			default:
				s.log.Warn("Unknown token event type received",
					logger.String("event_type", string(event.EventType)),
				)
			}
		}
	}
}

// handleTokenAdded starts monitoring a newly added token
func (s *tokenMonitorService) handleTokenAdded(token *types.Token) {
	address, err := types.NewAddress(token.ChainType, token.Address)
	if err != nil {
		s.log.Error("Invalid token address for newly added token",
			logger.String("token_address", token.Address),
			logger.String("token_symbol", token.Symbol),
			logger.String("chain_type", string(token.ChainType)),
			logger.Error(err),
		)
		return
	}

	// Only monitor for the Transfer event
	events := []string{string(types.ERC20TransferEventSignature)}

	err = s.txMonitorService.MonitorContractAddress(*address, events)
	if err != nil {
		s.log.Error("Failed to monitor newly added token contract",
			logger.String("token_address", token.Address),
			logger.String("token_symbol", token.Symbol),
			logger.String("chain_type", string(token.ChainType)),
			logger.Error(err),
		)
		return
	}

	s.log.Info("Started monitoring newly added token contract",
		logger.String("token_address", token.Address),
		logger.String("token_symbol", token.Symbol),
		logger.String("chain_type", string(token.ChainType)),
	)
}

// handleTokenUpdated handles token update events
// Currently, we don't need special logic for token updates since the address doesn't change,
// but this method provides a hook for future enhancements
func (s *tokenMonitorService) handleTokenUpdated(token *types.Token) {
	// Currently, we don't need to do anything special for token updates
	// since we're monitoring based on the contract address which doesn't change
	s.log.Debug("Token updated event received, no action needed",
		logger.String("token_address", token.Address),
		logger.String("token_symbol", token.Symbol),
	)
}

// handleTokenDeleted stops monitoring a deleted token
func (s *tokenMonitorService) handleTokenDeleted(token *types.Token) {
	address, err := types.NewAddress(token.ChainType, token.Address)
	if err != nil {
		s.log.Error("Invalid token address for deleted token",
			logger.String("token_address", token.Address),
			logger.String("token_symbol", token.Symbol),
			logger.String("chain_type", string(token.ChainType)),
			logger.Error(err),
		)
		return
	}

	err = s.txMonitorService.UnmonitorContractAddress(*address)
	if err != nil {
		s.log.Error("Failed to unmonitor deleted token contract",
			logger.String("token_address", token.Address),
			logger.String("token_symbol", token.Symbol),
			logger.String("chain_type", string(token.ChainType)),
			logger.Error(err),
		)
		return
	}

	s.log.Info("Stopped monitoring deleted token contract",
		logger.String("token_address", token.Address),
		logger.String("token_symbol", token.Symbol),
		logger.String("chain_type", string(token.ChainType)),
	)
}
