package tokenstore

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"vault0/internal/types"
)

// dbTokenStore implements the TokenStore interface using an SQL database
type dbTokenStore struct {
	db *sql.DB
}

// AddToken adds a new token to the database
func (s *dbTokenStore) AddToken(ctx context.Context, token *types.Token) error {
	if err := ValidateTokenData(token); err != nil {
		return err
	}

	// Normalize address for consistent storage
	normalizedAddress := NormalizeAddress(token.Address)

	// Check if the token already exists
	var exists bool
	err := s.db.QueryRowContext(
		ctx,
		"SELECT 1 FROM tokens WHERE address = ? AND chain_type = ? LIMIT 1",
		normalizedAddress,
		token.ChainType,
	).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check if token exists: %w", err)
	}

	if err == nil {
		return ErrTokenAlreadyExists
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
		return fmt.Errorf("failed to insert token: %w", err)
	}

	return nil
}

// GetToken retrieves a token by its address and chain type
func (s *dbTokenStore) GetToken(ctx context.Context, address string, chainType types.ChainType) (*types.Token, error) {
	normalizedAddress := NormalizeAddress(address)

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
			return nil, ErrTokenNotFound
		}
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	return &token, nil
}

// GetTokenByID retrieves a token by its composite ID (address:chainType)
func (s *dbTokenStore) GetTokenByID(ctx context.Context, id string) (*types.Token, error) {
	address, chainType, err := types.ParseTokenID(id)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("failed to query tokens by chain: %w", err)
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
		return nil, fmt.Errorf("failed to query tokens by type: %w", err)
	}
	defer rows.Close()

	return s.scanTokensFromRows(rows)
}

// UpdateToken updates an existing token
func (s *dbTokenStore) UpdateToken(ctx context.Context, token *types.Token) error {
	if err := ValidateTokenData(token); err != nil {
		return err
	}

	normalizedAddress := NormalizeAddress(token.Address)

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
		return fmt.Errorf("failed to update token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTokenNotFound
	}

	return nil
}

// DeleteToken removes a token from the store
func (s *dbTokenStore) DeleteToken(ctx context.Context, address string, chainType types.ChainType) error {
	normalizedAddress := NormalizeAddress(address)

	result, err := s.db.ExecContext(
		ctx,
		"DELETE FROM tokens WHERE address = ? AND chain_type = ?",
		normalizedAddress,
		chainType,
	)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrTokenNotFound
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
		return nil, fmt.Errorf("failed to list all tokens: %w", err)
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
			return nil, fmt.Errorf("failed to scan token: %w", err)
		}
		tokens = append(tokens, &token)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating tokens rows: %w", err)
	}

	return tokens, nil
}

// ImportTokensFromConfig imports tokens from a string map representation
// This is used for importing tokens from a configuration file
func (s *dbTokenStore) ImportTokensFromConfig(ctx context.Context, configTokens map[string][]map[string]interface{}) error {
	for chainTypeStr, tokens := range configTokens {
		for _, tokenData := range tokens {
			token, err := mapToToken(chainTypeStr, tokenData)
			if err != nil {
				return fmt.Errorf("failed to parse token data: %w", err)
			}

			// We use AddToken which handles validation and duplicate checking
			if err := s.AddToken(ctx, token); err != nil {
				// Skip already existing tokens
				if err == ErrTokenAlreadyExists {
					continue
				}
				return fmt.Errorf("failed to import token %s on %s: %w", token.Symbol, token.ChainType, err)
			}
		}
	}

	return nil
}

// mapToToken converts a map representation of a token to a types.Token
func mapToToken(chainTypeStr string, data map[string]interface{}) (*types.Token, error) {
	var token types.Token

	// Process chain type
	token.ChainType = types.ChainType(strings.ToLower(chainTypeStr))

	// Process symbol
	if symbolVal, ok := data["symbol"]; ok {
		if symbol, ok := symbolVal.(string); ok {
			token.Symbol = symbol
		} else {
			return nil, fmt.Errorf("invalid symbol type: %T", symbolVal)
		}
	} else {
		return nil, fmt.Errorf("missing required field: symbol")
	}

	// Process type
	if typeVal, ok := data["type"]; ok {
		if tokenType, ok := typeVal.(string); ok {
			token.Type = types.TokenType(strings.ToLower(tokenType))
		} else {
			return nil, fmt.Errorf("invalid type field type: %T", typeVal)
		}
	} else {
		return nil, fmt.Errorf("missing required field: type")
	}

	// Process address
	if addrVal, ok := data["address"]; ok {
		if addr, ok := addrVal.(string); ok {
			token.Address = NormalizeAddress(addr)
		} else {
			return nil, fmt.Errorf("invalid address field type: %T", addrVal)
		}
	} else {
		return nil, fmt.Errorf("missing required field: address")
	}

	// Process decimals
	if decVal, ok := data["decimals"]; ok {
		switch v := decVal.(type) {
		case int:
			token.Decimals = uint8(v)
		case float64:
			token.Decimals = uint8(v)
		case uint8:
			token.Decimals = v
		default:
			return nil, fmt.Errorf("invalid decimals field type: %T", decVal)
		}
	} else {
		return nil, fmt.Errorf("missing required field: decimals")
	}

	// Validate the token
	if err := token.Validate(); err != nil {
		return nil, err
	}

	return &token, nil
}
