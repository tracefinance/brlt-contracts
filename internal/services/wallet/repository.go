package wallet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"vault0/internal/core/db"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Repository defines the interface for wallet data access
type Repository interface {
	// Create creates a new wallet in the database
	Create(ctx context.Context, wallet *Wallet) error

	// Get retrieves a wallet by its chain type and address
	Get(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error)

	// GetByID retrieves a wallet by its ID
	GetByID(ctx context.Context, id string) (*Wallet, error)

	// Update updates a wallet's name, tags and last block number
	Update(ctx context.Context, wallet *Wallet) error

	// Delete deletes a wallet by its chain type and address
	Delete(ctx context.Context, chainType types.ChainType, address string) error

	// List retrieves wallets with optional filtering
	List(ctx context.Context, limit, offset int) ([]*Wallet, error)

	// Exists checks if a wallet exists by its chain type and address
	Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error)
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
		return errors.NewOperationFailedError("marshal wallet tags", err)
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
		return err
	}

	return nil
}

// Get retrieves a wallet by its chain type and address
func (r *repository) Get(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
	query := `
		SELECT id, key_id, chain_type, address, name, tags, last_block_number, created_at, updated_at, deleted_at
		FROM wallets
		WHERE chain_type = ? AND address = ? AND deleted_at IS NULL
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, address)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewResourceNotFoundError("wallet", fmt.Sprintf("%s:%s", chainType, address))
	}

	wallet, err := ScanWallet(rows)
	if err != nil {
		return nil, errors.NewOperationFailedError("scan wallet", err)
	}

	return wallet, nil
}

// GetByID retrieves a wallet by its ID
func (r *repository) GetByID(ctx context.Context, id string) (*Wallet, error) {
	query := `
		SELECT id, key_id, chain_type, address, name, tags, last_block_number, created_at, updated_at, deleted_at
		FROM wallets
		WHERE id = ? AND deleted_at IS NULL
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, id)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewResourceNotFoundError("wallet", id)
	}

	wallet, err := ScanWallet(rows)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return wallet, nil
}

// Update updates a wallet's name, tags and last block number
func (r *repository) Update(ctx context.Context, wallet *Wallet) error {
	// Convert tags to JSON
	tagsJSON, err := json.Marshal(wallet.Tags)
	if err != nil {
		return errors.NewOperationFailedError("marshal wallet tags", err)
	}

	wallet.UpdatedAt = time.Now()

	query := `
		UPDATE wallets
		SET name = ?, tags = ?, last_block_number = ?, updated_at = ?
		WHERE id = ? AND deleted_at IS NULL
	`

	result, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		wallet.Name,
		string(tagsJSON),
		wallet.LastBlockNumber,
		wallet.UpdatedAt,
		wallet.ID,
	)

	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewResourceNotFoundError("wallet", wallet.ID)
	}

	return nil
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
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewResourceNotFoundError("wallet", fmt.Sprintf("%s:%s", chainType, address))
	}

	return nil
}

// List retrieves wallets with optional filtering
func (r *repository) List(ctx context.Context, limit, offset int) ([]*Wallet, error) {
	query := `
		SELECT id, key_id, chain_type, address, name, tags, last_block_number, created_at, updated_at, deleted_at
		FROM wallets
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var wallets []*Wallet
	for rows.Next() {
		wallet, err := ScanWallet(rows)
		if err != nil {
			return nil, errors.NewDatabaseError(err)
		}
		wallets = append(wallets, wallet)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return wallets, nil
}

// Exists checks if a wallet exists by its chain type and address
func (r *repository) Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
	query := `
		SELECT 1
		FROM wallets
		WHERE chain_type = ? AND address = ? AND deleted_at IS NULL
		LIMIT 1
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, address)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	exists := rows.Next()

	if err = rows.Err(); err != nil {
		return false, errors.NewDatabaseError(err)
	}

	return exists, nil
}
