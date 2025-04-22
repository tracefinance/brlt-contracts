package vault

import (
	"context"
	"database/sql" // Import encoding/json
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/logger" // Assuming logger is needed
	"vault0/internal/types"
)

// VaultFilter defines the filtering criteria for listing vaults
type VaultFilter struct {
	ID       *int64
	WalletID *int64
	Address  *string
	Status   *VaultStatus
}

// Repository defines the interface for vault data access
type Repository interface {
	// Create creates a new vault in the database
	Create(ctx context.Context, vault *Vault) error

	// List retrieves vaults with filtering and token-based pagination
	List(ctx context.Context, filter VaultFilter, limit int, nextToken string) (*types.Page[*Vault], error)

	// GetByID retrieves a vault by its ID
	GetByID(ctx context.Context, vaultID int64) (*Vault, error)

	// GetByHash retrieves a vault by its deployment transaction hash
	GetByHash(ctx context.Context, txHash string) (*Vault, error)

	// GetByAddress retrieves a vault by its deployed contract address
	GetByAddress(ctx context.Context, address string) (*Vault, error)

	// UpdateStatus updates a vault's status
	UpdateStatus(ctx context.Context, vaultID int64, status VaultStatus) error

	// Update updates specific fields of a vault: name, recovery_request_timestamp, failure_reason
	Update(ctx context.Context, vaultID int64, vault *Vault) error

	// Delete marks a vault as deleted
	Delete(ctx context.Context, vaultID int64) error
}

// repository implements Repository interface for the database
type repository struct {
	db             *db.DB
	logger         logger.Logger // Assuming logger is needed
	vaultStructMap *sqlbuilder.Struct
}

// NewRepository creates a new repository for vaults
func NewRepository(db *db.DB, logger logger.Logger) Repository {
	vaultStructMap := sqlbuilder.NewStruct(new(Vault))
	// No need for MustMapper here, handle JSON manually

	return &repository{
		db:             db,
		logger:         logger,
		vaultStructMap: vaultStructMap,
	}
}

// ScanVault scans a single row into a Vault struct
func ScanVault(row *sql.Rows) (*Vault, error) {
	var v Vault
	var recoveryRequestTimestamp sql.NullTime
	var failureReason sql.NullString
	var deletedAt sql.NullTime
	var address sql.NullString

	// Assume the columns are in the order defined in the Vault struct for simplicity
	// Adjust scan order based on actual SELECT statement if needed
	err := row.Scan(
		&v.ID,
		&v.Name,
		&v.WalletID,
		&v.ChainType,
		&v.TxHash,
		&v.RecoveryAddress,
		&v.Signers, // Scan directly into v.Signers (uses types.JSONArray.Scan)
		&v.Status,
		&v.SignatureThreshold,
		&address,
		&recoveryRequestTimestamp,
		&failureReason,
		&v.CreatedAt,
		&v.UpdatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, errors.NewDatabaseError(err)
	}

	// Handle nullable fields
	if recoveryRequestTimestamp.Valid {
		v.RecoveryRequestTimestamp = &recoveryRequestTimestamp.Time
	}
	if failureReason.Valid {
		v.FailureReason = &failureReason.String
	}
	if deletedAt.Valid {
		v.DeletedAt = &deletedAt.Time
	}
	if address.Valid {
		v.Address = address.String
	} else {
		v.Address = ""
	}

	return &v, nil
}

// executeVaultQuery executes a query and scans the results into Vault objects
func (r *repository) executeVaultQuery(ctx context.Context, sql string, args ...any) ([]*Vault, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
	if err != nil {
		// Use the correct constructor signature
		return nil, errors.NewDatabaseError(err)
	}
	defer rows.Close()

	var vaults []*Vault
	for rows.Next() {
		vault, err := ScanVault(rows)
		if err != nil {
			return nil, err // Error already wrapped by ScanVault
		}
		vaults = append(vaults, vault)
	}

	if err := rows.Err(); err != nil {
		// Use the correct constructor signature
		return nil, errors.NewDatabaseError(err)
	}

	return vaults, nil
}

