package transaction

import (
	"context"
	"time"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Repository defines the interface for transaction data access
type Repository interface {
	// Create creates a new transaction in the database
	Create(ctx context.Context, tx *Transaction) error

	// Update updates an existing transaction in the database
	Update(ctx context.Context, tx *Transaction) error

	// GetByTxHash retrieves a transaction by its hash
	GetByTxHash(ctx context.Context, hash string) (*Transaction, error)

	// ListByWalletID retrieves transactions for a specific wallet
	// If limit is 0, returns all transactions without pagination
	ListByWalletID(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error)

	// ListByWalletAddress retrieves transactions for a specific blockchain address
	// If limit is 0, returns all transactions without pagination
	ListByWalletAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error)

	// List retrieves transactions based on the provided filter criteria
	List(ctx context.Context, filter *Filter) (*types.Page[*Transaction], error)

	// Exists checks if a transaction exists by its hash
	Exists(ctx context.Context, hash string) (bool, error)
}

// repository implements Repository interface for SQLite
type repository struct {
	db  *db.DB
	log logger.Logger
}

// NewRepository creates a new SQLite repository for transactions
func NewRepository(db *db.DB, log logger.Logger) Repository {
	return &repository{db: db, log: log}
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
			status, timestamp, block_number, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		tx.BlockNumber,
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
			status, timestamp, block_number, created_at, updated_at
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

// ListByWalletID retrieves transactions for a specific wallet
func (r *repository) ListByWalletID(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error) {
	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, block_number, created_at, updated_at
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

// ListByWalletAddress retrieves transactions for a specific blockchain address
func (r *repository) ListByWalletAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error) {
	// Normalize the address for consistent database queries
	normalizedAddress := types.NormalizeAddress(address)

	query := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, block_number, created_at, updated_at
		FROM transactions
		WHERE chain_type = ? AND (lower(from_address) = ? OR lower(to_address) = ?)
		ORDER BY timestamp DESC
	`

	args := []any{chainType, normalizedAddress, normalizedAddress}

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

// List retrieves transactions based on the provided filter criteria
func (r *repository) List(ctx context.Context, filter *Filter) (*types.Page[*Transaction], error) {
	// Base query
	queryBuilder := `
		SELECT 
			id, wallet_id, chain_type, hash, from_address, to_address, 
			value, data, nonce, gas_price, gas_limit, type, token_address, 
			status, timestamp, block_number, created_at, updated_at
		FROM transactions
		WHERE 1=1
	`
	var args []any

	// Apply filters
	if filter.Status != nil {
		queryBuilder += " AND status = ?"
		args = append(args, *filter.Status)
	}

	if filter.ChainType != nil {
		queryBuilder += " AND chain_type = ?"
		args = append(args, *filter.ChainType)
	}

	if filter.WalletID != nil {
		queryBuilder += " AND wallet_id = ?"
		args = append(args, *filter.WalletID)
	}

	if filter.Address != nil {
		// Normalize the address for consistent database queries
		normalizedAddress := types.NormalizeAddress(*filter.Address)
		queryBuilder += " AND (lower(from_address) = ? OR lower(to_address) = ?)"
		args = append(args, normalizedAddress, normalizedAddress)
	}

	// Order by most recent first
	queryBuilder += " ORDER BY timestamp DESC"

	// Add pagination if limit > 0
	if filter.Limit > 0 {
		queryBuilder += " LIMIT ? OFFSET ?"
		args = append(args, filter.Limit, filter.Offset)
	}

	// Execute query
	rows, err := r.db.ExecuteQueryContext(ctx, queryBuilder, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Process results
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

	return types.NewPage(transactions, filter.Offset, filter.Limit), nil
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

// Update updates an existing transaction in the database
func (r *repository) Update(ctx context.Context, tx *Transaction) error {
	// Ensure the transaction ID is provided
	if tx.ID == 0 {
		return errors.NewInvalidInputError("Transaction ID is required for update", "id", "")
	}

	// Update the timestamp
	tx.UpdatedAt = time.Now()

	// Convert big.Int values to strings for storage
	valueStr := ""
	if tx.Value != nil {
		valueStr = tx.Value.String()
	}

	gasPriceStr := ""
	if tx.GasPrice != nil {
		gasPriceStr = tx.GasPrice.String()
	}

	// Update the transaction
	query := `
		UPDATE transactions
		SET wallet_id = ?, 
			chain_type = ?, 
			hash = ?, 
			from_address = ?, 
			to_address = ?, 
			value = ?, 
			data = ?, 
			nonce = ?, 
			gas_price = ?, 
			gas_limit = ?, 
			type = ?, 
			token_address = ?, 
			status = ?, 
			timestamp = ?,
			block_number = ?,
			updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecuteStatementContext(
		ctx,
		query,
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
		tx.BlockNumber,
		tx.UpdatedAt,
		tx.ID,
	)

	if err != nil {
		return err
	}

	// Check if any rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewTransactionNotFoundError(tx.Hash)
	}

	return nil
}
