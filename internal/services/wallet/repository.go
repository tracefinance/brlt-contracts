package wallet

import (
	"context"
	"database/sql"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"

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
	UpdateBalance(ctx context.Context, id int64, balance *big.Int) error

	// UpdateTokenBalance updates or creates a token balance for a wallet
	UpdateTokenBalance(ctx context.Context, walletID int64, tokenAddress string, balance *big.Int) error

	// GetTokenBalances retrieves all token balances for a wallet
	GetTokenBalances(ctx context.Context, walletID int64) ([]*TokenBalance, error)

	// UpdateBlockNumber updates only the last_block_number for a given wallet ID
	UpdateBlockNumber(ctx context.Context, walletID int64, blockNumber int64) error
}

// repository implements Repository interface for SQLite
type repository struct {
	db                    *db.DB
	walletStructMap       *sqlbuilder.Struct
	tokenBalanceStructMap *sqlbuilder.Struct
}

// NewRepository creates a new SQLite repository for wallets
func NewRepository(db *db.DB) Repository {
	walletStructMap := sqlbuilder.NewStruct(new(Wallet))
	tokenBalanceStructMap := sqlbuilder.NewStruct(new(TokenBalance))

	return &repository{
		db:                    db,
		walletStructMap:       walletStructMap,
		tokenBalanceStructMap: tokenBalanceStructMap,
	}
}

