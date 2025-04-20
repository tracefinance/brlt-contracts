package tokenstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// dbTokenStore implements the TokenStore interface using an SQL database
type dbTokenStore struct {
	db          *db.DB
	log         logger.Logger
	tokenEvents chan TokenEvent
}

// TokenEvents returns a channel that emits token events
func (s *dbTokenStore) TokenEvents() <-chan TokenEvent {
	return s.tokenEvents
}

// AddToken adds a new token to the database
func (s *dbTokenStore) AddToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return errors.NewInvalidTokenError("token is nil", nil)
	}

	if err := token.Validate(); err != nil {
		return errors.NewInvalidTokenError("validation failed", err)
	}

	// Parse and normalize address using the new Address struct
	addr, err := types.NewAddress(token.ChainType, token.Address)
	if err != nil {
		return errors.NewInvalidTokenError("invalid address", err)
	}

	// Update token with normalized address
	token.Address = addr.Address

	// Check if the token already exists
	exists, err := s.Exists(ctx, token.Address)
	if err != nil {
		return err
	}

	if exists {
		return errors.NewResourceAlreadyExistsError("token", "address", token.Address)
	}

	// Insert the new token
	_, err = s.db.ExecuteStatementContext(
		ctx,
		`INSERT INTO tokens (address, chain_type, symbol, decimals, type) 
		VALUES (?, ?, ?, ?, ?)`,
		token.Address,
		token.ChainType,
		token.Symbol,
		token.Decimals,
		token.Type,
	)

	if err != nil {
		return err
	}

	// Emit a token added event
	select {
	case s.tokenEvents <- TokenEvent{EventType: TokenEventAdded, Token: token}:
		s.log.Debug("Emitted token added event",
			logger.String("symbol", token.Symbol),
			logger.String("address", token.Address),
			logger.String("chain", string(token.ChainType)))
	default:
		// Don't block if channel buffer is full
		s.log.Warn("Token events channel full, event not emitted",
			logger.String("symbol", token.Symbol),
			logger.String("address", token.Address))
	}

	return nil
}

// GetToken retrieves a token by its address
func (s *dbTokenStore) GetToken(ctx context.Context, address string) (*types.Token, error) {
	// Just lowercase the address for querying
	lowercaseAddress := strings.ToLower(address)

	var token types.Token
	rows, err := s.db.ExecuteQueryContext(
		ctx,
		`SELECT address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE lower(address) = ?`,
		lowercaseAddress,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewResourceNotFoundError("token", address)
	}

	err = rows.Scan(
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
func (s *dbTokenStore) ListTokensByChain(ctx context.Context, chainType types.ChainType, limit int, nextToken string) (*types.Page[types.Token], error) {
	query := `SELECT address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE chain_type = ?`

	args := []any{chainType}

	// Default sort column
	paginationColumn := "symbol"

	// Handle token-based pagination
	token, err := types.DecodeNextPageToken(nextToken, paginationColumn)
	if err != nil {
		return nil, err
	}

	if token != nil {
		query += fmt.Sprintf(" AND %s > ?", paginationColumn)
		args = append(args, token.Value)
	}

	// Add consistent ordering
	query += fmt.Sprintf(" ORDER BY %s ASC", paginationColumn)

	// Add pagination if limit > 0
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit+1) // Fetch one extra item to determine if there are more results
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

	// Generate token function for pagination
	generateToken := func(token types.Token) *types.NextPageToken {
		return &types.NextPageToken{
			Column: paginationColumn,
			Value:  token.Symbol,
		}
	}

	return types.NewPage(tokenItems, limit, generateToken), nil
}

// ListTokens retrieves tokens in the store with pagination
func (s *dbTokenStore) ListTokens(ctx context.Context, limit int, nextToken string) (*types.Page[types.Token], error) {
	query := `SELECT address, chain_type, symbol, decimals, type 
		FROM tokens`

	args := []any{}

	// Default sort columns
	paginationColumn := "symbol"
	secondaryColumn := "chain_type"

	// Handle token-based pagination
	if nextToken != "" {
		token, err := types.DecodeNextPageToken(nextToken, paginationColumn)
		if err != nil {
			return nil, err
		}

		if token != nil {
			query += fmt.Sprintf(" WHERE %s > ?", paginationColumn)
			args = append(args, token.Value)
		}
	}

	// Add consistent ordering
	query += fmt.Sprintf(" ORDER BY %s ASC, %s ASC", paginationColumn, secondaryColumn)

	// Add pagination if limit > 0
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit+1) // Fetch one extra item
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

	// Generate token function for pagination
	generateToken := func(token types.Token) *types.NextPageToken {
		return &types.NextPageToken{
			Column: paginationColumn,
			Value:  token.Symbol,
		}
	}

	return types.NewPage(tokenItems, limit, generateToken), nil
}

