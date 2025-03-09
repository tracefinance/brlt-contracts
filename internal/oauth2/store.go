package oauth2

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"vault0/internal/core/db"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/models"
)

// Vault0Token implements the TokenInfo interface as requested
type Vault0Token struct {
	ClientID         string
	UserID           string
	RedirectURI      string
	Scope            string
	Code             string
	CodeCreateAt     time.Time
	CodeExpiresIn    time.Duration
	Access           string
	AccessCreateAt   time.Time
	AccessExpiresIn  time.Duration
	Refresh          string
	RefreshCreateAt  time.Time
	RefreshExpiresIn time.Duration
}

// New creates a new token based on the given data
func (t *Vault0Token) New() oauth2.TokenInfo {
	return &models.Token{}
}

// GetClientID returns the client ID
func (t *Vault0Token) GetClientID() string {
	return t.ClientID
}

// SetClientID sets the client ID
func (t *Vault0Token) SetClientID(clientID string) {
	t.ClientID = clientID
}

// GetUserID returns the user ID
func (t *Vault0Token) GetUserID() string {
	return t.UserID
}

// SetUserID sets the user ID
func (t *Vault0Token) SetUserID(userID string) {
	t.UserID = userID
}

// GetRedirectURI returns the redirect URI
func (t *Vault0Token) GetRedirectURI() string {
	return t.RedirectURI
}

// SetRedirectURI sets the redirect URI
func (t *Vault0Token) SetRedirectURI(redirectURI string) {
	t.RedirectURI = redirectURI
}

// GetScope returns the scope
func (t *Vault0Token) GetScope() string {
	return t.Scope
}

// SetScope sets the scope
func (t *Vault0Token) SetScope(scope string) {
	t.Scope = scope
}

// GetCode returns the authorization code
func (t *Vault0Token) GetCode() string {
	return t.Code
}

// SetCode sets the authorization code
func (t *Vault0Token) SetCode(code string) {
	t.Code = code
}

// GetCodeCreateAt returns the authorization code creation time
func (t *Vault0Token) GetCodeCreateAt() time.Time {
	return t.CodeCreateAt
}

// SetCodeCreateAt sets the authorization code creation time
func (t *Vault0Token) SetCodeCreateAt(createAt time.Time) {
	t.CodeCreateAt = createAt
}

// GetCodeExpiresIn returns the authorization code expiration time
func (t *Vault0Token) GetCodeExpiresIn() time.Duration {
	return t.CodeExpiresIn
}

// SetCodeExpiresIn sets the authorization code expiration time
func (t *Vault0Token) SetCodeExpiresIn(exp time.Duration) {
	t.CodeExpiresIn = exp
}

// GetAccess returns the access token
func (t *Vault0Token) GetAccess() string {
	return t.Access
}

// SetAccess sets the access token
func (t *Vault0Token) SetAccess(access string) {
	t.Access = access
}

// GetAccessCreateAt returns the access token creation time
func (t *Vault0Token) GetAccessCreateAt() time.Time {
	return t.AccessCreateAt
}

// SetAccessCreateAt sets the access token creation time
func (t *Vault0Token) SetAccessCreateAt(createAt time.Time) {
	t.AccessCreateAt = createAt
}

// GetAccessExpiresIn returns the access token expiration time
func (t *Vault0Token) GetAccessExpiresIn() time.Duration {
	return t.AccessExpiresIn
}

// SetAccessExpiresIn sets the access token expiration time
func (t *Vault0Token) SetAccessExpiresIn(exp time.Duration) {
	t.AccessExpiresIn = exp
}

// GetRefresh returns the refresh token
func (t *Vault0Token) GetRefresh() string {
	return t.Refresh
}

// SetRefresh sets the refresh token
func (t *Vault0Token) SetRefresh(refresh string) {
	t.Refresh = refresh
}

// GetRefreshCreateAt returns the refresh token creation time
func (t *Vault0Token) GetRefreshCreateAt() time.Time {
	return t.RefreshCreateAt
}

// SetRefreshCreateAt sets the refresh token creation time
func (t *Vault0Token) SetRefreshCreateAt(createAt time.Time) {
	t.RefreshCreateAt = createAt
}

// GetRefreshExpiresIn returns the refresh token expiration time
func (t *Vault0Token) GetRefreshExpiresIn() time.Duration {
	return t.RefreshExpiresIn
}

// SetRefreshExpiresIn sets the refresh token expiration time
func (t *Vault0Token) SetRefreshExpiresIn(exp time.Duration) {
	t.RefreshExpiresIn = exp
}

