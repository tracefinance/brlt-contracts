package tokenprice

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Repository defines the interface for data access operations related to token prices.
type Repository interface {
	// UpsertMany inserts new token prices or updates existing ones based on the Symbol.
	// It should perform this operation efficiently, potentially using batching or specific
	// database features for bulk upserts.
	//
	// Returns:
	//   - The number of rows affected.
	//   - An error if the database operation fails (e.g., ErrDatabaseOperationFailed).
	UpsertMany(ctx context.Context, prices []*TokenPrice) (int64, error)

	// GetBySymbol retrieves the stored price data for a specific token symbol.
	//
	// Returns:
	//   - A pointer to the TokenPrice if found.
	//   - ErrTokenPriceNotFound if no price data exists for the symbol.
	//   - Other database errors (e.g., ErrDatabaseOperationFailed).
	GetBySymbol(ctx context.Context, symbol string) (*TokenPrice, error)

	// List retrieves a paginated list of stored token prices, optionally filtered and sorted.
	// Uses token-based pagination for consistent results.
	//
	// Parameters:
	//   - ctx: The context for the operation
	//   - filter: Optional filtering criteria
	//   - limit: Maximum number of items to return (0 for all items)
	//   - nextToken: Token for pagination (empty string for first page)
	//
	// Returns:
	//   - A page of token prices with pagination information
	//   - An error if the database operation fails
	List(ctx context.Context, filter *TokenPriceFilter, limit int, nextToken string) (*types.Page[*TokenPrice], error)
}

// repository implements Repository interface for SQLite
type repository struct {
	db        *db.DB
	log       logger.Logger
	structMap *sqlbuilder.Struct
}

// NewRepository creates a new SQLite repository for token prices
func NewRepository(db *db.DB, log logger.Logger) Repository {
	// Create a struct mapper for TokenPrice
	structMap := sqlbuilder.NewStruct(new(TokenPrice))

	return &repository{
		db:        db,
		log:       log,
		structMap: structMap,
	}
}

