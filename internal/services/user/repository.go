package user

import (
	"context"
	"strconv"
	"time"
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
	List(ctx context.Context, limit, offset int) (*types.Page[*User], error)
}

// repository implements Repository using SQLite database
type repository struct {
	db *db.DB
}

// NewRepository creates a new SQLite user repository
func NewRepository(db *db.DB) Repository {
	return &repository{db: db}
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

	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`

	_, err = r.db.ExecuteStatementContext(ctx, query,
		user.ID, user.Email, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// Update updates an existing user in the database
func (r *repository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET email = ?, password_hash = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecuteStatementContext(ctx, query,
		user.Email, user.PasswordHash, user.UpdatedAt, user.ID)
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
	query := "DELETE FROM users WHERE id = ?"

	result, err := r.db.ExecuteStatementContext(ctx, query, id)
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
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, errors.NewUserNotFoundError()
	}

	var user User
	err = rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// GetByEmail finds a user by email
func (r *repository) GetByEmail(ctx context.Context, email string) (*User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}

	var user User
	err = rows.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// List retrieves a paginated list of users
func (r *repository) List(ctx context.Context, limit, offset int) (*types.Page[*User], error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		ORDER BY id
	`

	args := []any{}

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

	return types.NewPage(users, offset, limit), nil
}
