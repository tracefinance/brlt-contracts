package user

import (
	"context"
	"database/sql"

	"vault0/internal/errors"

	"golang.org/x/crypto/bcrypt"
)

// Service defines the user service interface
type Service interface {
	Create(ctx context.Context, email, password string) (*User, error)
	Update(ctx context.Context, id int64, email, password string) (*User, error)
	Delete(ctx context.Context, id int64) error
	Get(ctx context.Context, id int64) (*User, error)
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
func (s *service) Create(ctx context.Context, email, password string) (*User, error) {
	// Check if email already exists
	existingUser, err := s.repository.FindByEmail(ctx, email)
	if err != nil && err != sql.ErrNoRows {
		return nil, errors.NewDatabaseError(err)
	}
	if existingUser != nil {
		return nil, errors.NewEmailExistsError(email)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.NewOperationFailedError("password hashing", err)
	}

	// Create the user
	user := &User{
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	// Save the user
	if err := s.repository.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Update updates an existing user
func (s *service) Update(ctx context.Context, id int64, email, password string) (*User, error) {
	// Get the existing user
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if email != "" && email != user.Email {
		// Check if new email already exists
		existingUser, err := s.repository.FindByEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		if existingUser != nil {
			return nil, errors.NewEmailExistsError(email)
		}

		user.Email = email
	}

	// Update password if provided
	if password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.NewOperationFailedError("password hashing", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	// Save the updated user
	if err := s.repository.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Delete removes a user
func (s *service) Delete(ctx context.Context, id int64) error {
	if err := s.repository.Delete(ctx, id); err != nil {
		return err
	}
	return nil
}

// Get retrieves a user by ID
func (s *service) Get(ctx context.Context, id int64) (*User, error) {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return nil, err
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
		return nil, 0, err
	}

	// Get the total count
	total, err := s.repository.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}
