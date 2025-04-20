package signer

import (
	"context"
	"database/sql"
	"time"

	"github.com/huandu/go-sqlbuilder"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Repository defines the signer data access interface
type Repository interface {
	// Create adds a new signer to the database
	Create(ctx context.Context, signer *Signer) error

	// Update modifies an existing signer's information
	Update(ctx context.Context, signer *Signer) error

	// Delete removes a signer from the database
	Delete(ctx context.Context, id int64) error

	// GetByID retrieves a signer by their unique ID
	GetByID(ctx context.Context, id int64) (*Signer, error)

	// GetByUserID retrieves all signers for a user
	GetByUserID(ctx context.Context, userID int64) ([]*Signer, error)

	// List retrieves a paginated collection of signers
	// When limit=0, returns all signers without pagination
	// nextToken is used for token-based pagination
	List(ctx context.Context, limit int, nextToken string) (*types.Page[*Signer], error)

	// AddAddress creates a new address for a signer
	AddAddress(ctx context.Context, address *Address) error

	// DeleteAddress removes an address from a signer
	DeleteAddress(ctx context.Context, id int64) error

	// GetAddresses retrieves all addresses for a signer
	GetAddresses(ctx context.Context, signerID int64) ([]*Address, error)
}

// repository implements Repository using SQLite database
type repository struct {
	db               *db.DB
	signerStructMap  *sqlbuilder.Struct
	addressStructMap *sqlbuilder.Struct
}

// NewRepository creates a new SQLite signer repository
func NewRepository(db *db.DB) Repository {
	signerStructMap := sqlbuilder.NewStruct(new(Signer))
	addressStructMap := sqlbuilder.NewStruct(new(Address))

	return &repository{
		db:               db,
		signerStructMap:  signerStructMap,
		addressStructMap: addressStructMap,
	}
}

// executeSignerQuery executes a query and scans the results into Signer objects
func (r *repository) executeSignerQuery(ctx context.Context, query string, args ...interface{}) ([]*Signer, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signers []*Signer
	for rows.Next() {
		var signer Signer
		var userID sql.NullInt64

		err = rows.Scan(
			&signer.ID,
			&signer.Name,
			&signer.Type,
			&userID,
			&signer.CreatedAt,
			&signer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			uid := userID.Int64
			signer.UserID = &uid
		}

		signers = append(signers, &signer)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return signers, nil
}

// executeAddressQuery executes a query and scans the results into Address objects
func (r *repository) executeAddressQuery(ctx context.Context, query string, args ...interface{}) ([]*Address, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var addresses []*Address
	for rows.Next() {
		var address Address
		err = rows.Scan(
			&address.ID,
			&address.SignerID,
			&address.ChainType,
			&address.Address,
			&address.CreatedAt,
			&address.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		addresses = append(addresses, &address)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return addresses, nil
}

// Create adds a new signer to the database
func (r *repository) Create(ctx context.Context, signer *Signer) error {
	// Generate a new Snowflake ID if not provided
	if signer.ID == 0 {
		var err error
		signer.ID, err = r.db.GenerateID()
		if err != nil {
			return err
		}
	}

	// Only set timestamps if not already set
	if signer.CreatedAt.IsZero() {
		now := time.Now()
		signer.CreatedAt = now
		signer.UpdatedAt = now
	}

	// Create a struct-based insert builder
	ib := r.signerStructMap.InsertInto("signers", signer)

	// Build the SQL and args
	sql, args := ib.Build()

	// Execute the insert
	_, err := r.db.ExecuteStatementContext(ctx, sql, args...)
	return err
}

// Update modifies an existing signer's information
func (r *repository) Update(ctx context.Context, signer *Signer) error {
	// Only update the timestamp if not already set
	if signer.UpdatedAt.IsZero() {
		signer.UpdatedAt = time.Now()
	}

	// Create a struct-based update builder
	ub := r.signerStructMap.Update("signers", signer)
	ub.Where(ub.Equal("id", signer.ID))

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
		return errors.NewSignerNotFoundError(signer.ID)
	}

	return nil
}

// Delete removes a signer from the database
func (r *repository) Delete(ctx context.Context, id int64) error {
	tx, err := r.db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First delete all addresses for this signer
	// Create a delete builder for addresses
	addrDb := sqlbuilder.NewDeleteBuilder()
	addrDb.DeleteFrom("signer_addresses")
	addrDb.Where(addrDb.Equal("signer_id", id))
	addrSql, addrArgs := addrDb.Build()

	_, err = tx.ExecContext(ctx, addrSql, addrArgs...)
	if err != nil {
		return err
	}

	// Then delete the signer
	// Create a delete builder for signer
	signerDb := sqlbuilder.NewDeleteBuilder()
	signerDb.DeleteFrom("signers")
	signerDb.Where(signerDb.Equal("id", id))
	signerSql, signerArgs := signerDb.Build()

	result, err := tx.ExecContext(ctx, signerSql, signerArgs...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewSignerNotFoundError(id)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetByID retrieves a signer by their unique ID
func (r *repository) GetByID(ctx context.Context, id int64) (*Signer, error) {
	// Create a struct-based select builder
	sb := r.signerStructMap.SelectFrom("signers")
	sb.Where(sb.Equal("id", id))

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	signers, err := r.executeSignerQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(signers) == 0 {
		return nil, errors.NewSignerNotFoundError(id)
	}

	// Load addresses
	addresses, err := r.GetAddresses(ctx, id)
	if err != nil {
		return nil, err
	}
	signers[0].Addresses = addresses

	return signers[0], nil
}

// GetByUserID retrieves all signers for a user
func (r *repository) GetByUserID(ctx context.Context, userID int64) ([]*Signer, error) {
	// Create a struct-based select builder
	sb := r.signerStructMap.SelectFrom("signers")
	sb.Where(sb.Equal("user_id", userID))

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	signers, err := r.executeSignerQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Load addresses for each signer
	for _, signer := range signers {
		addresses, err := r.GetAddresses(ctx, signer.ID)
		if err != nil {
			return nil, err
		}
		signer.Addresses = addresses
	}

	return signers, nil
}

// List retrieves a paginated collection of signers
func (r *repository) List(ctx context.Context, limit int, nextToken string) (*types.Page[*Signer], error) {
	// Create a struct-based select builder
	sb := r.signerStructMap.SelectFrom("signers")

	// Default sort by id
	paginationColumn := "id"

	// If nextToken is provided, decode it to get the starting point
	token, err := types.DecodeNextPageToken(nextToken, paginationColumn)
	if err != nil {
		return nil, err
	}

	if token != nil {
		sb.Where(sb.GreaterThan(paginationColumn, token.Value))
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
	signers, err := r.executeSignerQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Load addresses for each signer
	for _, signer := range signers {
		addresses, err := r.GetAddresses(ctx, signer.ID)
		if err != nil {
			return nil, err
		}
		signer.Addresses = addresses
	}

	// Generate the token function
	generateToken := func(signer *Signer) *types.NextPageToken {
		return &types.NextPageToken{
			Column: paginationColumn,
			Value:  signer.ID,
		}
	}

	return types.NewPage(signers, limit, generateToken), nil
}

// AddAddress creates a new address for a signer
func (r *repository) AddAddress(ctx context.Context, address *Address) error {
	// Generate a new Snowflake ID if not provided
	if address.ID == 0 {
		var err error
		address.ID, err = r.db.GenerateID()
		if err != nil {
			return err
		}
	}

	// Only set timestamps if not already set
	if address.CreatedAt.IsZero() {
		now := time.Now()
		address.CreatedAt = now
		address.UpdatedAt = now
	}

	// Create a struct-based insert builder
	ib := r.addressStructMap.InsertInto("signer_addresses", address)

	// Build the SQL and args
	sql, args := ib.Build()

	// Execute the insert
	_, err := r.db.ExecuteStatementContext(ctx, sql, args...)
	return err
}

// DeleteAddress removes an address from a signer
func (r *repository) DeleteAddress(ctx context.Context, id int64) error {
	// Create a delete builder
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom("signer_addresses")
	db.Where(db.Equal("id", id))

	// Build the SQL and args
	sql, args := db.Build()

	// Execute the delete
	result, err := r.db.ExecuteStatementContext(ctx, sql, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.NewSignerAddressNotFoundError(id)
	}

	return nil
}

// GetAddresses retrieves all addresses for a signer
func (r *repository) GetAddresses(ctx context.Context, signerID int64) ([]*Address, error) {
	// Create a struct-based select builder
	sb := r.addressStructMap.SelectFrom("signer_addresses")
	sb.Where(sb.Equal("signer_id", signerID))
	sb.OrderBy("chain_type ASC")

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	return r.executeAddressQuery(ctx, sql, args...)
}
