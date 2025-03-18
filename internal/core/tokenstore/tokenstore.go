package tokenstore

import (
	"context"

	"vault0/internal/db"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// TokenStore defines the interface for managing tokens
type TokenStore interface {
	// AddToken adds a new token to the store
	AddToken(ctx context.Context, token *types.Token) error

	// GetToken retrieves a token by its identifier (address and chain type)
	GetToken(ctx context.Context, address string, chainType types.ChainType) (*types.Token, error)

	// GetTokenByID retrieves a token by its ID
	GetTokenByID(ctx context.Context, id int64) (*types.Token, error)

	// UpdateToken updates an existing token
	UpdateToken(ctx context.Context, token *types.Token) error

	// DeleteToken removes a token from the store by its ID
	DeleteToken(ctx context.Context, id int64) error

	// ListTokens retrieves tokens in the store with pagination
	// If limit is 0, returns all tokens without pagination
	ListTokens(ctx context.Context, offset, limit int) (*types.Page[types.Token], error)

	// ListTokensByChain retrieves tokens for a specific blockchain with pagination
	// If limit is 0, returns all tokens without pagination
	ListTokensByChain(ctx context.Context, chainType types.ChainType, offset, limit int) (*types.Page[types.Token], error)
}

// NewTokenStore creates a new TokenStore instance
func NewTokenStore(db *db.DB, log logger.Logger) TokenStore {
	return &dbTokenStore{db: db, log: log}
}
