package wallet

import (
	"context"
	"database/sql"
	"fmt"
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

	// List retrieves wallets with token-based pagination
	List(ctx context.Context, limit int, nextToken string) (*types.Page[*Wallet], error)

	// Exists checks if a wallet exists by its chain type and address
	Exists(ctx context.Context, chainType types.ChainType, address string) (bool, error)

	// UpdateBalance updates a wallet's native balance
	UpdateBalance(ctx context.Context, wallet *Wallet, balance *big.Int) error

	// UpdateTokenBalance updates or creates a token balance for a wallet
	UpdateTokenBalance(ctx context.Context, wallet *Wallet, tokenAddress string, balance *big.Int) error

	// GetTokenBalances retrieves all token balances for a wallet
	GetTokenBalances(ctx context.Context, walletID int64) ([]*TokenBalance, error)

	// GetTokenBalance retrieves a specific token balance for a wallet and token address
	GetTokenBalance(ctx context.Context, walletID int64, tokenAddress string) (*TokenBalance, error)

	// UpdateBlockNumber updates only the last_block_number for a given wallet ID
	UpdateBlockNumber(ctx context.Context, walletID int64, blockNumber int64) error

	// TokenBalanceExists checks if a token balance entry exists for a given wallet and token address
	TokenBalanceExists(ctx context.Context, wallet *Wallet, tokenAddress string) (bool, error)

	// GetWalletsByKeyID retrieves all non-deleted wallets associated with a specific keystore key ID.
	GetWalletsByKeyID(ctx context.Context, keyID string) ([]*Wallet, error)
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

// List retrieves wallets with token-based pagination
func (r *repository) List(ctx context.Context, limit int, nextToken string) (*types.Page[*Wallet], error) {
	// Create a struct-based select builder
	sb := r.walletStructMap.SelectFrom("wallets")
	sb.Where(sb.IsNull("deleted_at"))

	// Default pagination column
	paginationColumn := "id"

	// Decode the next token
	token, err := types.DecodeNextPageToken(nextToken, paginationColumn)
	if err != nil {
		return nil, err
	}

	// If there is a next token, add the condition to start after the token value
	if token != nil {
		idVal, ok := token.GetValueInt64()
		if !ok {
			return nil, errors.NewInvalidPaginationTokenError(nextToken,
				fmt.Errorf("expected integer ID in token, got %T", token.Value))
		}
		sb.Where(sb.GreaterThan(paginationColumn, idVal))
	}

	// Ensure consistent ordering
	sb.OrderBy(paginationColumn + " ASC")

	// Add pagination (fetch one extra to determine if more exist)
	if limit > 0 {
		sb.Limit(limit + 1)
	}

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	wallets, err := r.executeWalletQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Generate the token function
	generateToken := func(wallet *Wallet) *types.NextPageToken {
		return &types.NextPageToken{
			Column: paginationColumn,
			Value:  wallet.ID,
		}
	}

	return types.NewPage(wallets, limit, generateToken), nil
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
func (r *repository) UpdateBalance(ctx context.Context, wallet *Wallet, balance *big.Int) error {
	// Update the balance and timestamp directly on the passed wallet object
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
		return errors.NewWalletNotFoundError(strconv.FormatInt(wallet.ID, 10))
	}

	return nil
}

// UpdateTokenBalance updates or creates a token balance for a wallet
func (r *repository) UpdateTokenBalance(ctx context.Context, wallet *Wallet, tokenAddress string, balance *big.Int) error {
	// Use wallet's ChainType directly
	chainType := wallet.ChainType

	// Normalize and validate the token address using the wallet's ChainType
	normalizedAddr, err := types.NewAddress(chainType, tokenAddress)
	if err != nil {
		// If address creation fails, return the validation error
		return err
	}
	normalizedTokenAddress := normalizedAddr.ToChecksum() // Use the normalized address string

	// Check if token balance exists using the normalized address and wallet.ID
	sb := r.tokenBalanceStructMap.SelectFrom("wallet_balances")
	sb.Where(sb.Equal("wallet_id", wallet.ID)) // Use wallet.ID
	sb.Where(sb.Equal("token_address", normalizedTokenAddress))
	sqlQuery, args := sb.Build()

	balances, err := r.executeTokenBalanceQuery(ctx, sqlQuery, args...)
	if err != nil {
		// Return a database error if the query fails
		return errors.NewDatabaseError(err)
	}

	// Convert balance to BigInt
	bigIntBalance := types.NewBigInt(balance)
	now := time.Now()

	if len(balances) > 0 {
		// Update existing token balance
		tokenBalance := balances[0]
		// Ensure the existing balance uses the normalized address for comparison/update key
		tokenBalance.TokenAddress = normalizedTokenAddress
		tokenBalance.Balance = bigIntBalance
		tokenBalance.UpdatedAt = now

		// Create update builder
		ub := r.tokenBalanceStructMap.Update("wallet_balances", tokenBalance)
		ub.Where(ub.Equal("wallet_id", wallet.ID)) // Use wallet.ID
		ub.Where(ub.Equal("token_address", normalizedTokenAddress))

		// Build the SQL and args
		updateSQL, updateArgs := ub.Build()

		// Execute the update
		_, err = r.db.ExecuteStatementContext(ctx, updateSQL, updateArgs...)
		if err != nil {
			// Return a database error if the update fails
			return errors.NewDatabaseError(err)
		}
		return nil // Return nil on successful update
	} else {
		// Create new token balance using the normalized address and wallet.ID
		tokenBalance := &TokenBalance{
			WalletID:     wallet.ID, // Use wallet.ID
			TokenAddress: normalizedTokenAddress,
			Balance:      bigIntBalance,
			UpdatedAt:    now,
		}

		// Create insert builder
		ib := r.tokenBalanceStructMap.InsertInto("wallet_balances", tokenBalance)

		// Build the SQL and args
		insertSQL, insertArgs := ib.Build()

		// Execute the insert
		_, err = r.db.ExecuteStatementContext(ctx, insertSQL, insertArgs...)
		if err != nil {
			// Return a database error if the insert fails
			return errors.NewDatabaseError(err)
		}
		return nil // Return nil on successful insert
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

// GetTokenBalance retrieves a specific token balance for a wallet and token address.
// If the balance entry does not exist, it returns a TokenBalance with a zero value and nil error.
func (r *repository) GetTokenBalance(ctx context.Context, walletID int64, tokenAddress string) (*TokenBalance, error) {
	// Normalize and validate the token address. Need wallet ChainType for this.
	// Since we only have walletID, we need to fetch the wallet first or assume the address is already normalized.
	// For simplicity here, let's assume the caller provides a normalized address.
	// A better approach might be to require the Wallet object or fetch it internally.
	normalizedTokenAddress := strings.ToLower(tokenAddress) // Basic normalization, consider checksum

	// Create a struct-based select builder
	sb := r.tokenBalanceStructMap.SelectFrom("wallet_balances")
	sb.Where(sb.Equal("wallet_id", walletID))
	// Use the potentially normalized address for the query
	sb.Where(sb.Equal("lower(token_address)", normalizedTokenAddress))

	// Build the SQL and args
	sqlQuery, args := sb.Build()

	// Execute the query
	tokenBalances, err := r.executeTokenBalanceQuery(ctx, sqlQuery, args...)
	if err != nil {
		// Propagate actual database errors
		return nil, errors.NewDatabaseError(err) // Use a general DB error
	}

	if len(tokenBalances) == 0 {
		// Return a zero balance if not found, using the provided (normalized) address
		return &TokenBalance{
			WalletID:     walletID,
			TokenAddress: tokenAddress, // Return the original address provided by caller
			Balance:      types.ZeroBigInt(),
			UpdatedAt:    time.Time{}, // Zero time indicates it's not from DB
		}, nil
	}

	// Return the found token balance
	return tokenBalances[0], nil
}

// UpdateBlockNumber updates only the last_block_number for a given wallet ID
func (r *repository) UpdateBlockNumber(ctx context.Context, walletID int64, blockNumber int64) error {
	query := `UPDATE wallets SET last_block_number = ? WHERE id = ?`
	_, err := r.db.ExecuteStatementContext(ctx, query, blockNumber, walletID)
	return err
}

// TokenBalanceExists checks if a token balance entry exists for a given wallet and token address
func (r *repository) TokenBalanceExists(ctx context.Context, wallet *Wallet, tokenAddress string) (bool, error) {
	// Normalize and validate the token address using the wallet's ChainType
	normalizedAddr, err := types.NewAddress(wallet.ChainType, tokenAddress)
	if err != nil {
		// If address creation fails, return the validation error
		return false, err
	}
	normalizedTokenAddress := normalizedAddr.ToChecksum()

	// Create a select builder to check for existence
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("1")
	sb.From("wallet_balances")
	sb.Where(sb.Equal("wallet_id", wallet.ID))
	sb.Where(sb.Equal("token_address", normalizedTokenAddress))
	sb.Limit(1)

	// Build the SQL and args
	sqlQuery, args := sb.Build()

	// Execute the query
	rows, err := r.db.ExecuteQueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return false, err // Return raw DB error
	}
	defer rows.Close()

	// rows.Next() returns true if a row was found, false otherwise
	exists := rows.Next()

	// Check for any errors during row iteration
	if err := rows.Err(); err != nil {
		return false, err // Return raw DB error
	}

	return exists, nil
}

// GetWalletsByKeyID retrieves all non-deleted wallets associated with a specific keystore key ID.
func (r *repository) GetWalletsByKeyID(ctx context.Context, keyID string) ([]*Wallet, error) {
	sb := r.walletStructMap.SelectFrom("wallets")
	sb.Where(sb.Equal("key_id", keyID))
	sb.Where(sb.IsNull("deleted_at"))

	sqlQuery, args := sb.Build()

	wallets, err := r.executeWalletQuery(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err
	}

	return wallets, nil
}
