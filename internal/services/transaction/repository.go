package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	"vault0/internal/core/db"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Repository defines the interface for transaction data access
type Repository interface {
	// Create creates a new transaction in the database
	Create(ctx context.Context, tx *Transaction) error

	// Get retrieves a transaction by its chain type and hash
	Get(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error)

	// GetByWallet retrieves transactions for a specific wallet
	GetByWallet(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error)

	// GetByAddress retrieves transactions for a specific blockchain address
	GetByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) ([]*Transaction, error)

	// Count counts transactions for a specific wallet
	Count(ctx context.Context, walletID string) (int, error)

	// CountByAddress counts transactions for a specific blockchain address
	CountByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error)

	// Exists checks if a transaction exists by its chain type and hash
	Exists(ctx context.Context, chainType types.ChainType, hash string) (bool, error)
}

// repository implements Repository interface for SQLite
type repository struct {
	db *db.DB
}

// NewRepository creates a new SQLite repository for transactions
func NewRepository(db *db.DB) Repository {
	return &repository{db: db}
}

// Create inserts a new transaction into the database
func (r *repository) Create(ctx context.Context, tx *Transaction) error {
	// Generate a new UUID if not provided
	if tx.ID == "" {
		tx.ID = uuid.New().String()
	}

	// Set timestamps
	now := time.Now()
	tx.CreatedAt = now
	tx.UpdatedAt = now

	// Convert big.Int values to strings for storage
	valueStr := ""
	if tx.Value != nil {
		valueStr = tx.Value.String()
	}

	gasPriceStr := ""
	if tx.GasPrice != nil {
		gasPriceStr = tx.GasPrice.String()
	}

	// Insert the transaction
	query := `
		INSERT INTO transactions (
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecuteStatementContext(
		ctx,
		query,
		tx.ID,
		tx.WalletID,
		tx.ChainType,
		tx.Hash,
		tx.FromAddress,
		tx.ToAddress,
		valueStr,
		tx.Data,
		tx.Nonce,
		gasPriceStr,
		tx.GasLimit,
		tx.Type,
		tx.TokenAddress,
		tx.Status,
		tx.Timestamp,
		tx.CreatedAt,
		tx.UpdatedAt,
	)

	if err != nil {
		return errors.NewDatabaseError(err)
	}

	return nil
}

// Get retrieves a transaction by its chain type and hash
func (r *repository) Get(ctx context.Context, chainType types.ChainType, hash string) (*Transaction, error) {
	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, created_at, updated_at
		FROM transactions
		WHERE chain_type = ? AND hash = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, hash)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewTransactionNotFoundError(hash)
	}

	tx, err := ScanTransaction(rows)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return tx, nil
}

// GetByWallet retrieves transactions for a specific wallet
func (r *repository) GetByWallet(ctx context.Context, walletID string, limit, offset int) ([]*Transaction, error) {
	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, created_at, updated_at
		FROM transactions
		WHERE wallet_id = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, walletID, limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx, err := ScanTransaction(rows)
		if err != nil {
			return nil, errors.NewDatabaseError(err)
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return transactions, nil
}

// GetByAddress retrieves transactions for a specific blockchain address
func (r *repository) GetByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) ([]*Transaction, error) {
	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, created_at, updated_at
		FROM transactions
		WHERE chain_type = ? AND (from_address = ? OR to_address = ?)
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, address, address, limit, offset)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx, err := ScanTransaction(rows)
		if err != nil {
			return nil, errors.NewDatabaseError(err)
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	return transactions, nil
}

// Count counts transactions for a specific wallet
func (r *repository) Count(ctx context.Context, walletID string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM transactions
		WHERE wallet_id = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, walletID)
	if err != nil {
		return 0, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.NewDatabaseError(nil)
	}

	var count int
	if err := rows.Scan(&count); err != nil {
		return 0, errors.NewDatabaseError(err)
	}

	return count, nil
}

// CountByAddress counts transactions for a specific blockchain address
func (r *repository) CountByAddress(ctx context.Context, chainType types.ChainType, address string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM transactions
		WHERE chain_type = ? AND (from_address = ? OR to_address = ?)
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, address, address)
	if err != nil {
		return 0, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, errors.NewDatabaseError(nil)
	}

	var count int
	if err := rows.Scan(&count); err != nil {
		return 0, errors.NewDatabaseError(err)
	}

	return count, nil
}

// Exists checks if a transaction exists by its chain type and hash
func (r *repository) Exists(ctx context.Context, chainType types.ChainType, hash string) (bool, error) {
	query := `
		SELECT 1 
		FROM transactions
		WHERE chain_type = ? AND hash = ?
		LIMIT 1
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, chainType, hash)
	if err != nil {
		return false, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	return rows.Next(), nil
}
