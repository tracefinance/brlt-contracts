package user

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"vault0/internal/core/db"
)

// Repository defines the user data access interface
type Repository interface {
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, error)
	Count(ctx context.Context) (int, error)
}

// SQLiteRepository implements Repository using SQLite database
type SQLiteRepository struct {
	db *db.DB
}

// NewSQLiteRepository creates a new SQLite user repository
func NewSQLiteRepository(db *db.DB) Repository {
	return &SQLiteRepository{db: db}
}

// Create adds a new user to the database
func (r *SQLiteRepository) Create(ctx context.Context, user *User) error {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	query := `
		INSERT INTO users (username, password_hash, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`

	result, err := r.db.ExecuteStatementContext(ctx, query,
		user.Username, user.PasswordHash, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Get the last insert ID
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	user.ID = id
	return nil
}

// Update updates an existing user in the database
func (r *SQLiteRepository) Update(ctx context.Context, user *User) error {
	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET username = ?, password_hash = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecuteStatementContext(ctx, query,
		user.Username, user.PasswordHash, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", user.ID)
	}

	return nil
}

// Delete removes a user from the database
func (r *SQLiteRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE id = ?"

	result, err := r.db.ExecuteStatementContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user with ID %d not found", id)
	}

	return nil
}

// FindByID finds a user by ID
func (r *SQLiteRepository) FindByID(ctx context.Context, id int64) (*User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by ID: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}

	var user User
	err = rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &user, nil
}

// FindByUsername finds a user by username
func (r *SQLiteRepository) FindByUsername(ctx context.Context, username string) (*User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		WHERE username = ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by username: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, sql.ErrNoRows
	}

	var user User
	err = rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &user, nil
}

// List retrieves a paginated list of users
func (r *SQLiteRepository) List(ctx context.Context, limit, offset int) ([]*User, error) {
	query := `
		SELECT id, username, password_hash, created_at, updated_at
		FROM users
		ORDER BY id
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.ExecuteQueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*User
	for rows.Next() {
		var user User
		err = rows.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return users, nil
}

// Count returns the total number of users
func (r *SQLiteRepository) Count(ctx context.Context) (int, error) {
	query := "SELECT COUNT(*) FROM users"

	rows, err := r.db.ExecuteQueryContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return 0, fmt.Errorf("failed to get count")
	}

	var count int
	if err = rows.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return count, nil
}
