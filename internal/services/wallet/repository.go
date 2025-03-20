package wallet

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/govalues/decimal"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Repository defines the interface for wallet data access
type Repository interface {
	// Create creates a new wallet in the database
	Create(ctx context.Context, wallet *Wallet) error

	// GetByAddress retrieves a wallet by its chain type and address
	GetByAddress(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error)

	// GetByID retrieves a wallet by its ID
	GetByID(ctx context.Context, id int64) (*Wallet, error)

	// Update updates a wallet's name, tags and last block number
	Update(ctx context.Context, wallet *Wallet) error

	// Delete deletes a wallet by its chain type and address
	Delete(ctx context.Context, chainType types.ChainType, address string) error

	// List retrieves wallets with optional filtering
	// If limit is 0, returns all wallets without pagination
	List(ctx context.Context, limit, offset int) (*types.Page[*Wallet], error)

	// Exists checks if a wallet exists by its chain type and address
	Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error)

	// UpdateBalance updates a wallet's native balance
	UpdateBalance(ctx context.Context, id int64, balance decimal.Decimal) error

	// GetWalletBalances retrieves a wallet's native and token balances
	GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalance, error)

	// UpdateTokenBalance updates or creates a token balance for a wallet
	UpdateTokenBalance(ctx context.Context, walletID, tokenID int64, balance decimal.Decimal) error

	// GetTokenBalances retrieves all token balances for a wallet
	GetTokenBalances(ctx context.Context, walletID int64) ([]*TokenBalance, error)
}

// repository implements Repository interface for SQLite
type repository struct {
	db *db.DB
}

// NewRepository creates a new SQLite repository for wallets
func NewRepository(db *db.DB) Repository {
	return &repository{db: db}
}

