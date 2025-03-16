package tokenstore

import (
	"context"

	"vault0/internal/core/db"
	"vault0/internal/types"
)

// TokenStore defines the interface for managing tokens
type TokenStore interface {
	// AddToken adds a new token to the store
	AddToken(ctx context.Context, token *types.Token) error

	// GetToken retrieves a token by its identifier (address and chain type)
	GetToken(ctx context.Context, address string, chainType types.ChainType) (*types.Token, error)

	// GetTokenByID retrieves a token by its composite ID (address:chainType)
	GetTokenByID(ctx context.Context, id string) (*types.Token, error)

	// GetTokensByChain retrieves all tokens for a specific blockchain
	GetTokensByChain(ctx context.Context, chainType types.ChainType) ([]*types.Token, error)

	// GetTokensByType retrieves all tokens of a specific type (native, ERC20)
	GetTokensByType(ctx context.Context, tokenType types.TokenType) ([]*types.Token, error)

	// UpdateToken updates an existing token
	UpdateToken(ctx context.Context, token *types.Token) error

	// DeleteToken removes a token from the store
	DeleteToken(ctx context.Context, address string, chainType types.ChainType) error

	// ListAllTokens retrieves all tokens in the store
	ListAllTokens(ctx context.Context) ([]*types.Token, error)
}

// NewTokenStore creates a new TokenStore instance
func NewTokenStore(db *db.DB) TokenStore {
	return &dbTokenStore{
		db: db.GetConnection(),
	}
}
