package user

import (
	"context"
	"database/sql"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Service defines the user service interface
type Service interface {
	Create(ctx context.Context, username, password string) (*User, error)
	Update(ctx context.Context, id int64, username, password string) (*User, error)
	Delete(ctx context.Context, id int64) error
	GetByID(ctx context.Context, id int64) (*User, error)
	List(ctx context.Context, page, pageSize int) ([]*User, int, error)
}

// service implements the Service interface
type service struct {
	repository Repository
}

// NewService creates a new user service
func NewService(repository Repository) Service {
	return &service{
		repository: repository,
	}
}

// Create creates a new user
func (s *service) Create(ctx context.Context, username, password string) (*User, error) {
	// Check if username already exists
	existingUser, err := s.repository.FindByUsername(ctx, username)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user
	user := &User{
		Username:     username,
		PasswordHash: string(hashedPassword),
	}

	// Save the user
	if err := s.repository.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Update updates an existing user
func (s *service) Update(ctx context.Context, id int64, username, password string) (*User, error) {
	// Get the existing user
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Update fields if provided
	if username != "" && username != user.Username {
		// Check if new username already exists
		existingUser, err := s.repository.FindByUsername(ctx, username)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to check username: %w", err)
		}
		if existingUser != nil {
			return nil, fmt.Errorf("username already exists")
		}

		user.Username = username
	}

	// Update password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	// Save the updated user
	if err := s.repository.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// Delete removes a user
func (s *service) Delete(ctx context.Context, id int64) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (s *service) GetByID(ctx context.Context, id int64) (*User, error) {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// List retrieves a paginated list of users
func (s *service) List(ctx context.Context, page, pageSize int) ([]*User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Get the users
	users, err := s.repository.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}

	// Get the total count
	total, err := s.repository.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	return users, total, nil
}
