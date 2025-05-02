package transaction

import (
	"context"
	"fmt"
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

	// GetByHash retrieves a transaction by its hash
	GetByHash(ctx context.Context, hash string) (*Transaction, error)

	// List retrieves transactions based on the provided filter criteria
	// This is the primary method for retrieving transactions with different filters:
	// - Use filter.WalletID to get transactions for a specific wallet
	// - Use filter.ChainType and filter.Address to get transactions for a specific blockchain address
	// - Use other filter fields for more specific queries
	// If limit is 0, returns all transactions without pagination
	List(ctx context.Context, filter *Filter, limit int, nextToken string) (*types.Page[*Transaction], error)

	// Exists checks if a transaction exists by its hash
	Exists(ctx context.Context, hash string) (bool, error)

	// UpdateTransactionStatus updates only the status and updated_at fields of a transaction by its hash.
	UpdateTransactionStatus(ctx context.Context, txHash string, status types.TransactionStatus) error
}

// repository implements Repository interface for SQLite
type repository struct {
	db        *db.DB
	log       logger.Logger
	structMap *sqlbuilder.Struct
}

// NewRepository creates a new SQLite repository for transactions
func NewRepository(db *db.DB, log logger.Logger) Repository {
	structMap := sqlbuilder.NewStruct(new(Transaction))

	return &repository{
		db:        db,
		log:       log,
		structMap: structMap,
	}
}

// Create inserts a new transaction into the database
func (r *repository) Create(ctx context.Context, tx *Transaction) error {
	if tx.Hash == "" {
		return errors.NewInvalidInputError("Transaction hash cannot be empty for existence check", "hash", "")
	}

	exists, err := r.Exists(ctx, tx.Hash)
	if err != nil {
		return err
	}
	if exists {
		return errors.NewAlreadyExistsError(fmt.Sprintf("transaction with hash %s", tx.Hash))
	}

	if tx.ID == 0 {
		id, err := r.db.GenerateID()
		if err != nil {
			return err
		}
		tx.ID = id
	}

	now := time.Now()
	tx.CreatedAt = now
	tx.UpdatedAt = now

	if tx.From != "" {
		fromAddr, err := types.NewAddress(tx.Chain, tx.From)
		if err != nil {
			return err
		}
		tx.From = fromAddr.Address
	}

	if tx.To != "" {
		toAddr, err := types.NewAddress(tx.Chain, tx.To)
		if err != nil {
			return err
		}
		tx.To = toAddr.Address
	}

	insertBuilder := r.structMap.InsertInto("transactions", tx)

	sql, args := insertBuilder.Build()

	_, err = r.db.ExecuteStatementContext(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}

// GetByHash retrieves a transaction by its hash
func (r *repository) GetByHash(ctx context.Context, hash string) (*Transaction, error) {
	sb := r.structMap.SelectFrom("transactions")
	sb.Where(sb.E("hash", hash))
	sb.Limit(1)

	sql, args := sb.Build()

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
		return nil, errors.NewDatabaseError(err)
	}

	return transactions, nil
}

