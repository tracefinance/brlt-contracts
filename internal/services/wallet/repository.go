package wallet

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vault0/internal/core/db"
	"vault0/internal/types"
)

// Repository defines the interface for wallet data access
type Repository interface {
	// Create creates a new wallet in the database
	Create(ctx context.Context, wallet *Wallet) error

	// Get retrieves a wallet by its chain type and address
	Get(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error)

	// Update updates a wallet's name and tags by chain type and address
	Update(ctx context.Context, chainType types.ChainType, address string, name string, tags map[string]string) (*Wallet, error)

	// Delete deletes a wallet by its chain type and address
	Delete(ctx context.Context, chainType types.ChainType, address string) error

	// List retrieves wallets with optional filtering
	List(ctx context.Context, limit, offset int) ([]*Wallet, error)
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
	// Generate a new UUID if not provided
	if wallet.ID == "" {
		wallet.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	wallet.CreatedAt = now
	wallet.UpdatedAt = now

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(wallet.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		INSERT INTO wallets (id, key_id, chain_type, address, name, tags, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
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
		wallet.CreatedAt,
		wallet.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create wallet: %w", err)
	}

	return nil
}

// Get retrieves a wallet by its chain type and address
func (r *repository) Get(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
	query := `
		SELECT id, key_id, chain_type, address, name, tags, created_at, updated_at, deleted_at
		FROM wallets
		WHERE chain_type = ? AND address = ? AND deleted_at IS NULL
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, address)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallet: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	wallet, err := ScanWallet(rows)
	if err != nil {
		return nil, fmt.Errorf("failed to scan wallet: %w", err)
	}

	return wallet, nil
}

// Update updates a wallet's name and tags by chain type and address
func (r *repository) Update(ctx context.Context, chainType types.ChainType, address string, name string, tags map[string]string) (*Wallet, error) {
	// Get the wallet first
	wallet, err := r.Get(ctx, chainType, address)
	if err != nil {
		return nil, err
	}

	// Update the wallet fields
	wallet.Name = name
	wallet.Tags = tags
	wallet.UpdatedAt = time.Now()

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(wallet.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	query := `
		UPDATE wallets
		SET name = ?, tags = ?, updated_at = ?
		WHERE chain_type = ? AND address = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		wallet.Name,
		string(tagsJSON),
		wallet.UpdatedAt,
		chainType,
		address,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}

	return wallet, nil
}

// Delete soft deletes a wallet by its chain type and address
func (r *repository) Delete(ctx context.Context, chainType types.ChainType, address string) error {
	query := `
		UPDATE wallets
		SET deleted_at = ?, updated_at = ?
		WHERE chain_type = ? AND address = ? AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		now,
		now,
		chainType,
		address,
	)

	if err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// List retrieves wallets with optional filtering
func (r *repository) List(ctx context.Context, limit, offset int) ([]*Wallet, error) {
	query := `
		SELECT id, key_id, chain_type, address, name, tags, created_at, updated_at, deleted_at
		FROM wallets
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallets: %w", err)
	}
	defer rows.Close()

	var wallets []*Wallet
	for rows.Next() {
		wallet, err := ScanWallet(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan wallet: %w", err)
		}
		wallets = append(wallets, wallet)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating wallet rows: %w", err)
	}

	return wallets, nil
}
