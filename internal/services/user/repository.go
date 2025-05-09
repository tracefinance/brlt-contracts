package user

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/huandu/go-sqlbuilder"

	"vault0/internal/db"
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Repository defines the user data access interface
type Repository interface {
	// Create adds a new user to the database
	Create(ctx context.Context, user *User) error

	// Update modifies an existing user's information
	Update(ctx context.Context, user *User) error

	// Delete removes a user from the database
	Delete(ctx context.Context, id int64) error

	// GetByID retrieves a user by their unique ID
	GetByID(ctx context.Context, id int64) (*User, error)

	// GetByEmail retrieves a user by their email address
	GetByEmail(ctx context.Context, email string) (*User, error)

	// List retrieves a paginated collection of users
	// When limit=0, returns all users without pagination
	// nextToken is used for token-based pagination
	List(ctx context.Context, limit int, nextToken string) (*types.Page[*User], error)
}

// repository implements Repository using SQLite database
type repository struct {
	db            *db.DB
	userStructMap *sqlbuilder.Struct
}

// NewRepository creates a new SQLite user repository
func NewRepository(db *db.DB) Repository {
	userStructMap := sqlbuilder.NewStruct(new(User))

	return &repository{
		db:            db,
		userStructMap: userStructMap,
	}
}

// executeUserQuery executes a query and scans the results into User objects
func (r *repository) executeUserQuery(ctx context.Context, sql string, args ...interface{}) ([]*User, error) {
	rows, err := r.db.ExecuteQueryContext(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// Create adds a new user to the database
func (r *repository) Create(ctx context.Context, user *User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Generate a Snowflake ID
	id, err := r.db.GenerateID()
	if err != nil {
		return err
	}
	user.ID = id

	// Create a struct-based insert builder
	ib := r.userStructMap.InsertInto("users", user)

	// Build the SQL and args
	sql, args := ib.Build()

	// Execute the insert
	_, err = r.db.ExecuteStatementContext(ctx, sql, args...)
	return err
}

// Update updates an existing user in the database
func (r *repository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	// Create a struct-based update builder
	ub := r.userStructMap.Update("users", user)
	ub.Where(ub.Equal("id", user.ID))

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
		return errors.NewResourceNotFoundError("User", strconv.FormatInt(user.ID, 10))
	}

	return nil
}

// Delete removes a user from the database
func (r *repository) Delete(ctx context.Context, id int64) error {
	// Create a struct-based delete builder
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom("users")
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
		return errors.NewResourceNotFoundError("User", strconv.FormatInt(id, 10))
	}

	return nil
}

// GetByID finds a user by ID
func (r *repository) GetByID(ctx context.Context, id int64) (*User, error) {
	// Create a struct-based select builder
	sb := r.userStructMap.SelectFrom("users")
	sb.Where(sb.Equal("id", id))

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	users, err := r.executeUserQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, errors.NewUserNotFoundError()
	}

	return users[0], nil
}

// GetByEmail finds a user by email
func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	// Create a struct-based select builder
	sb := r.userStructMap.SelectFrom("users")
	sb.Where(sb.Equal("email", email))

	// Build the SQL and args
	sql, args := sb.Build()

	// Execute the query
	users, err := r.executeUserQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, nil
	}

	return users[0], nil
}

// List retrieves a paginated list of users using token-based pagination
func (r *repository) List(ctx context.Context, limit int, nextToken string) (*types.Page[*User], error) {
	// Create a struct-based select builder
	sb := r.userStructMap.SelectFrom("users")

	// Default sort by id
	paginationColumn := "id"

	// If nextToken is provided, decode it to get the starting point
	token, err := types.DecodeNextPageToken(nextToken, paginationColumn)
	if err != nil {
		return nil, err
	}

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
	users, err := r.executeUserQuery(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	// Generate the token function
	generateToken := func(user *User) *types.NextPageToken {
		return &types.NextPageToken{
			Column: paginationColumn,
			Value:  user.ID,
		}
	}

	return types.NewPage(users, limit, generateToken), nil
}