// GetCodeChallenge returns the code challenge (PKCE)
func (t *Vault0Token) GetCodeChallenge() string {
	return ""
}

// SetCodeChallenge sets the code challenge (PKCE)
func (t *Vault0Token) SetCodeChallenge(challenge string) {
}

// GetCodeChallengeMethod returns the code challenge method (PKCE)
func (t *Vault0Token) GetCodeChallengeMethod() oauth2.CodeChallengeMethod {
	return oauth2.CodeChallengeMethod("")
}

// SetCodeChallengeMethod sets the code challenge method (PKCE)
func (t *Vault0Token) SetCodeChallengeMethod(method oauth2.CodeChallengeMethod) {
}

// TokenStore implements the oauth2.TokenStore interface for SQLite
type TokenStore struct {
	db *db.DB
}

// NewTokenStore creates a new token store
func NewTokenStore(database *db.DB) (*TokenStore, error) {
	return &TokenStore{db: database}, nil
}

// Create creates and stores the new token information
func (s *TokenStore) Create(ctx context.Context, info oauth2.TokenInfo) error {
	// Convert to our token model
	token := &Vault0Token{
		ClientID:         info.GetClientID(),
		UserID:           info.GetUserID(),
		RedirectURI:      info.GetRedirectURI(),
		Scope:            info.GetScope(),
		Code:             info.GetCode(),
		CodeCreateAt:     info.GetCodeCreateAt(),
		CodeExpiresIn:    info.GetCodeExpiresIn(),
		Access:           info.GetAccess(),
		AccessCreateAt:   info.GetAccessCreateAt(),
		AccessExpiresIn:  info.GetAccessExpiresIn(),
		Refresh:          info.GetRefresh(),
		RefreshCreateAt:  info.GetRefreshCreateAt(),
		RefreshExpiresIn: info.GetRefreshExpiresIn(),
	}

	query := `
	INSERT INTO tokens (
		client_id, user_id, redirect_uri, scope, 
		code, code_created_at, code_expires_in,
		access, access_created_at, access_expires_in,
		refresh, refresh_created_at, refresh_expires_in
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.ExecuteStatementContext(ctx, query,
		token.ClientID,
		token.UserID,
		token.RedirectURI,
		token.Scope,
		token.Code,
		token.CodeCreateAt.Unix(),
		int64(token.CodeExpiresIn.Seconds()),
		token.Access,
		token.AccessCreateAt.Unix(),
		int64(token.AccessExpiresIn.Seconds()),
		token.Refresh,
		token.RefreshCreateAt.Unix(),
		int64(token.RefreshExpiresIn.Seconds()),
	)

	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	return nil
}

// RemoveByCode deletes the authorization code
func (s *TokenStore) RemoveByCode(ctx context.Context, code string) error {
	query := "UPDATE tokens SET code = '' WHERE code = ?"
	_, err := s.db.ExecuteStatementContext(ctx, query, code)
	if err != nil {
		return fmt.Errorf("failed to remove code: %w", err)
	}
	return nil
}

// RemoveByAccess deletes the access token
func (s *TokenStore) RemoveByAccess(ctx context.Context, access string) error {
	query := "UPDATE tokens SET access = '' WHERE access = ?"
	_, err := s.db.ExecuteStatementContext(ctx, query, access)
	if err != nil {
		return fmt.Errorf("failed to remove access token: %w", err)
	}
	return nil
}

// RemoveByRefresh deletes the refresh token
func (s *TokenStore) RemoveByRefresh(ctx context.Context, refresh string) error {
	query := "UPDATE tokens SET refresh = '' WHERE refresh = ?"
	_, err := s.db.ExecuteStatementContext(ctx, query, refresh)
	if err != nil {
		return fmt.Errorf("failed to remove refresh token: %w", err)
	}
	return nil
}

// GetByCode retrieves token by authorization code
func (s *TokenStore) GetByCode(ctx context.Context, code string) (oauth2.TokenInfo, error) {
	query := `
	SELECT 
		client_id, user_id, redirect_uri, scope, 
		code, code_created_at, code_expires_in,
		access, access_created_at, access_expires_in,
		refresh, refresh_created_at, refresh_expires_in
	FROM tokens WHERE code = ?
	`

	rows, err := s.db.ExecuteQueryContext(ctx, query, code)
	if err != nil {
		return nil, fmt.Errorf("failed to query token by code: %w", err)
	}
	defer rows.Close()

	token, err := s.scanToken(rows)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// GetByAccess retrieves token by access token
func (s *TokenStore) GetByAccess(ctx context.Context, access string) (oauth2.TokenInfo, error) {
	query := `
	SELECT 
		client_id, user_id, redirect_uri, scope, 
		code, code_created_at, code_expires_in,
		access, access_created_at, access_expires_in,
		refresh, refresh_created_at, refresh_expires_in
	FROM tokens WHERE access = ?
	`

	rows, err := s.db.ExecuteQueryContext(ctx, query, access)
	if err != nil {
		return nil, fmt.Errorf("failed to query token by access: %w", err)
	}
	defer rows.Close()

	token, err := s.scanToken(rows)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// GetByRefresh retrieves token by refresh token
func (s *TokenStore) GetByRefresh(ctx context.Context, refresh string) (oauth2.TokenInfo, error) {
	query := `
	SELECT 
		client_id, user_id, redirect_uri, scope, 
		code, code_created_at, code_expires_in,
		access, access_created_at, access_expires_in,
		refresh, refresh_created_at, refresh_expires_in
	FROM tokens WHERE refresh = ?
	`

	rows, err := s.db.ExecuteQueryContext(ctx, query, refresh)
	if err != nil {
		return nil, fmt.Errorf("failed to query token by refresh: %w", err)
	}
	defer rows.Close()

	token, err := s.scanToken(rows)
	if err != nil {
		return nil, err
	}

	return token, nil
}

// scanToken scans a database row into a token
func (s *TokenStore) scanToken(rows *sql.Rows) (*models.Token, error) {
	if !rows.Next() {
		return nil, fmt.Errorf("token not found")
	}

	var (
		clientID         string
		userID           string
		redirectURI      string
		scope            string
		code             string
		codeCreatedAt    int64
		codeExpiresIn    int64
		access           string
		accessCreatedAt  int64
		accessExpiresIn  int64
		refresh          string
		refreshCreatedAt int64
		refreshExpiresIn int64
	)

	err := rows.Scan(
		&clientID,
		&userID,
		&redirectURI,
		&scope,
		&code,
		&codeCreatedAt,
		&codeExpiresIn,
		&access,
		&accessCreatedAt,
		&accessExpiresIn,
		&refresh,
		&refreshCreatedAt,
		&refreshExpiresIn,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan token: %w", err)
	}

	token := &models.Token{
		ClientID:         clientID,
		UserID:           userID,
		RedirectURI:      redirectURI,
		Scope:            scope,
		Code:             code,
		CodeCreateAt:     time.Unix(codeCreatedAt, 0),
		CodeExpiresIn:    time.Duration(codeExpiresIn) * time.Second,
		Access:           access,
		AccessCreateAt:   time.Unix(accessCreatedAt, 0),
		AccessExpiresIn:  time.Duration(accessExpiresIn) * time.Second,
		Refresh:          refresh,
		RefreshCreateAt:  time.Unix(refreshCreatedAt, 0),
		RefreshExpiresIn: time.Duration(refreshExpiresIn) * time.Second,
	}

	return token, nil
}

// ===============================================================================
// Client Store Implementation
// ===============================================================================

// ClientStore implements the oauth2.ClientStore interface for SQLite
type ClientStore struct {
	db *db.DB
}

// NewClientStore creates a new client store
func NewClientStore(database *db.DB) (*ClientStore, error) {
	return &ClientStore{db: database}, nil
}

// GetByID retrieves a client by ID
func (s *ClientStore) GetByID(ctx context.Context, id string) (oauth2.ClientInfo, error) {
	query := "SELECT client_id, client_secret, redirect_uri FROM clients WHERE client_id = ?"
	rows, err := s.db.ExecuteQueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query client: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("client not found")
	}

	var (
		clientID     string
		clientSecret string
		redirectURI  string
	)

	err = rows.Scan(&clientID, &clientSecret, &redirectURI)
	if err != nil {
		return nil, fmt.Errorf("failed to scan client: %w", err)
	}

	client := &models.Client{
		ID:     clientID,
		Secret: clientSecret,
	}

	if redirectURI != "" {
		client.Domain = redirectURI
	}

	return client, nil
}

// Create adds a new client
func (s *ClientStore) Create(clientID, clientSecret, redirectURI string) error {
	query := "INSERT INTO clients (client_id, client_secret, redirect_uri) VALUES (?, ?, ?)"
	_, err := s.db.ExecuteStatement(query, clientID, clientSecret, redirectURI)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}
	return nil
}