// executeWalletQuery executes a query and scans the results into Wallet objects
func (r *repository) executeWalletQuery(ctx context.Context, sql string, args ...any) ([]*Wallet, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []*Wallet
	for rows.Next() {
		wallet, err := ScanWallet(rows)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wallets, nil
}

// executeTokenBalanceQuery executes a query and scans the results into TokenBalance objects
func (r *repository) executeTokenBalanceQuery(ctx context.Context, sql string, args ...any) ([]*TokenBalance, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
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

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tokenBalances, nil
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

	// Set balance to 0
	wallet.Balance = types.ZeroBigInt()

	// Normalize wallet address using the new Address struct
	if wallet.Address != "" {
		addr, err := types.NewAddress(wallet.ChainType, wallet.Address)
		if err != nil {
			return err
		}
		wallet.Address = addr.Address
	}

	// Create a struct-based insert builder
	ib := r.walletStructMap.InsertInto("wallets", wallet)

	// Build the SQL and args
	sql, args := ib.Build()

	// Execute the insert
	_, err := r.db.ExecuteStatementContext(ctx, sql, args...)
	return err
}

// GetByAddress retrieves a wallet by its chain type and address
func (r *repository) GetByAddress(ctx context.Context, chainType types.ChainType, address string) (*Wallet, error) {
	// Just lowercase the address for querying
	lowercaseAddress := strings.ToLower(address)

	// Create a struct-based select builder
	sb := r.walletStructMap.SelectFrom("wallets")
	sb.Where(sb.Equal("chain_type", chainType))
	sb.Where(sb.Equal("lower(address)", lowercaseAddress))
	sb.Where(sb.IsNull("deleted_at"))

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	wallets, err := r.executeWalletQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(wallets) == 0 {
		return nil, errors.NewWalletNotFoundError(address)
	}

	return wallets[0], nil
}

// GetByID retrieves a wallet by its ID
func (r *repository) GetByID(ctx context.Context, id int64) (*Wallet, error) {
	// Create a struct-based select builder
	sb := r.walletStructMap.SelectFrom("wallets")
	sb.Where(sb.Equal("id", id))
	sb.Where(sb.IsNull("deleted_at"))

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	wallets, err := r.executeWalletQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(wallets) == 0 {
		return nil, errors.NewWalletNotFoundError(strconv.FormatInt(id, 10))
	}

	return wallets[0], nil
}

// Update updates a wallet's name, tags and last block number
func (r *repository) Update(ctx context.Context, wallet *Wallet) error {
	now := time.Now()
	query := `UPDATE wallets SET name = ?, tags = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL`
	result, err := r.db.ExecuteStatementContext(ctx, query, wallet.Name, wallet.Tags, now, wallet.ID)
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
	// Just lowercase the address for querying
	lowercaseAddress := strings.ToLower(address)

	// First, get the wallet to update
	wallet, err := r.GetByAddress(ctx, chainType, lowercaseAddress)
	if err != nil {
		return err
	}

	// Set deletion timestamp
	deleteTime := time.Now().UTC()
	wallet.DeletedAt = sql.NullTime{Time: deleteTime, Valid: true}

	// Create a struct-based update builder
	ub := r.walletStructMap.Update("wallets", wallet)
	ub.Where(ub.Equal("id", wallet.ID))
	ub.Where(ub.IsNull("deleted_at"))

	// Build the SQL and args
	sql, args := ub.Build()

	// Execute the update
	result, err := r.db.ExecuteStatementContext(ctx, sql, args...)
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
	// Create a struct-based select builder
	sb := r.walletStructMap.SelectFrom("wallets")
	sb.Where(sb.IsNull("deleted_at"))

	// Default pagination values
	if offset < 0 {
		offset = 0
	}

	// Add pagination
	if limit > 0 {
		sb.Limit(limit + 1) // Fetch one extra item to check for HasMore
		sb.Offset(offset)
	}

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	wallets, err := r.executeWalletQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Use the built-in NewPage function which calculates hasMore based on
	// whether we got at least as many items as the limit
	return types.NewPage(wallets, offset, limit), nil
}

// Exists checks if a wallet exists by its chain type and address
func (r *repository) Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error) {
	// Just lowercase the address for querying
	lowercaseAddress := strings.ToLower(address)

	// Create a select builder
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("1")
	sb.From("wallets")
	sb.Where(sb.Equal("chain_type", chainType))
	sb.Where(sb.Equal("lower(address)", lowercaseAddress))
	sb.Where(sb.IsNull("deleted_at"))
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

// UpdateBalance updates a wallet's native balance
func (r *repository) UpdateBalance(ctx context.Context, id int64, balance *big.Int) error {
	// Get the wallet to update
	wallet, err := r.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Update the balance and timestamp
	wallet.Balance = types.NewBigInt(balance)
	wallet.UpdatedAt = time.Now()

	// Create a struct-based update builder
	ub := r.walletStructMap.Update("wallets", wallet)
	ub.Where(ub.Equal("id", wallet.ID))
	ub.Where(ub.IsNull("deleted_at"))

	// Build the SQL and args
	sql, args := ub.Build()

	// Execute the update
	result, err := r.db.ExecuteStatementContext(ctx, sql, args...)
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

// UpdateTokenBalance updates or creates a token balance for a wallet
func (r *repository) UpdateTokenBalance(ctx context.Context, walletID int64, tokenAddress string, balance *big.Int) error {
	// Check if token balance exists
	sb := r.tokenBalanceStructMap.SelectFrom("wallet_balances")
	sb.Where(sb.Equal("wallet_id", walletID))
	sb.Where(sb.Equal("token_address", tokenAddress))
	sql, args := sb.Build()

	balances, err := r.executeTokenBalanceQuery(ctx, sql, args...)
	if err != nil {
		return err
	}

	// Convert balance to BigInt
	bigIntBalance := types.NewBigInt(balance)
	now := time.Now()

	if len(balances) > 0 {
		// Update existing token balance
		tokenBalance := balances[0]
		tokenBalance.Balance = bigIntBalance
		tokenBalance.UpdatedAt = now

		// Create update builder
		ub := r.tokenBalanceStructMap.Update("wallet_balances", tokenBalance)
		ub.Where(ub.Equal("wallet_id", tokenBalance.WalletID))
		ub.Where(ub.Equal("token_address", tokenBalance.TokenAddress))

		// Build the SQL and args
		sql, args := ub.Build()

		// Execute the update
		_, err := r.db.ExecuteStatementContext(ctx, sql, args...)
		return err
	} else {
		// Create new token balance
		tokenBalance := &TokenBalance{
			WalletID:     walletID,
			TokenAddress: tokenAddress,
			Balance:      bigIntBalance,
			UpdatedAt:    now,
		}

		// Create insert builder
		ib := r.tokenBalanceStructMap.InsertInto("wallet_balances", tokenBalance)

		// Build the SQL and args
		sql, args := ib.Build()

		// Execute the insert
		_, err := r.db.ExecuteStatementContext(ctx, sql, args...)
		return err
	}
}

// GetTokenBalances retrieves all token balances for a wallet
func (r *repository) GetTokenBalances(ctx context.Context, walletID int64) ([]*TokenBalance, error) {
	// Create a struct-based select builder
	sb := r.tokenBalanceStructMap.SelectFrom("wallet_balances")
	sb.Where(sb.Equal("wallet_id", walletID))

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	return r.executeTokenBalanceQuery(ctx, sql, args...)
}

// UpdateBlockNumber updates only the last_block_number for a given wallet ID
func (r *repository) UpdateBlockNumber(ctx context.Context, walletID int64, blockNumber int64) error {
	query := `UPDATE wallets SET last_block_number = ? WHERE id = ?`
	_, err := r.db.ExecuteStatementContext(ctx, query, blockNumber, walletID)
	return err
}