// Create inserts a new vault into the database
func (r *repository) Create(ctx context.Context, vault *Vault) error {
	// Generate a new Snowflake ID if not provided
	if vault.ID == 0 {
		var err error
		vault.ID, err = r.db.GenerateID()
		if err != nil {
			// Use the correct constructor signature
			return errors.NewDatabaseError(err)
		}
	}

	// Set timestamps
	now := time.Now().UTC()
	vault.CreatedAt = now
	vault.UpdatedAt = now
	vault.DeletedAt = nil // Ensure DeletedAt is nil on creation

	// Set default status if empty
	if vault.Status == "" {
		vault.Status = VaultStatusPending
	}

	// Validate required fields based on NOT NULL columns in schema
	if vault.Name == "" {
		return errors.NewValidationError(map[string]any{"name": "vault name cannot be empty"})
	}
	if vault.WalletID == 0 {
		return errors.NewValidationError(map[string]any{"wallet_id": "wallet_id cannot be zero"})
	}
	if vault.ChainType == "" {
		return errors.NewValidationError(map[string]any{"chain_type": "chain_type cannot be empty"})
	}
	if vault.TxHash == "" {
		return errors.NewValidationError(map[string]any{"tx_hash": "tx_hash cannot be empty"})
	}
	if vault.RecoveryAddress == "" {
		return errors.NewValidationError(map[string]any{"recovery_address": "recovery_address cannot be empty"})
	}
	if vault.Signers == nil || len(vault.Signers) == 0 {
		return errors.NewValidationError(map[string]any{"signers": "signers cannot be empty"})
	}
	if vault.SignatureThreshold <= 0 {
		return errors.NewValidationError(map[string]any{"signature_threshold": "signature_threshold must be positive"})
	}

	// Use the Value() method of types.JSONArray for insertion
	signersValue, err := vault.Signers.Value()
	if err != nil {
		return errors.NewDatabaseError(err) // Error during JSON marshalling
	}

	// Use direct INSERT statement
	query := `
		INSERT INTO vaults (
			id, name, wallet_id, chain_type, tx_hash, recovery_address, signers,
			status, signature_threshold, address, recovery_request_timestamp,
			failure_reason, created_at, updated_at, deleted_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	// Handle nullable pointers correctly when preparing args
	var recoveryTimestampArg sql.NullTime
	if vault.RecoveryRequestTimestamp != nil {
		recoveryTimestampArg = sql.NullTime{Time: *vault.RecoveryRequestTimestamp, Valid: true}
	}
	var failureReasonArg sql.NullString
	if vault.FailureReason != nil {
		failureReasonArg = sql.NullString{String: *vault.FailureReason, Valid: true}
	}

	args := []interface{}{
		vault.ID, vault.Name, vault.WalletID, vault.ChainType, vault.TxHash,
		vault.RecoveryAddress, signersValue, vault.Status, vault.SignatureThreshold,
		sql.NullString{String: vault.Address, Valid: vault.Address != ""},
		recoveryTimestampArg,
		failureReasonArg,
		vault.CreatedAt, vault.UpdatedAt, vault.DeletedAt,
	}

	_, err = r.db.ExecuteStatementContext(ctx, query, args...)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	return nil
}

// List retrieves vaults with filtering and token-based pagination
func (r *repository) List(ctx context.Context, filter VaultFilter, limit int, nextToken string) (*types.Page[*Vault], error) {
	sb := sqlbuilder.NewSelectBuilder()
	// Select all columns explicitly as defined in ScanVault
	sb.Select(
		"id", "name", "wallet_id", "chain_type", "tx_hash", "recovery_address",
		"signers", "status", "signature_threshold", "address", "recovery_request_timestamp",
		"failure_reason", "created_at", "updated_at", "deleted_at",
	)
	sb.From("vaults")
	sb.Where(sb.IsNull("deleted_at"))

	// Apply filters
	if filter.ID != nil {
		sb.Where(sb.Equal("id", *filter.ID))
	}
	if filter.WalletID != nil {
		sb.Where(sb.Equal("wallet_id", *filter.WalletID))
	}
	if filter.Address != nil {
		sb.Where(sb.Equal("address", *filter.Address))
	}
	if filter.Status != nil {
		sb.Where(sb.Equal("status", *filter.Status))
	}

	// Default pagination column
	paginationColumn := "id"

	// Decode the next token
	token, err := types.DecodeNextPageToken(nextToken, paginationColumn)
	if err != nil {
		return nil, err // Error already wrapped by DecodeNextPageToken
	}

	// Apply pagination condition
	if token != nil {
		// Assuming ID is always increasing for simplicity
		sb.Where(sb.GreaterThan(paginationColumn, token.Value))
	}

	// Ensure consistent ordering
	sb.OrderBy(paginationColumn + " ASC")

	// Add pagination limit (fetch one extra to determine if more exist)
	fetchLimit := 0
	if limit > 0 {
		fetchLimit = limit + 1
		sb.Limit(fetchLimit)
	}

	// Build the SQL and args
	sqlQuery, args := sb.Build()

	// Execute the query
	vaults, err := r.executeVaultQuery(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err // Error already wrapped
	}

	// Generate the token function for the next page
	generateToken := func(vault *Vault) *types.NextPageToken {
		return &types.NextPageToken{
			Column: paginationColumn,
			Value:  vault.ID,
		}
	}

	// Create and return the paginated response
	return types.NewPage(vaults, limit, generateToken), nil
}

// GetByID retrieves a vault by its ID
func (r *repository) GetByID(ctx context.Context, vaultID int64) (*Vault, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"id", "name", "wallet_id", "chain_type", "tx_hash", "recovery_address",
		"signers", "status", "signature_threshold", "address", "recovery_request_timestamp",
		"failure_reason", "created_at", "updated_at", "deleted_at",
	)
	sb.From("vaults")
	sb.Where(sb.Equal("id", vaultID))
	sb.Where(sb.IsNull("deleted_at"))
	sb.Limit(1)

	sqlQuery, args := sb.Build()
	vaults, err := r.executeVaultQuery(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err // Error already wrapped
	}

	if len(vaults) == 0 {
		// Use the correct signature: NewVaultNotFoundError(id int64)
		return nil, errors.NewVaultNotFoundError(vaultID)
	}

	return vaults[0], nil
}

// GetByHash retrieves a vault by its deployment transaction hash
func (r *repository) GetByHash(ctx context.Context, txHash string) (*Vault, error) {
	if txHash == "" {
		return nil, errors.NewValidationError(map[string]any{"tx_hash": "transaction hash cannot be empty"})
	}

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"id", "name", "wallet_id", "chain_type", "tx_hash", "recovery_address",
		"signers", "status", "signature_threshold", "address", "recovery_request_timestamp",
		"failure_reason", "created_at", "updated_at", "deleted_at",
	)
	sb.From("vaults")
	sb.Where(sb.Equal("tx_hash", txHash))
	sb.Where(sb.IsNull("deleted_at"))
	sb.Limit(1)

	sqlQuery, args := sb.Build()
	vaults, err := r.executeVaultQuery(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err // Error already wrapped
	}

	if len(vaults) == 0 {
		// Use the correct signature: NewNotFoundError(message string)
		return nil, errors.NewNotFoundError(fmt.Sprintf("vault not found for tx_hash: %s", txHash))
	}

	return vaults[0], nil
}

// GetByAddress retrieves a vault by its deployed contract address
func (r *repository) GetByAddress(ctx context.Context, address string) (*Vault, error) {
	if address == "" {
		return nil, errors.NewValidationError(map[string]any{"address": "address cannot be empty"})
	}

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"id", "name", "wallet_id", "chain_type", "tx_hash", "recovery_address",
		"signers", "status", "signature_threshold", "address", "recovery_request_timestamp",
		"failure_reason", "created_at", "updated_at", "deleted_at",
	)
	sb.From("vaults")
	sb.Where(sb.Equal("address", address))
	sb.Where(sb.IsNull("deleted_at"))
	sb.Limit(1)

	sqlQuery, args := sb.Build()
	vaults, err := r.executeVaultQuery(ctx, sqlQuery, args...)
	if err != nil {
		return nil, err // Error already wrapped
	}

	if len(vaults) == 0 {
		// Use the correct signature: NewNotFoundError(message string)
		return nil, errors.NewNotFoundError(fmt.Sprintf("vault not found for address: %s", address))
	}

	return vaults[0], nil
}

// UpdateStatus updates a vault's status
func (r *repository) UpdateStatus(ctx context.Context, vaultID int64, status VaultStatus) error {
	if status == "" {
		return errors.NewValidationError(map[string]any{"status": "status cannot be empty"})
	}

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("vaults")
	ub.Set(
		ub.Assign("status", status),
		ub.Assign("updated_at", time.Now().UTC()),
	)
	ub.Where(ub.Equal("id", vaultID))
	ub.Where(ub.IsNull("deleted_at"))

	sqlQuery, args := ub.Build()
	result, err := r.db.ExecuteStatementContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}
	if rowsAffected == 0 {
		// Use the correct signature
		return errors.NewVaultNotFoundError(vaultID)
	}

	return nil
}

// Update updates specific fields of a vault: name, recovery_request_timestamp, failure_reason
func (r *repository) Update(ctx context.Context, vaultID int64, vault *Vault) error {
	if vault == nil {
		return errors.NewValidationError(map[string]any{"vault": "vault data cannot be nil for update"})
	}

	// Prepare assignments for nullable fields
	var recoveryTimestampArg sql.NullTime
	if vault.RecoveryRequestTimestamp != nil {
		recoveryTimestampArg = sql.NullTime{Time: *vault.RecoveryRequestTimestamp, Valid: true}
	}
	var failureReasonArg sql.NullString
	if vault.FailureReason != nil {
		failureReasonArg = sql.NullString{String: *vault.FailureReason, Valid: true}
	}

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("vaults")
	ub.Set(
		ub.Assign("name", vault.Name), // Assume name is always provided for update
		ub.Assign("recovery_request_timestamp", recoveryTimestampArg),
		ub.Assign("failure_reason", failureReasonArg),
		ub.Assign("updated_at", time.Now().UTC()),
	)
	ub.Where(ub.Equal("id", vaultID))
	ub.Where(ub.IsNull("deleted_at"))

	sqlQuery, args := ub.Build()
	result, err := r.db.ExecuteStatementContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}
	if rowsAffected == 0 {
		// Use the correct signature
		return errors.NewVaultNotFoundError(vaultID)
	}

	return nil
}

// Delete marks a vault as deleted
func (r *repository) Delete(ctx context.Context, vaultID int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	now := time.Now().UTC()
	ub.Update("vaults")
	ub.Set(
		ub.Assign("deleted_at", now),
		ub.Assign("updated_at", now),
	)
	ub.Where(ub.Equal("id", vaultID))
	ub.Where(ub.IsNull("deleted_at")) // Only delete if not already deleted

	sqlQuery, args := ub.Build()
	result, err := r.db.ExecuteStatementContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.NewDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewDatabaseError(err)
	}
	if rowsAffected == 0 {
		// Use the correct signature
		return errors.NewVaultNotFoundError(vaultID)
	}

	return nil
}