// UpdateToken updates an existing token
func (s *dbTokenStore) UpdateToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return errors.NewInvalidTokenError("token is nil", nil)
	}

	if err := token.Validate(); err != nil {
		return errors.NewInvalidTokenError("validation failed", err)
	}

	// Parse and normalize address using the new Address struct
	addr, err := types.NewAddress(token.ChainType, token.Address)
	if err != nil {
		return errors.NewInvalidTokenError("invalid address", err)
	}

	// Update token with normalized address
	token.Address = addr.Address

	query := `UPDATE tokens 
		SET symbol = ?, decimals = ?, type = ?, updated_at = ?
		WHERE address = ? AND chain_type = ?`
	args := []any{
		token.Symbol,
		token.Decimals,
		token.Type,
		time.Now(),
		token.Address,
		token.ChainType,
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
		return errors.NewResourceNotFoundError("token", token.Address)
	}

	// Emit a token updated event
	select {
	case s.tokenEvents <- TokenEvent{EventType: TokenEventUpdated, Token: token}:
		s.log.Debug("Emitted token updated event",
			logger.String("symbol", token.Symbol),
			logger.String("address", token.Address))
	default:
		// Don't block if channel buffer is full
		s.log.Warn("Token events channel full, update event not emitted",
			logger.String("symbol", token.Symbol),
			logger.String("address", token.Address))
	}

	return nil
}

// DeleteToken removes a token from the store by its address
func (s *dbTokenStore) DeleteToken(ctx context.Context, address string) error {
	// Just lowercase the address for querying
	lowercaseAddress := strings.ToLower(address)

	// First get the token to emit in the event
	token, err := s.GetToken(ctx, address)
	if err != nil {
		return err
	}

	result, err := s.db.ExecuteStatementContext(
		ctx,
		"DELETE FROM tokens WHERE lower(address) = ?",
		lowercaseAddress,
	)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewResourceNotFoundError("token", address)
	}

	// Emit a token deleted event
	select {
	case s.tokenEvents <- TokenEvent{EventType: TokenEventDeleted, Token: token}:
		s.log.Debug("Emitted token deleted event",
			logger.String("symbol", token.Symbol),
			logger.String("address", address))
	default:
		// Don't block if channel buffer is full
		s.log.Warn("Token events channel full, delete event not emitted",
			logger.String("symbol", token.Symbol),
			logger.String("address", address))
	}

	return nil
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

// Exists checks if a token exists by its address
func (s *dbTokenStore) Exists(ctx context.Context, address string) (bool, error) {
	// Just lowercase the address for querying
	lowercaseAddress := strings.ToLower(address)

	rows, err := s.db.ExecuteQueryContext(
		ctx,
		"SELECT 1 FROM tokens WHERE lower(address) = ? LIMIT 1",
		lowercaseAddress,
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

// ListTokensByAddresses retrieves tokens by a list of token addresses for a specific chain
func (s *dbTokenStore) ListTokensByAddresses(ctx context.Context, chainType types.ChainType, addresses []string) ([]types.Token, error) {
	if len(addresses) == 0 {
		return []types.Token{}, nil
	}

	// Just lowercase all addresses for querying
	lowercaseAddresses := make([]string, len(addresses))
	for i, address := range addresses {
		lowercaseAddresses[i] = strings.ToLower(address)
	}

	// Create placeholders for the SQL IN clause
	placeholders := make([]string, len(lowercaseAddresses))
	args := make([]any, len(lowercaseAddresses)+1) // +1 for chainType
	args[0] = chainType

	for i, address := range lowercaseAddresses {
		placeholders[i] = "?"
		args[i+1] = address
	}

	// Join the placeholders with commas
	placeholdersStr := strings.Join(placeholders, ",")

	// Build the query using the placeholders
	query := `SELECT address, chain_type, symbol, decimals, type 
		FROM tokens 
		WHERE chain_type = ? AND lower(address) IN (` + placeholdersStr + `)`

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

	// Convert []*types.Token to []types.Token
	tokenItems := make([]types.Token, 0, len(tokens))
	for _, token := range tokens {
		tokenItems = append(tokenItems, *token)
	}

	return tokenItems, nil
}
