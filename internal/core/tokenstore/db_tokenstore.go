package tokenstore

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// dbTokenStore implements the TokenStore interface using an SQL database
type dbTokenStore struct {
	db  *db.DB
	log logger.Logger
}

// AddToken adds a new token to the database
func (s *dbTokenStore) AddToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return errors.NewInvalidTokenError("token is nil", nil)
	}

	if err := token.Validate(); err != nil {
		return errors.NewInvalidTokenError("validation failed", err)
	}

	// Check if the token already exists
	exists, err := s.Exists(ctx, token.Address, token.ChainType)
	if err != nil {
		return err
	}

	if exists {
		return errors.NewResourceAlreadyExistsError("token", "address", token.Address)
	}

	// Generate a new Snowflake ID for the token
	tokenID, err := s.db.GenerateID()
	if err != nil {
		return err
	}

	// Insert the new token
	_, err = s.db.ExecuteStatementContext(
		ctx,
		`INSERT INTO tokens (id, address, chain_type, symbol, decimals, type) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		tokenID,
		token.Address,
		token.ChainType,
		token.Symbol,
		token.Decimals,
		token.Type,
	)

	if err != nil {
		return err
	}

	// Set the ID in the token struct
	token.ID = tokenID

	return nil
}

// GetToken retrieves a token by its address and chain type
func (s *dbTokenStore) GetToken(ctx context.Context, address string, chainType types.ChainType) (*types.Token, error) {
	normalizedAddress := types.NormalizeAddress(address)

	var token types.Token
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT id, address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE lower(address) = ? AND chain_type = ?`,
		normalizedAddress,
		chainType,
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &token, nil
}

// GetTokenByID retrieves a token by its primary ID
func (s *dbTokenStore) GetTokenByID(ctx context.Context, id int64) (*types.Token, error) {
	var token types.Token
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT id, address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE id = ?`,
		id,
	)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return &token, nil
}

// ListTokensByChain retrieves tokens for a specific blockchain with pagination
func (s *dbTokenStore) ListTokensByChain(ctx context.Context, chainType types.ChainType, offset, limit int) (*types.Page[types.Token], error) {
	query := `SELECT id, address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE chain_type = ?
		ORDER BY symbol`

	args := []any{chainType}

	// Add pagination if limit > 0
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := s.db.ExecuteQueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tokens, err := s.scanTokensFromRows(rows)
	if err != nil {
		return nil, err
	}

	// Convert []*types.Token to []types.Token for the Page
	tokenItems := make([]types.Token, 0, len(tokens))
	for _, token := range tokens {
		tokenItems = append(tokenItems, *token)
	}

	return types.NewPage(tokenItems, offset, limit), nil
}

// ListTokens retrieves tokens in the store with pagination
func (s *dbTokenStore) ListTokens(ctx context.Context, offset, limit int) (*types.Page[types.Token], error) {
	query := `SELECT id, address, chain_type, symbol, decimals, type 
		FROM tokens 
		ORDER BY chain_type, symbol`

	args := []any{}

	// Add pagination if limit > 0
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := s.db.ExecuteQueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	tokens, err := s.scanTokensFromRows(rows)
	if err != nil {
		return nil, err
	}

	// Convert []*types.Token to []types.Token for the Page
	tokenItems := make([]types.Token, 0, len(tokens))
	for _, token := range tokens {
		tokenItems = append(tokenItems, *token)
	}

	return types.NewPage(tokenItems, offset, limit), nil
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

// Exists checks if a token exists by its address and chain type
func (s *dbTokenStore) Exists(ctx context.Context, address string, chainType types.ChainType) (bool, error) {
	normalizedAddress := types.NormalizeAddress(address)

	rows, err := s.db.ExecuteQueryContext(
		ctx,
		"SELECT 1 FROM tokens WHERE lower(address) = ? AND chain_type = ? LIMIT 1",
		normalizedAddress,
		chainType,
	)
	if err != nil {
		return false, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	exists := rows.Next()
	if err = rows.Err(); err != nil {
		return false, errors.NewDatabaseError(err)
	}

	return exists, nil
}

// ListTokensByIDs retrieves tokens by a list of token IDs with pagination
func (s *dbTokenStore) ListTokensByIDs(ctx context.Context, ids []int64, offset, limit int) (*types.Page[types.Token], error) {
	if len(ids) == 0 {
		return types.NewPage([]types.Token{}, offset, limit), nil
	}

	// Create placeholders for the SQL IN clause
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	// Join the placeholders with commas using strings.Join
	placeholdersStr := strings.Join(placeholders, ",")

	// Build the query using the placeholders
	query := `SELECT id, address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE id IN (` + placeholdersStr + `)`

	// Add pagination if limit > 0
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := s.db.ExecuteQueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	// Scan tokens from rows
	tokens, err := s.scanTokensFromRows(rows)
	if err != nil {
		return nil, err
	}

	// Create a map of token IDs to tokens for quick lookup
	tokenMap := make(map[int64]*types.Token, len(tokens))
	for _, token := range tokens {
		tokenMap[token.ID] = token
	}

	// Create result slice with tokens in the same order as the input IDs
	// Skip IDs that don't have corresponding tokens
	// Apply pagination manually since SQL limit/offset doesn't respect the original order
	resultTokens := make([]types.Token, 0, len(tokens))

	// If limit is 0, process all IDs
	if limit <= 0 {
		for _, id := range ids {
			if token, exists := tokenMap[id]; exists {
				resultTokens = append(resultTokens, *token)
			}
		}
	} else {
		// Apply manual pagination to maintain order of input IDs
		// Skip tokens before offset
		startIndex := offset
		endIndex := offset + limit

		// Ensure we don't go out of bounds
		if startIndex >= len(ids) {
			startIndex = len(ids)
		}
		if endIndex > len(ids) {
			endIndex = len(ids)
		}

		// Process only the IDs within the paginated range
		for i := startIndex; i < endIndex; i++ {
			if token, exists := tokenMap[ids[i]]; exists {
				resultTokens = append(resultTokens, *token)
			}
		}
	}

	return types.NewPage(resultTokens, offset, limit), nil
}

// GetNativeToken retrieves the native token for a blockchain
func (s *dbTokenStore) GetNativeToken(ctx context.Context, chainType types.ChainType) (*types.Token, error) {
	// First check if the native token exists in the database
	var token types.Token
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT id, address, chain_type, symbol, decimals, type
		FROM tokens
		WHERE chain_type = ? AND type = ?`,
		chainType,
		types.TokenTypeNative,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(
			&token.ID,
			&token.Address,
			&token.ChainType,
			&token.Symbol,
			&token.Decimals,
			&token.Type,
		)
		if err != nil {
			return nil, err
		}
		return &token, nil
	}

	// If native token not found, return error
	return nil, errors.NewInvalidTokenError(
		fmt.Sprintf("native token for chain %s not found", chainType),
		nil,
	)
}