// List retrieves transactions based on the provided filter criteria
func (r *repository) List(ctx context.Context, filter *Filter, limit int, nextToken string) (*types.Page[*Transaction], error) {

	sb := r.structMap.SelectFrom("transactions")

	if filter != nil {
		if filter.Status != nil {
			sb.Where(sb.E("status", *filter.Status))
		}

		if filter.ChainType != nil {
			sb.Where(sb.E("chain_type", *filter.ChainType))
		}

		if filter.WalletID != nil {
			sb.Where(sb.E("wallet_id", *filter.WalletID))
		}

		if filter.Address != nil {
			if filter.ChainType == nil {
				return nil, errors.NewInvalidInputError("ChainType is required when filtering by Address", "chain_type", "")
			}

			addrNorm, err := types.NewAddress(*filter.ChainType, *filter.Address)
			if err != nil {
				return nil, err
			}

			sb.Where(sb.Or(
				sb.E("from_address", addrNorm.Address),
				sb.E("to_address", addrNorm.Address),
			))
		}

		if filter.Type != nil {
			sb.Where(sb.E("type", *filter.Type))
		}

		if filter.TokenAddress != nil {
			sb.Where(sb.E("token_address", *filter.TokenAddress))
		}

		if filter.BlockNumber != nil {
			sb.Where(sb.E("block_number", filter.BlockNumber.String())) // Compare as string due to DECIMAL storage
		}

		if filter.MinBlock != nil {
			sb.Where(sb.GE("block_number", filter.MinBlock.String())) // Compare as string
		}

		if filter.MaxBlock != nil {
			sb.Where(sb.LE("block_number", filter.MaxBlock.String())) // Compare as string
		}
	}

	// Always exclude soft-deleted records unless explicitly requested (not supported by current filter)
	sb.Where(sb.IsNull("deleted_at"))

	paginationColumn := "id"

	// If nextToken is provided, decode it to get the starting point
	token, err := types.DecodeNextPageToken(nextToken, paginationColumn)
	if err != nil {
		return nil, err
	}

	if token != nil {
		idVal, ok := token.Value.(int64)
		if !ok {
			return nil, errors.NewInvalidParameterError("next_token",
				fmt.Errorf("unexpected token value type: %T", token.Value).Error())
		}

		sb.Where(sb.L(paginationColumn, idVal)) // Using less than because we're ordering by DESC
	}

	sb.OrderBy("id DESC")

	// Add pagination if limit > 0
	if limit > 0 {
		sb.Limit(limit + 1) // Fetch one extra item to check for HasMore
	}

	sql, args := sb.Build()
	r.log.Debug("Listing transactions", logger.String("sql", sql), logger.Any("args", args))

	transactions, err := r.executeTransactionQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	generateToken := func(tx *Transaction) *types.NextPageToken {
		return &types.NextPageToken{
			Column: paginationColumn,
			Value:  tx.ID,
		}
	}

	return types.NewPage(transactions, limit, generateToken), nil
}

// Exists checks if a transaction exists by its hash
func (r *repository) Exists(ctx context.Context, hash string) (bool, error) {
	// Create a struct-based select builder using a minimized struct selection
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("1")
	sb.From("transactions")
	sb.Where(sb.E("hash", hash))
	sb.Limit(1)

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
	if err != nil {
		return false, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	exists := rows.Next()
	if err := rows.Err(); err != nil {
		return false, errors.NewDatabaseError(err)
	}

	return exists, nil
}

// Update updates an existing transaction in the database
func (r *repository) Update(ctx context.Context, tx *Transaction) error {
	if tx.ID == 0 {
		return errors.NewMissingParameterError("transaction ID")
	}

	tx.UpdatedAt = time.Now()

	if tx.From != "" {
		fromAddr, err := types.NewAddress(tx.Chain, tx.From)
		if err != nil {
			return err
		}
		tx.From = fromAddr.Address
	}

	if tx.To != "" {
		toAddr, err := types.NewAddress(tx.Chain, tx.To)
		if err != nil {
			return err
		}
		tx.To = toAddr.Address
	}

	ub := r.structMap.Update("transactions", tx)
	ub.Where(ub.E("id", tx.ID))

	sql, args := ub.Build()
	r.log.Debug("Updating transaction", logger.String("sql", sql), logger.Any("args", args))

	result, err := r.db.ExecuteStatementContext(ctx, sql, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Log this error as it might indicate a driver issue
		r.log.Error("Failed to get rows affected after update", logger.Error(err), logger.Int64("tx_id", tx.ID))
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		// Use the hash if available for a more informative error, otherwise use ID
		hashOrID := fmt.Sprintf("ID %d", tx.ID)
		if tx.Hash != "" {
			hashOrID = fmt.Sprintf("hash %s", tx.Hash)
		}
		return errors.NewTransactionNotFoundError(hashOrID)
	}

	return nil
}

// UpdateTransactionStatus updates only the status and updated_at fields of a transaction by its hash.
func (r *repository) UpdateTransactionStatus(ctx context.Context, txHash string, status types.TransactionStatus) error {
	if txHash == "" {
		return errors.NewMissingParameterError("transaction hash")
	}

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("transactions")
	ub.Set(
		ub.Assign("status", status),
		ub.Assign("updated_at", time.Now()),
	)
	ub.Where(ub.E("hash", txHash))

	sql, args := ub.Build()
	r.log.Debug("Updating transaction status", logger.String("sql", sql), logger.Any("args", args))

	result, err := r.db.ExecuteStatementContext(ctx, sql, args...)
	if err != nil {
		return err // Already wrapped by db layer
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.log.Error("Failed to get rows affected after status update", logger.Error(err), logger.String("tx_hash", txHash))
		return errors.NewDatabaseError(err)
	}

	if rowsAffected == 0 {
		return errors.NewTransactionNotFoundError(txHash)
	}

	return nil
}
