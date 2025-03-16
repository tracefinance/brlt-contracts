package tokenstore

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"vault0/internal/core/db"
	"vault0/internal/types"
)

var (
	// ErrTokenNotFound is returned when a token is not found
	ErrTokenNotFound = errors.New("token not found")

	// ErrTokenAlreadyExists is returned when trying to add a token that already exists
	ErrTokenAlreadyExists = errors.New("token already exists")

	// ErrInvalidToken is returned when a token is invalid
	ErrInvalidToken = errors.New("invalid token")
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

// NormalizeAddress ensures consistent address format for storage and comparison
func NormalizeAddress(address string) string {
	// Convert to lowercase for case-insensitive comparisons
	address = strings.ToLower(address)

	// Ensure the address has 0x prefix for EVM addresses
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}

	return address
}

// IsZeroAddress checks if the address is the zero address
func IsZeroAddress(address string) bool {
	normalized := NormalizeAddress(address)
	return normalized == types.ZeroAddress || normalized == "0x0"
}

// ValidateTokenData performs basic validation of token data
func ValidateTokenData(token *types.Token) error {
	if token == nil {
		return fmt.Errorf("%w: token is nil", ErrInvalidToken)
	}

	if err := token.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidToken, err)
	}

	return nil
}
