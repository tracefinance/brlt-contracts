package signer

import (
	"context"
	"database/sql"
	"time"

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
	List(ctx context.Context, limit, offset int) (*types.Page[*Signer], error)

	// AddAddress creates a new address for a signer
	AddAddress(ctx context.Context, address *Address) error

	// DeleteAddress removes an address from a signer
	DeleteAddress(ctx context.Context, id int64) error

	// GetAddresses retrieves all addresses for a signer
	GetAddresses(ctx context.Context, signerID int64) ([]*Address, error)
}

// repository implements Repository using SQLite database
type repository struct {
	db *db.DB
}

// NewRepository creates a new SQLite signer repository
func NewRepository(db *db.DB) Repository {
	return &repository{db: db}
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

	query := `
	INSERT INTO signers (id, name, type, user_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecuteStatementContext(
		ctx, query, signer.ID, signer.Name, signer.Type, signer.UserID, signer.CreatedAt, signer.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// Update modifies an existing signer's information
func (r *repository) Update(ctx context.Context, signer *Signer) error {
	// Only update the timestamp if not already set
	if signer.UpdatedAt.IsZero() {
		signer.UpdatedAt = time.Now()
	}

	query := `
	UPDATE signers
	SET name = ?, type = ?, user_id = ?, updated_at = ?
	WHERE id = ?`

	result, err := r.db.ExecuteStatementContext(
		ctx, query, signer.Name, signer.Type, signer.UserID, signer.UpdatedAt, signer.ID,
	)
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
	_, err = tx.ExecContext(ctx, "DELETE FROM signer_addresses WHERE signer_id = ?", id)
	if err != nil {
		return err
	}

	// Then delete the signer
	result, err := tx.ExecContext(ctx, "DELETE FROM signers WHERE id = ?", id)
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
	query := `
	SELECT id, name, type, user_id, created_at, updated_at
	FROM signers
	WHERE id = ?`

	rows, err := r.db.ExecuteQueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewSignerNotFoundError(id)
	}

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

	// Load addresses
	addresses, err := r.GetAddresses(ctx, id)
	if err != nil {
		return nil, err
	}
	signer.Addresses = addresses

	return &signer, nil
}

// GetByUserID retrieves all signers for a user
func (r *repository) GetByUserID(ctx context.Context, userID int64) ([]*Signer, error) {
	query := `
	SELECT id, name, type, user_id, created_at, updated_at
	FROM signers
	WHERE user_id = ?`

	rows, err := r.db.ExecuteQueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var signers []*Signer
	for rows.Next() {
		var signer Signer
		var userIDValue sql.NullInt64

		err = rows.Scan(
			&signer.ID,
			&signer.Name,
			&signer.Type,
			&userIDValue,
			&signer.CreatedAt,
			&signer.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		if userIDValue.Valid {
			uid := userIDValue.Int64
			signer.UserID = &uid
		}

		signers = append(signers, &signer)
	}

	if err = rows.Err(); err != nil {
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
func (r *repository) List(ctx context.Context, limit, offset int) (*types.Page[*Signer], error) {
	// Apply default pagination if not specified
	if limit < 1 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Get signers with pagination
	query := `
	SELECT id, name, type, user_id, created_at, updated_at
	FROM signers
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?`

	rows, err := r.db.ExecuteQueryContext(ctx, query, limit, offset)
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

	// Load addresses for each signer
	for _, signer := range signers {
		addresses, err := r.GetAddresses(ctx, signer.ID)
		if err != nil {
			return nil, err
		}
		signer.Addresses = addresses
	}

	// Get total count to determine if there are more results
	var count int
	countQuery := "SELECT COUNT(*) FROM signers"
	countRows, err := r.db.ExecuteQueryContext(ctx, countQuery)
	if err != nil {
		return nil, err
	}
	defer countRows.Close()

	if !countRows.Next() {
		return nil, err
	}

	if err = countRows.Scan(&count); err != nil {
		return nil, err
	}

	// Create pagination result
	hasMore := (offset + limit) < count

	return &types.Page[*Signer]{
		Items:   signers,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}, nil
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

	query := `
	INSERT INTO signer_addresses (id, signer_id, chain_type, address, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecuteStatementContext(
		ctx, query, address.ID, address.SignerID, address.ChainType, address.Address, address.CreatedAt, address.UpdatedAt,
	)

	if err != nil {
		return err
	}

	return nil
}

// DeleteAddress removes an address from a signer
func (r *repository) DeleteAddress(ctx context.Context, id int64) error {
	query := "DELETE FROM signer_addresses WHERE id = ?"

	result, err := r.db.ExecuteStatementContext(ctx, query, id)
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
	query := `
	SELECT id, signer_id, chain_type, address, created_at, updated_at
	FROM signer_addresses
	WHERE signer_id = ?
	ORDER BY chain_type ASC`

	rows, err := r.db.ExecuteQueryContext(ctx, query, signerID)
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
