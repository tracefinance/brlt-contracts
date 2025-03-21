package token

import (
	"context"

	"vault0/internal/core/tokenstore"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Service defines the token service interface
type Service interface {
	// ListTokens retrieves tokens with optional filtering
	ListTokens(ctx context.Context, filter TokenFilter, offset, limit int) (*types.Page[types.Token], error)

	// AddToken adds a new token
	AddToken(ctx context.Context, token *types.Token) error

	// GetTokenByID retrieves a token by ID
	GetTokenByID(ctx context.Context, id int64) (*types.Token, error)

	// DeleteToken removes a token by ID
	DeleteToken(ctx context.Context, id int64) error

	// VerifyToken checks if a token exists by ID
	VerifyToken(ctx context.Context, id int64) (*types.Token, error)

	// GetToken retrieves a token by address
	GetToken(ctx context.Context, chainType types.ChainType, address string) (*types.Token, error)

	// ListTokensByID retrieves tokens by a list of token IDs
	ListTokensByID(ctx context.Context, ids []int64) ([]types.Token, error)

	// ListTokensByAddresses retrieves tokens by a list of token addresses for a specific chain
	// If an address is not found, it will be skipped in the result
	ListTokensByAddresses(ctx context.Context, chainType types.ChainType, addresses []string) ([]types.Token, error)
}

// service implements the Service interface
type service struct {
	tokenStore tokenstore.TokenStore
	log        logger.Logger
}

// NewService creates a new token service instance
func NewService(tokenStore tokenstore.TokenStore, log logger.Logger) Service {
	return &service{
		tokenStore: tokenStore,
		log:        log,
	}
}

// ListTokens implements the Service interface
func (s *service) ListTokens(ctx context.Context, filter TokenFilter, offset, limit int) (*types.Page[types.Token], error) {
	// Set default values for offset and limit
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 10
	}
	// If chain type filter is provided, use the specific list method
	if filter.ChainType != nil {
		tokens, err := s.tokenStore.ListTokensByChain(ctx, *filter.ChainType, offset, limit)
		if err != nil {
			s.log.Error("Failed to list tokens by chain",
				logger.Error(err),
				logger.String("chain_type", string(*filter.ChainType)))
			return nil, err
		}

		// If token type filter is also provided, filter the results
		if filter.TokenType != nil {
			filtered := make([]types.Token, 0)
			for _, token := range tokens.Items {
				if token.Type == *filter.TokenType {
					filtered = append(filtered, token)
				}
			}

			// Return a new page with filtered items
			return types.NewPage(filtered, offset, limit), nil
		}

		return tokens, nil
	}

	// Otherwise, list all tokens
	tokens, err := s.tokenStore.ListTokens(ctx, offset, limit)
	if err != nil {
		s.log.Error("Failed to list tokens", logger.Error(err))
		return nil, err
	}

	// If token type filter is provided, filter the results
	if filter.TokenType != nil {
		filtered := make([]types.Token, 0)
		for _, token := range tokens.Items {
			if token.Type == *filter.TokenType {
				filtered = append(filtered, token)
			}
		}

		// Return a new page with filtered items
		return types.NewPage(filtered, offset, limit), nil
	}

	return tokens, nil
}

// AddToken implements the Service interface
func (s *service) AddToken(ctx context.Context, token *types.Token) error {
	// Validate the token
	if err := token.Validate(); err != nil {
		s.log.Error("Token validation failed",
			logger.Error(err),
			logger.String("symbol", token.Symbol),
			logger.String("address", token.Address),
			logger.String("chain_type", string(token.ChainType)))
		return err
	}

	// Add the token to the store
	err := s.tokenStore.AddToken(ctx, token)
	if err != nil {
		s.log.Error("Failed to add token",
			logger.Error(err),
			logger.String("symbol", token.Symbol),
			logger.String("address", token.Address),
			logger.String("chain_type", string(token.ChainType)))
		return err
	}

	s.log.Info("Token added successfully",
		logger.String("symbol", token.Symbol),
		logger.String("address", token.Address),
		logger.String("chain_type", string(token.ChainType)))

	return nil
}

// GetTokenByID implements the Service interface
func (s *service) GetTokenByID(ctx context.Context, id int64) (*types.Token, error) {
	token, err := s.tokenStore.GetTokenByID(ctx, id)
	if err != nil {
		s.log.Error("Failed to get token", logger.Error(err), logger.Int("token_id", int(id)))
		return nil, err
	}

	return token, nil
}

// GetToken implements the Service interface
func (s *service) GetToken(ctx context.Context, chainType types.ChainType, address string) (*types.Token, error) {
	token, err := s.tokenStore.GetToken(ctx, address, chainType)
	if err != nil {
		s.log.Error("Failed to get token by address", logger.Error(err), logger.String("address", address))
		return nil, err
	}

	return token, nil
}

// DeleteToken implements the Service interface
func (s *service) DeleteToken(ctx context.Context, id int64) error {
	// Check if the token exists first
	token, err := s.tokenStore.GetTokenByID(ctx, id)
	if err != nil {
		s.log.Error("Failed to get token for deletion", logger.Error(err), logger.Int("token_id", int(id)))
		return err
	}

	// Delete the token
	err = s.tokenStore.DeleteToken(ctx, id)
	if err != nil {
		s.log.Error("Failed to delete token", logger.Error(err), logger.Int("token_id", int(id)))
		return err
	}

	s.log.Info("Token deleted successfully",
		logger.Int("token_id", int(id)),
		logger.String("symbol", token.Symbol),
		logger.String("chain_type", string(token.ChainType)))

	return nil
}

// VerifyToken implements the Service interface
func (s *service) VerifyToken(ctx context.Context, id int64) (*types.Token, error) {
	token, err := s.tokenStore.GetTokenByID(ctx, id)
	if err != nil {
		s.log.Error("Token verification failed", logger.Error(err), logger.Int("token_id", int(id)))
		return nil, err
	}

	return token, nil
}

// ListTokensByID implements the Service interface
func (s *service) ListTokensByID(ctx context.Context, ids []int64) ([]types.Token, error) {
	tokens, err := s.tokenStore.ListTokensByIDs(ctx, ids)
	if err != nil {
		s.log.Error("Failed to list tokens by IDs",
			logger.Error(err),
			logger.Any("token_ids", ids))
		return nil, err
	}

	return tokens, nil
}

// ListTokensByAddresses implements the Service interface
func (s *service) ListTokensByAddresses(ctx context.Context, chainType types.ChainType, addresses []string) ([]types.Token, error) {
	tokens, err := s.tokenStore.ListTokensByAddresses(ctx, chainType, addresses)
	if err != nil {
		s.log.Error("Failed to list tokens by addresses",
			logger.Error(err),
			logger.String("chain_type", string(chainType)),
			logger.Any("addresses", addresses))
		return nil, err
	}

	return tokens, nil
}
