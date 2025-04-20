package tokenstore

import (
	"context"

	"vault0/internal/db"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// TokenEventType defines the type of token event
type TokenEventType string

const (
	// TokenEventAdded is emitted when a token is added to the store
	TokenEventAdded TokenEventType = "TOKEN_ADDED"
	// TokenEventUpdated is emitted when a token is updated
	TokenEventUpdated TokenEventType = "TOKEN_UPDATED"
	// TokenEventDeleted is emitted when a token is deleted
	TokenEventDeleted TokenEventType = "TOKEN_DELETED"
)

// TokenEvent represents an event related to a token
type TokenEvent struct {
	EventType TokenEventType
	Token     *types.Token
}

// TokenStore defines the interface for managing tokens
type TokenStore interface {
	// AddToken adds a new token to the store
	AddToken(ctx context.Context, token *types.Token) error

	// GetToken retrieves a token by its address
	GetToken(ctx context.Context, address string) (*types.Token, error)

	// UpdateToken updates an existing token
	UpdateToken(ctx context.Context, token *types.Token) error

	// DeleteToken removes a token from the store by its address
	DeleteToken(ctx context.Context, address string) error

	// ListTokens retrieves tokens in the store with pagination
	// If limit is 0, returns all tokens without pagination
	// nextToken is used for token-based pagination (empty string for first page)
	ListTokens(ctx context.Context, limit int, nextToken string) (*types.Page[types.Token], error)

	// ListTokensByChain retrieves tokens for a specific blockchain with pagination
	// If limit is 0, returns all tokens without pagination
	// nextToken is used for token-based pagination (empty string for first page)
	ListTokensByChain(ctx context.Context, chainType types.ChainType, limit int, nextToken string) (*types.Page[types.Token], error)

	// ListTokensByAddresses retrieves tokens by a list of token addresses for a specific chain
	// Returns tokens in the same order as the input addresses
	// If an address is not found, it will be skipped in the result
	ListTokensByAddresses(ctx context.Context, chainType types.ChainType, addresses []string) ([]types.Token, error)

	// TokenEvents returns a channel that emits token events.
	// This channel notifies subscribers when tokens are added, updated, or deleted.
	// The channel is closed when the token store is closed or the subscription is canceled.
	TokenEvents() <-chan TokenEvent
}

// NewTokenStore creates a new TokenStore instance
func NewTokenStore(db *db.DB, log logger.Logger) TokenStore {
	const tokenEventBufferSize = 100
	return &dbTokenStore{
		db:          db,
		log:         log,
		tokenEvents: make(chan TokenEvent, tokenEventBufferSize),
	}
}
