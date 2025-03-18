package transaction

import (
	"context"
	"time"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Repository defines the interface for transaction data access
type Repository interface {
	// Create creates a new transaction in the database
	Create(ctx context.Context, tx *Transaction) error

	// GetByTxHash retrieves a transaction by its hash
	GetByTxHash(ctx context.Context, hash string) (*Transaction, error)

	// ListByWallet retrieves transactions for a specific wallet
	// If limit is 0, returns all transactions without pagination
	ListByWallet(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error)

	// ListByAddress retrieves transactions for a specific blockchain address
	// If limit is 0, returns all transactions without pagination
	ListByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error)

	// Exists checks if a transaction exists by its hash
	Exists(ctx context.Context, hash string) (bool, error)
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
	// Generate a Snowflake ID if not provided
	if tx.ID == 0 {
		id, err := r.db.GenerateID()
		if err != nil {
			return err
		}
		tx.ID = id
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

	return err
}

// GetByTxHash retrieves a transaction by its hash
func (r *repository) GetByTxHash(ctx context.Context, hash string) (*Transaction, error) {
	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, created_at, updated_at
		FROM transactions
		WHERE hash = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, hash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewTransactionNotFoundError(hash)
	}

	tx, err := ScanTransaction(rows)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// ListByWallet retrieves transactions for a specific wallet
func (r *repository) ListByWallet(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error) {
	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, created_at, updated_at
		FROM transactions
		WHERE wallet_id = ?
		ORDER BY timestamp DESC
	`

	args := []any{walletID}

	// Add pagination if limit > 0
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := r.db.ExecuteQueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx, err := ScanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return types.NewPage(transactions, offset, limit), nil
}

// ListByAddress retrieves transactions for a specific blockchain address
func (r *repository) ListByAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error) {
	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, created_at, updated_at
		FROM transactions
		WHERE chain_type = ? AND (from_address = ? OR to_address = ?)
		ORDER BY timestamp DESC
	`

	args := []any{chainType, address, address}

	// Add pagination if limit > 0
	if limit > 0 {
		query += " LIMIT ? OFFSET ?"
		args = append(args, limit, offset)
	}

	rows, err := r.db.ExecuteQueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx, err := ScanTransaction(rows)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return types.NewPage(transactions, offset, limit), nil
}

// Exists checks if a transaction exists by its hash
func (r *repository) Exists(ctx context.Context, hash string) (bool, error) {
	query := `
		SELECT 1 
		FROM transactions
		WHERE hash = ?
		LIMIT 1
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, hash)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	return rows.Next(), nil
}
