package token

import (
	"context"

	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Service defines the token service interface
type Service interface {
	// ListTokens retrieves tokens with optional filtering
	// nextToken is used for token-based pagination (empty string for first page)
	ListTokens(ctx context.Context, filter *tokenstore.TokenFilter, limit int, nextToken string) (*types.Page[types.Token], error)

	// AddToken adds a new token
	AddToken(ctx context.Context, token *types.Token) error

	// DeleteToken removes a token by address
	DeleteToken(ctx context.Context, address string) error

	// VerifyToken checks if a token exists by address
	VerifyToken(ctx context.Context, address string) (*types.Token, error)

	// GetToken retrieves a token by address
	GetToken(ctx context.Context, address string) (*types.Token, error)

	// GetTokenByChainAndAddress retrieves a token by chain type and address
	// If address is "native" or zero address, it returns the native token for the chain
	GetTokenByChainAndAddress(ctx context.Context, chainType types.ChainType, address string) (*types.Token, error)

	// GetTokensByAddresses retrieves tokens by a list of token addresses for a specific chain
	// If an address is not found, it will be skipped in the result
	GetTokensByAddresses(ctx context.Context, chainType types.ChainType, addresses []string) ([]types.Token, error)

	// UpdateToken updates a token's symbol, type, and decimals by address
	UpdateToken(ctx context.Context, address string, symbol string, tokenType types.TokenType, decimals uint8) error
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
func (s *service) ListTokens(ctx context.Context, filter *tokenstore.TokenFilter, limit int, nextToken string) (*types.Page[types.Token], error) {
	// Set default limit
	if limit <= 0 {
		limit = 10
	}

	// Create a TokenStore filter from our Service filter
	storeFilter := tokenstore.TokenFilter{
		ChainType: filter.ChainType,
		TokenType: filter.TokenType,
	}

	// Use the updated ListTokens method with filter support
	tokens, err := s.tokenStore.ListTokens(ctx, &storeFilter, limit, nextToken)
	if err != nil {
		// Prepare chainType and tokenType strings for logging
		chainTypeStr := "all"
		if filter.ChainType != nil {
			chainTypeStr = string(*filter.ChainType)
		}

		tokenTypeStr := "all"
		if filter.TokenType != nil {
			tokenTypeStr = string(*filter.TokenType)
		}

		s.log.Error("Failed to list tokens",
			logger.Error(err),
			logger.String("chain_type", chainTypeStr),
			logger.String("token_type", tokenTypeStr))
		return nil, err
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

// GetToken implements the Service interface
func (s *service) GetToken(ctx context.Context, address string) (*types.Token, error) {
	token, err := s.tokenStore.GetToken(ctx, address)
	if err != nil {
		s.log.Error("Failed to get token by address", logger.Error(err), logger.String("address", address))
		return nil, err
	}

	return token, nil
}

// DeleteToken implements the Service interface
func (s *service) DeleteToken(ctx context.Context, address string) error {
	// Check if the token exists first
	token, err := s.tokenStore.GetToken(ctx, address)
	if err != nil {
		s.log.Error("Failed to get token for deletion", logger.Error(err), logger.String("address", address))
		return err
	}

	// Delete the token
	err = s.tokenStore.DeleteToken(ctx, address)
	if err != nil {
		s.log.Error("Failed to delete token", logger.Error(err), logger.String("address", address))
		return err
	}

	s.log.Info("Token deleted successfully",
		logger.String("address", address),
		logger.String("symbol", token.Symbol),
		logger.String("chain_type", string(token.ChainType)))

	return nil
}

// VerifyToken implements the Service interface
func (s *service) VerifyToken(ctx context.Context, address string) (*types.Token, error) {
	token, err := s.tokenStore.GetToken(ctx, address)
	if err != nil {
		s.log.Error("Token verification failed", logger.Error(err), logger.String("address", address))
		return nil, err
	}

	return token, nil
}

// GetTokenByChainAndAddress implements the Service interface
func (s *service) GetTokenByChainAndAddress(ctx context.Context, chainType types.ChainType, address string) (*types.Token, error) {
	// Check if address is "native" or zero address
	if address == "native" || types.IsZeroAddress(address) {
		// Create a native token for the specified chain
		nativeToken, err := types.NewNativeToken(chainType)
		if err != nil {
			s.log.Error("Failed to create native token",
				logger.Error(err),
				logger.String("chain_type", string(chainType)))
			return nil, err
		}
		return nativeToken, nil
	}

	// Get token by address from the token store
	token, err := s.tokenStore.GetToken(ctx, address)
	if err != nil {
		s.log.Error("Failed to get token by address and chain",
			logger.Error(err),
			logger.String("address", address),
			logger.String("chain_type", string(chainType)))
		return nil, err
	}

	// Verify the chain type matches
	if token.ChainType != chainType {
		s.log.Error("Token chain type mismatch",
			logger.String("requested_chain", string(chainType)),
			logger.String("token_chain", string(token.ChainType)),
			logger.String("address", address))
		return nil, errors.NewTokenNotFoundError(address, string(chainType))
	}

	return token, nil
}

// GetTokensByAddresses implements the Service interface
func (s *service) GetTokensByAddresses(ctx context.Context, chainType types.ChainType, addresses []string) ([]types.Token, error) {
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

// UpdateToken implements the Service interface
func (s *service) UpdateToken(ctx context.Context, address string, symbol string, tokenType types.TokenType, decimals uint8) error {
	// Check if the token exists first
	existingToken, err := s.tokenStore.GetToken(ctx, address)
	if err != nil {
		s.log.Error("Failed to get token for update",
			logger.Error(err),
			logger.String("address", address))
		return err
	}

	// Update the token fields
	existingToken.Symbol = symbol
	existingToken.Type = tokenType
	existingToken.Decimals = decimals

	// Validate the updated token
	if err := existingToken.Validate(); err != nil {
		s.log.Error("Updated token validation failed",
			logger.Error(err),
			logger.String("symbol", symbol),
			logger.String("address", address),
			logger.String("token_type", string(tokenType)))
		return err
	}

	// Update the token in the store
	err = s.tokenStore.UpdateToken(ctx, existingToken)
	if err != nil {
		s.log.Error("Failed to update token",
			logger.Error(err),
			logger.String("address", address),
			logger.String("symbol", symbol),
			logger.String("token_type", string(tokenType)))
		return err
	}

	s.log.Info("Token updated successfully",
		logger.String("address", address),
		logger.String("symbol", symbol),
		logger.String("token_type", string(tokenType)))

	return nil
}