// executeTokenPriceQuery executes a query and scans the results into TokenPrice objects
func (r *repository) executeTokenPriceQuery(ctx context.Context, sql string, args ...any) ([]*TokenPrice, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prices []*TokenPrice
	for rows.Next() {
		price, err := ScanTokenPrice(rows)
		if err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prices, nil
}

// UpsertMany inserts or updates multiple token prices in the database
// First finds existing tokens by symbols, then updates them, and inserts new ones
func (r *repository) UpsertMany(ctx context.Context, prices []*TokenPrice) (int64, error) {
	if len(prices) == 0 {
		return 0, nil
	}

	// Start a transaction
	conn := r.db.GetConnection()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Update timestamp for all prices
	now := time.Now()
	for i := range prices {
		prices[i].UpdatedAt = now
	}

	// Extract symbols for lookup
	symbols := make([]string, 0, len(prices))
	pricesBySymbol := make(map[string]*TokenPrice, len(prices))
	for _, price := range prices {
		upperSymbol := strings.ToUpper(price.Symbol)
		symbols = append(symbols, upperSymbol)
		pricesBySymbol[upperSymbol] = price
	}

	// Use FindBySymbols to get existing price data
	existingPrices, err := r.mapPriceBySymbol(ctx, symbols)
	if err != nil {
		r.log.Error("Failed to find existing token prices", logger.Error(err))
		return 0, err
	}

	// Separate prices into updates and inserts
	var toUpdate []*TokenPrice
	var toInsert []*TokenPrice

	for _, price := range prices {
		if _, exists := existingPrices[strings.ToUpper(price.Symbol)]; exists {
			toUpdate = append(toUpdate, price)
		} else {
			toInsert = append(toInsert, price)
		}
	}

	// Track affected rows
	var totalAffected int64

	// Prepare update statement if needed
	if len(toUpdate) > 0 {
		updateSQL := `UPDATE token_prices SET 
			rank = ?, price_usd = ?, supply = ?, 
			market_cap_usd = ?, volume_usd_24h = ?, updated_at = ? 
			WHERE symbol = ?`

		// For debugging, log the SQL
		r.log.Debug("Generated update SQL", logger.String("sql", updateSQL))

		updateStmt, err := tx.Prepare(updateSQL)
		if err != nil {
			r.log.Error("Failed to prepare update statement", logger.Error(err))
			return 0, err
		}
		defer updateStmt.Close()

		// Execute updates
		for _, price := range toUpdate {
			result, err := updateStmt.ExecContext(
				ctx,
				price.Rank,
				price.PriceUSD,
				price.Supply,
				price.MarketCapUSD,
				price.VolumeUSD24h,
				price.UpdatedAt,
				price.Symbol, // WHERE clause param
			)

			if err != nil {
				r.log.Error("Failed to update token price",
					logger.String("symbol", price.Symbol),
					logger.String("sql", updateSQL),
					logger.Error(err))
				continue // Skip this one but continue with others
			}

			// Add affected rows
			affected, err := result.RowsAffected()
			if err == nil {
				totalAffected += affected
			}
		}
	}

	// Prepare insert statement if needed
	if len(toInsert) > 0 {
		// Create a manual SQL statement for insert
		insertSQL := `INSERT INTO token_prices (
			symbol, rank, price_usd, supply, market_cap_usd, volume_usd_24h, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)`

		// For debugging, log the SQL
		r.log.Debug("Generated insert SQL", logger.String("sql", insertSQL))

		insertStmt, err := tx.Prepare(insertSQL)
		if err != nil {
			r.log.Error("Failed to prepare insert statement", logger.Error(err))
			return 0, err
		}
		defer insertStmt.Close()

		// Execute inserts
		for _, price := range toInsert {
			_, err := insertStmt.ExecContext(
				ctx,
				price.Symbol,
				price.Rank,
				price.PriceUSD,
				price.Supply,
				price.MarketCapUSD,
				price.VolumeUSD24h,
				price.UpdatedAt,
			)

			if err != nil {
				r.log.Error("Failed to insert token price",
					logger.String("symbol", price.Symbol),
					logger.String("sql", insertSQL),
					logger.Error(err))
				continue // Skip this one but continue with others
			}

			// Count successful inserts
			totalAffected++
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		r.log.Error("Failed to commit transaction", logger.Error(err))
		return 0, err
	}

	return totalAffected, nil
}

// GetBySymbol retrieves a token price by its symbol
func (r *repository) GetBySymbol(ctx context.Context, symbol string) (*TokenPrice, error) {
	// Convert symbol to uppercase for consistency
	upperSymbol := strings.ToUpper(symbol)

	// Create a struct-based select builder
	sb := r.structMap.SelectFrom("token_prices")
	sb.Where(sb.Equal("symbol", upperSymbol))
	sb.Limit(1)

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	prices, err := r.executeTokenPriceQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(prices) == 0 {
		return nil, errors.NewTokenPriceNotFoundError(symbol)
	}

	return prices[0], nil
}

// mapPriceBySymbol retrieves token prices for a list of symbols
func (r *repository) mapPriceBySymbol(ctx context.Context, symbols []string) (map[string]*TokenPrice, error) {
	if len(symbols) == 0 {
		return make(map[string]*TokenPrice), nil
	}

	// Normalize symbols to uppercase for consistent comparison
	upperSymbols := make([]any, len(symbols))
	for i, symbol := range symbols {
		upperSymbols[i] = strings.ToUpper(symbol)
	}

	// Create a select builder for the query
	sb := r.structMap.SelectFrom("token_prices")
	sb.Where(sb.In("symbol", upperSymbols...))
	// Order by rank
	sb.OrderBy("rank ASC")

	// Build and execute the query
	sql, args := sb.Build()
	prices, err := r.executeTokenPriceQuery(ctx, sql, args...)
	if err != nil {
		r.log.Error("Failed to find token prices by symbols",
			logger.Int("symbol_count", len(symbols)),
			logger.Error(err))
		return nil, err
	}

	// Create a map of symbol to token price
	result := make(map[string]*TokenPrice, len(prices))
	for _, price := range prices {
		result[price.Symbol] = price
	}

	return result, nil
}

// List retrieves a paginated list of token prices based on filter criteria
func (r *repository) List(ctx context.Context, filter *TokenPriceFilter, limit int, nextToken string) (*types.Page[*TokenPrice], error) {
	// Create a select builder for the query
	sb := r.structMap.SelectFrom("token_prices")

	// Apply filtering if symbols are provided
	if filter != nil && len(filter.Symbols) > 0 {
		// Normalize symbols to uppercase for consistent comparison
		upperSymbols := make([]any, len(filter.Symbols))
		for i, symbol := range filter.Symbols {
			upperSymbols[i] = strings.ToUpper(symbol)
		}

		// Add WHERE clause for symbols
		sb.Where(sb.In("symbol", upperSymbols...))
	}

	// Apply sorting by rank for consistent pagination
	sb.OrderBy("rank ASC")

	// Decode the token using rank as the expected column
	token, err := types.DecodeNextPageToken(nextToken, "rank")
	if err != nil {
		return nil, err
	}

	// Apply pagination only if token is not nil
	if token != nil {
		// Get rank value as an integer
		rankInt, ok := token.GetValueInt64()
		if !ok {
			return nil, errors.NewInvalidPaginationTokenError(nextToken,
				fmt.Errorf("rank value must be an integer, got %T", token.Value))
		}

		// Use rank directly as int
		sb.Where(sb.GreaterThan("rank", rankInt))
	}

	// Apply limit only if specified (> 0). Fetch one extra to determine if there are more pages
	if limit > 0 {
		sb.Limit(limit + 1)
	}

	// Build and execute the query
	sql, args := sb.Build()
	prices, err := r.executeTokenPriceQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Create a properly paginated response using NewPage
	// NewPage now handles both limit=0 and regular pagination cases
	return types.NewPage(prices, limit, func(price *TokenPrice) *types.NextPageToken {
		return &types.NextPageToken{
			Column: "rank",
			Value:  price.Rank,
		}
	}), nil
}

// ScanTokenPrice scans a database row into a TokenPrice struct
func ScanTokenPrice(rows interface {
	Scan(dest ...any) error
}) (*TokenPrice, error) {
	var price TokenPrice
	err := rows.Scan(
		&price.Symbol,
		&price.Rank,
		&price.PriceUSD,
		&price.Supply,
		&price.MarketCapUSD,
		&price.VolumeUSD24h,
		&price.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &price, nil
}
