package tokenstore

import (
	"context"
	"database/sql"
	"strconv"
	"time"

	"vault0/internal/core/db"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// dbTokenStore implements the TokenStore interface using an SQL database
type dbTokenStore struct {
	db *db.DB
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
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		"SELECT 1 FROM tokens WHERE address = ? AND chain_type = ? LIMIT 1",
		normalizedAddress,
		token.ChainType,
	)
	if err != nil {
		return errors.NewDatabaseError(err)
	}
	defer rows.Close()

	exists = rows.Next()
	if err = rows.Err(); err != nil {
		return errors.NewDatabaseError(err)
	}

	if exists {
		return errors.NewResourceAlreadyExistsError("token", "address", normalizedAddress)
	}

	// Generate a new Snowflake ID for the token
	tokenID, err := s.db.GenerateID()
	if err != nil {
		return errors.NewOperationFailedError("generate token id", err)
	}

	// Insert the new token
	_, err = s.db.ExecuteStatementContext(
		ctx,
		`INSERT INTO tokens (id, address, chain_type, symbol, decimals, type) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		tokenID,
		normalizedAddress,
		token.ChainType,
		token.Symbol,
		token.Decimals,
		token.Type,
	)

	if err != nil {
		return errors.NewDatabaseError(err)
	}

	// Set the ID in the token struct
	token.ID = tokenID

	return nil
}

// GetTokenByAddress retrieves a token by its address and chain type
func (s *dbTokenStore) GetTokenByAddress(ctx context.Context, address string, chainType types.ChainType) (*types.Token, error) {
	normalizedAddress := types.NormalizeAddress(address)

	var token types.Token
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT id, address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE address = ? AND chain_type = ?`,
		normalizedAddress,
		chainType,
	)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewResourceNotFoundError("token", normalizedAddress)
	}

	err = rows.Scan(
		&token.ID,
		&token.Address,
		&token.ChainType,
		&token.Symbol,
		&token.Decimals,
		&token.Type,
	)

	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return &token, nil
}

// GetToken retrieves a token by its primary ID
func (s *dbTokenStore) GetToken(ctx context.Context, id int64) (*types.Token, error) {
	var token types.Token
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT id, address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE id = ?`,
		id,
	)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewResourceNotFoundError("token", strconv.FormatInt(id, 10))
	}

	err = rows.Scan(
		&token.ID,
		&token.Address,
		&token.ChainType,
		&token.Symbol,
		&token.Decimals,
		&token.Type,
	)

	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return &token, nil
}

// ListTokensByChain retrieves all tokens for a specific blockchain
func (s *dbTokenStore) ListTokensByChain(ctx context.Context, chainType types.ChainType) ([]*types.Token, error) {
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT id, address, chain_type, symbol, decimals, type 
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

// UpdateToken updates an existing token
func (s *dbTokenStore) UpdateToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return errors.NewInvalidTokenError("token is nil", nil)
	}

	if err := token.Validate(); err != nil {
		return errors.NewInvalidTokenError("validation failed", err)
	}

	if token.ID == 0 {
		return errors.NewInvalidTokenError("token ID is required for update", nil)
	}

	query := `UPDATE tokens 
		SET symbol = ?, decimals = ?, type = ?, updated_at = ?
		WHERE id = ?`
	args := []any{
		token.Symbol,
		token.Decimals,
		token.Type,
		time.Now(),
		token.ID,
	}

	result, err := s.db.ExecuteStatementContext(ctx, query, args...)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewResourceNotFoundError("token", strconv.FormatInt(token.ID, 10))
	}

	return nil
}

// DeleteToken removes a token from the store by its ID
func (s *dbTokenStore) DeleteToken(ctx context.Context, id int64) error {
	result, err := s.db.ExecuteStatementContext(
		ctx,
		"DELETE FROM tokens WHERE id = ?",
		id,
	)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewResourceNotFoundError("token", strconv.FormatInt(id, 10))
	}

	return nil
}

// ListTokens retrieves all tokens in the store
func (s *dbTokenStore) ListTokens(ctx context.Context) ([]*types.Token, error) {
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT id, address, chain_type, symbol, decimals, type 
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
			&token.ID,
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