// Create inserts a new wallet into the database
func (r *repository) Create(ctx context.Context, wallet *Wallet) error {
	// Generate a new Snowflake ID if not provided
	if wallet.ID == 0 {
		var err error
		wallet.ID, err = r.db.GenerateID()
		if err != nil {
			return err
		}
	}

	// Set timestamps
	now := time.Now()
	wallet.CreatedAt = now
	wallet.UpdatedAt = now

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(wallet.Tags)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO wallets (id, key_id, chain_type, address, name, tags, balance, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecuteStatementContext(
		ctx,
		query,
		wallet.ID,
		wallet.KeyID,
		wallet.ChainType,
		wallet.Address,
		wallet.Name,
		string(tagsJSON),
		wallet.Balance,
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// GetByAddress retrieves a wallet by its chain type and address
func (r *repository) GetByAddress(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
	// Normalize the address for consistent database queries
	normalizedAddress := types.NormalizeAddress(address)

	query := `
		SELECT id, key_id, chain_type, address, name, tags, balance, last_block_number, created_at, updated_at, deleted_at
		FROM wallets
		WHERE chain_type = ? AND lower(address) = ? AND deleted_at IS NULL
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, normalizedAddress)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewWalletNotFoundError(address)
	}

	wallet, err := ScanWallet(rows)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// GetByID retrieves a wallet by its ID
func (r *repository) GetByID(ctx context.Context, id int64) (*Wallet, error) {
	query := `
		SELECT id, key_id, chain_type, address, name, tags, balance, last_block_number, created_at, updated_at, deleted_at
		FROM wallets
		WHERE id = ? AND deleted_at IS NULL
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewWalletNotFoundError(strconv.FormatInt(id, 10))
	}

	wallet, err := ScanWallet(rows)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}

// Update updates a wallet's name, tags and last block number
func (r *repository) Update(ctx context.Context, wallet *Wallet) error {
	// Convert tags to JSON
	tagsJSON, err := json.Marshal(wallet.Tags)
	if err != nil {
		return err
	}

	wallet.UpdatedAt = time.Now()

	query := `
		UPDATE wallets
		SET name = ?, tags = ?, balance = ?, last_block_number = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		wallet.Name,
		string(tagsJSON),
		wallet.Balance,
		wallet.LastBlockNumber,
		wallet.UpdatedAt,
		wallet.ID,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewWalletNotFoundError(strconv.FormatInt(wallet.ID, 10))
	}

	return nil
}

// Delete deletes a wallet by its chain type and address
func (r *repository) Delete(ctx context.Context, chainType types.ChainType, address string) error {
	// Normalize the address for consistent database queries
	normalizedAddress := types.NormalizeAddress(address)

	query := `
		UPDATE wallets
		SET deleted_at = ?
		WHERE chain_type = ? AND lower(address) = ? AND deleted_at IS NULL
	`

	now := time.Now().UTC()
	result, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		now,
		chainType,
		normalizedAddress,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewWalletNotFoundError(address)
	}

	return nil
}

// List retrieves wallets with optional filtering
func (r *repository) List(ctx context.Context, limit, offset int) (*types.Page[*Wallet], error) {
	query := `
		SELECT id, key_id, chain_type, address, name, tags, balance, last_block_number, created_at, updated_at, deleted_at
		FROM wallets
		WHERE deleted_at IS NULL
	`

	// Default pagination values
	if limit <= 0 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Add pagination
	query += " LIMIT ? OFFSET ?"

	// Execute the query
	rows, err := r.db.ExecuteQueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Scan the results
	var wallets []*Wallet
	for rows.Next() {
		wallet, err := ScanWallet(rows)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Count total wallets for pagination
	var total int
	countQuery := `
		SELECT COUNT(*) FROM wallets WHERE deleted_at IS NULL
	`

	countRows, err := r.db.ExecuteQueryContext(ctx, countQuery)
	if err != nil {
		return nil, err
	}
	defer countRows.Close()

	if countRows.Next() {
		if err := countRows.Scan(&total); err != nil {
			return nil, err
		}
	}

	// Calculate hasMore
	hasMore := offset+len(wallets) < total

	// Create and return the page
	return &types.Page[*Wallet]{
		Items:   wallets,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
}

// Exists checks if a wallet exists by its chain type and address
func (r *repository) Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
	// Normalize the address for consistent database queries
	normalizedAddress := types.NormalizeAddress(address)

	query := `
		SELECT EXISTS(
			SELECT 1 FROM wallets 
			WHERE chain_type = ? AND lower(address) = ? AND deleted_at IS NULL
		)
	`

	var exists bool
	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, normalizedAddress)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&exists); err != nil {
			return false, err
		}
	}

	return exists, nil
}

// UpdateBalance updates a wallet's native balance
func (r *repository) UpdateBalance(ctx context.Context, id int64, balance decimal.Decimal) error {
	query := `
		UPDATE wallets
		SET balance = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		balance.String(),
		time.Now(),
		id,
	)

	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewWalletNotFoundError(strconv.FormatInt(id, 10))
	}

	return nil
}

// GetWalletBalances retrieves a wallet's native and token balances
func (r *repository) GetWalletBalances(ctx context.Context, id int64) ([]*TokenBalance, error) {
	return r.GetTokenBalances(ctx, id)
}

// UpdateTokenBalance updates or creates a token balance for a wallet
func (r *repository) UpdateTokenBalance(ctx context.Context, walletID, tokenID int64, balance decimal.Decimal) error {
	query := `
		INSERT INTO token_balances (wallet_id, token_id, balance, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (wallet_id, token_id) DO UPDATE
		SET balance = ?, updated_at = ?
	`

	now := time.Now()
	_, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		walletID,
		tokenID,
		balance.String(),
		now,
		balance.String(),
		now,
	)

	if err != nil {
		return err
	}

	return nil
}

// GetTokenBalances retrieves all token balances for a wallet
func (r *repository) GetTokenBalances(ctx context.Context, walletID int64) ([]*TokenBalance, error) {
	query := `
		SELECT wallet_id, token_id, balance, updated_at
		FROM token_balances
		WHERE wallet_id = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokenBalances []*TokenBalance
	for rows.Next() {
		tokenBalance, err := ScanTokenBalance(rows)
		if err != nil {
			return nil, err
		}
		tokenBalances = append(tokenBalances, tokenBalance)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tokenBalances, nil
}
