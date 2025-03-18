package tokenstore

import (
	"context"

	"vault0/internal/db"
	"vault0/internal/types"
)

// TokenStore defines the interface for managing tokens
type TokenStore interface {
	// AddToken adds a new token to the store
	AddToken(ctx context.Context, token *types.Token) error

	// GetTokenByAddress retrieves a token by its identifier (address and chain type)
	GetTokenByAddress(ctx context.Context, address string, chainType types.ChainType) (*types.Token, error)

	// GetToken retrieves a token by its ID
	GetToken(ctx context.Context, id int64) (*types.Token, error)

	// UpdateToken updates an existing token
	UpdateToken(ctx context.Context, token *types.Token) error

	// DeleteToken removes a token from the store by its ID
	DeleteToken(ctx context.Context, id int64) error

	// ListTokens retrieves all tokens in the store
	ListTokens(ctx context.Context) ([]*types.Token, error)

	// ListTokensByChain retrieves all tokens for a specific blockchain
	ListTokensByChain(ctx context.Context, chainType types.ChainType) ([]*types.Token, error)
}

// NewTokenStore creates a new TokenStore instance
func NewTokenStore(db *db.DB) TokenStore {
	return &dbTokenStore{db: db}
}
