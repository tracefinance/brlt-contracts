package tokenstore

import (
	"context"
	"database/sql"
	"time"

	"vault0/internal/errors"
	"vault0/internal/types"
)

// dbTokenStore implements the TokenStore interface using an SQL database
type dbTokenStore struct {
	db *sql.DB
}

// AddToken adds a new token to the database
func (s *dbTokenStore) AddToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return errors.NewInvalidTokenError("token is nil", nil)
	}

	if err := token.Validate(); err != nil {
		return errors.NewInvalidTokenError("validation failed", err)
	}

	// Normalize address for consistent storage
	normalizedAddress := types.NormalizeAddress(token.Address)

	// Check if the token already exists
	var exists bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT 1 FROM tokens WHERE address = ? AND chain_type = ? LIMIT 1",
		normalizedAddress,
		token.ChainType,
	).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		return errors.NewDatabaseError(err)
	}

	if err == nil {
		return errors.NewResourceAlreadyExistsError("token", "address", normalizedAddress)
	}

	// Insert the new token
	_, err = s.db.ExecContext(
		ctx,
		`INSERT INTO tokens (address, chain_type, symbol, decimals, type) 
		VALUES (?, ?, ?, ?, ?)`,
		normalizedAddress,
		token.ChainType,
		token.Symbol,
		token.Decimals,
		token.Type,
	)

	if err != nil {
		return errors.NewDatabaseError(err)
	}

	return nil
}

// GetToken retrieves a token by its address and chain type
func (s *dbTokenStore) GetToken(ctx context.Context, address string, chainType types.ChainType) (*types.Token, error) {
	normalizedAddress := types.NormalizeAddress(address)

	var token types.Token
	err := s.db.QueryRowContext(
		ctx,
		`SELECT address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE address = ? AND chain_type = ?`,
		normalizedAddress,
		chainType,
	).Scan(
		&token.Address,
		&token.ChainType,
		&token.Symbol,
		&token.Decimals,
		&token.Type,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewResourceNotFoundError("token", normalizedAddress)
		}
		return nil, errors.NewDatabaseError(err)
	}

	return &token, nil
}

// GetTokenByID retrieves a token by its composite ID (address:chainType)
func (s *dbTokenStore) GetTokenByID(ctx context.Context, id string) (*types.Token, error) {
	address, chainType, err := types.ParseTokenID(id)
	if err != nil {
		return nil, errors.NewInvalidTokenError("invalid token ID format", err)
	}

	return s.GetToken(ctx, address, chainType)
}

// GetTokensByChain retrieves all tokens for a specific blockchain
func (s *dbTokenStore) GetTokensByChain(ctx context.Context, chainType types.ChainType) ([]*types.Token, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE chain_type = ?
		ORDER BY symbol`,
		chainType,
	)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	return s.scanTokensFromRows(rows)
}

// GetTokensByType retrieves all tokens of a specific type (native, ERC20)
func (s *dbTokenStore) GetTokensByType(ctx context.Context, tokenType types.TokenType) ([]*types.Token, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE type = ?
		ORDER BY chain_type, symbol`,
		tokenType,
	)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	return s.scanTokensFromRows(rows)
}

// UpdateToken updates an existing token
func (s *dbTokenStore) UpdateToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return errors.NewInvalidTokenError("token is nil", nil)
	}

	if err := token.Validate(); err != nil {
		return errors.NewInvalidTokenError("validation failed", err)
	}

	normalizedAddress := types.NormalizeAddress(token.Address)

	result, err := s.db.ExecContext(
		ctx,
		`UPDATE tokens 
		SET symbol = ?, decimals = ?, type = ?, updated_at = ?
		WHERE address = ? AND chain_type = ?`,
		token.Symbol,
		token.Decimals,
		token.Type,
		time.Now(),
		normalizedAddress,
		token.ChainType,
	)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewResourceNotFoundError("token", normalizedAddress)
	}

	return nil
}

// DeleteToken removes a token from the store
func (s *dbTokenStore) DeleteToken(ctx context.Context, address string, chainType types.ChainType) error {
	normalizedAddress := types.NormalizeAddress(address)

	result, err := s.db.ExecContext(
		ctx,
		"DELETE FROM tokens WHERE address = ? AND chain_type = ?",
		normalizedAddress,
		chainType,
	)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewResourceNotFoundError("token", normalizedAddress)
	}

	return nil
}

// ListAllTokens retrieves all tokens in the store
func (s *dbTokenStore) ListAllTokens(ctx context.Context) ([]*types.Token, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT address, chain_type, symbol, decimals, type 
		FROM tokens 
		ORDER BY chain_type, symbol`,
	)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	return s.scanTokensFromRows(rows)
}

// scanTokensFromRows is a helper function to scan tokens from sql.Rows
func (s *dbTokenStore) scanTokensFromRows(rows *sql.Rows) ([]*types.Token, error) {
	var tokens []*types.Token

	for rows.Next() {
		var token types.Token
		if err := rows.Scan(
			&token.Address,
			&token.ChainType,
			&token.Symbol,
			&token.Decimals,
			&token.Type,
		); err != nil {
			return nil, errors.NewDatabaseError(err)
		}
		tokens = append(tokens, &token)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return tokens, nil
}
