package transaction

import (
	"context"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"

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
	db        *db.DB
	log       logger.Logger
	structMap *sqlbuilder.Struct
}

// NewRepository creates a new SQLite repository for transactions
func NewRepository(db *db.DB, log logger.Logger) Repository {
	// Create a struct mapper for Transaction
	structMap := sqlbuilder.NewStruct(new(Transaction))

	return &repository{
		db:        db,
		log:       log,
		structMap: structMap,
	}
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

	// Normalize addresses using the new Address struct
	if tx.FromAddress != "" {
		fromAddr, err := types.NewAddress(tx.FromAddress, tx.ChainType)
		if err != nil {
			return err
		}
		tx.FromAddress = fromAddr.Address
	}

	if tx.ToAddress != "" {
		toAddr, err := types.NewAddress(tx.ToAddress, tx.ChainType)
		if err != nil {
			return err
		}
		tx.ToAddress = toAddr.Address
	}

	if tx.TokenAddress != "" {
		tokenAddr, err := types.NewAddress(tx.TokenAddress, tx.ChainType)
		if err != nil {
			return err
		}
		tx.TokenAddress = tokenAddr.Address
	}

	// Create a struct-based insert builder with the transaction value
	insertBuilder := r.structMap.InsertInto("transactions", tx)

	// Build the SQL and args
	sql, args := insertBuilder.Build()

	// Execute the insert
	_, err := r.db.ExecuteStatementContext(ctx, sql, args...)
	return err
}

// GetByTxHash retrieves a transaction by its hash
func (r *repository) GetByTxHash(ctx context.Context, hash string) (*Transaction, error) {
	// Create a struct-based select builder
	sb := r.structMap.SelectFrom("transactions")
	sb.Where(sb.Equal("hash", hash))
	sb.Limit(1)

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	transactions, err := r.executeTransactionQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return nil, errors.NewTransactionNotFoundError(hash)
	}

	return transactions[0], nil
}

// executeTransactionQuery executes a query and scans the results into Transaction objects
func (r *repository) executeTransactionQuery(ctx context.Context, sql string, args ...interface{}) ([]*Transaction, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
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

	return transactions, nil
}

// ListByWalletID retrieves transactions for a specific wallet
func (r *repository) ListByWalletID(ctx context.Context, walletID int64, limit, offset int) (*types.Page[*Transaction], error) {
	// Create a struct-based select builder
	sb := r.structMap.SelectFrom("transactions")
	sb.Where(sb.Equal("wallet_id", walletID))
	sb.OrderBy("timestamp DESC")

	// Add pagination if limit > 0
	if limit > 0 {
		sb.Limit(limit)
		sb.Offset(offset)
	}

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	transactions, err := r.executeTransactionQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return types.NewPage(transactions, offset, limit), nil
}

// ListByWalletAddress retrieves transactions for a specific blockchain address
func (r *repository) ListByWalletAddress(ctx context.Context, chainType types.ChainType, address string, limit, offset int) (*types.Page[*Transaction], error) {
	// Just lowercase the address for querying
	lowercaseAddress := strings.ToLower(address)

	// Create a struct-based select builder
	sb := r.structMap.SelectFrom("transactions")
	sb.Where(sb.Equal("chain_type", chainType))
	sb.Where(sb.Or(
		sb.Equal("lower(from_address)", lowercaseAddress),
		sb.Equal("lower(to_address)", lowercaseAddress),
	))
	sb.OrderBy("timestamp DESC")

	// Add pagination if limit > 0
	if limit > 0 {
		sb.Limit(limit)
		sb.Offset(offset)
	}

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	transactions, err := r.executeTransactionQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return types.NewPage(transactions, offset, limit), nil
}

// List retrieves transactions based on the provided filter criteria
func (r *repository) List(ctx context.Context, filter *Filter) (*types.Page[*Transaction], error) {
	// Create a struct-based select builder
	sb := r.structMap.SelectFrom("transactions")

	// Apply filters
	if filter.Status != nil {
		sb.Where(sb.Equal("status", *filter.Status))
	}

	if filter.ChainType != nil {
		sb.Where(sb.Equal("chain_type", *filter.ChainType))
	}

	if filter.WalletID != nil {
		sb.Where(sb.Equal("wallet_id", *filter.WalletID))
	}

	if filter.Address != nil {
		// Just lowercase the address for querying
		lowercaseAddress := strings.ToLower(*filter.Address)
		sb.Where(sb.Or(
			sb.Equal("lower(from_address)", lowercaseAddress),
			sb.Equal("lower(to_address)", lowercaseAddress),
		))
	}

	if filter.TokenAddress != nil {
		tokenAddress := strings.ToLower(*filter.TokenAddress)

		// Special handling for native transactions
		if tokenAddress == "native" || tokenAddress == types.ZeroAddress {
			sb.Where(sb.Or(
				sb.Equal("type", string(types.TransactionTypeNative)),
				sb.IsNull("token_address"),
			))
		} else {
			sb.Where(sb.Equal("lower(token_address)", tokenAddress))
		}
	}

	// Order by most recent first
	sb.OrderBy("timestamp DESC")

	// Add pagination if limit > 0
	if filter.Limit > 0 {
		sb.Limit(filter.Limit)
		sb.Offset(filter.Offset)
	}

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	transactions, err := r.executeTransactionQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	return types.NewPage(transactions, filter.Offset, filter.Limit), nil
}

// Exists checks if a transaction exists by its hash
func (r *repository) Exists(ctx context.Context, hash string) (bool, error) {
	// Create a struct-based select builder using a minimized struct selection
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("1")
	sb.From("transactions")
	sb.Where(sb.Equal("hash", hash))
	sb.Limit(1)

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
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

	// Normalize addresses using the new Address struct
	if tx.FromAddress != "" {
		fromAddr, err := types.NewAddress(tx.FromAddress, tx.ChainType)
		if err != nil {
			return err
		}
		tx.FromAddress = fromAddr.Address
	}

	if tx.ToAddress != "" {
		toAddr, err := types.NewAddress(tx.ToAddress, tx.ChainType)
		if err != nil {
			return err
		}
		tx.ToAddress = toAddr.Address
	}

	if tx.TokenAddress != "" {
		tokenAddr, err := types.NewAddress(tx.TokenAddress, tx.ChainType)
		if err != nil {
			return err
		}
		tx.TokenAddress = tokenAddr.Address
	}

	// Create a struct-based update builder
	ub := r.structMap.Update("transactions", tx)
	ub.Where(ub.Equal("id", tx.ID))

	// Build the SQL and args
	sql, args := ub.Build()

	// Execute the update
	result, err := r.db.ExecuteStatementContext(ctx, sql, args...)
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
