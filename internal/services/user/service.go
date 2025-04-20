package user

import (
	"context"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"

	"golang.org/x/crypto/bcrypt"
)

// Service defines the user management operations interface
type Service interface {
	// Create registers a new user with the given email and password
	// Returns ErrEmailExists if the email is already registered
	Create(ctx context.Context, email, password string) (*User, error)

	// Update modifies an existing user's email and/or password
	// Empty parameters are ignored. Returns ErrEmailExists if the new email is already in use
	Update(ctx context.Context, id int64, email, password string) (*User, error)

	// Delete removes a user by ID
	// Returns ErrResourceNotFound if the user doesn't exist
	Delete(ctx context.Context, id int64) error

	// Get retrieves a user by ID
	// Returns ErrResourceNotFound if the user doesn't exist
	Get(ctx context.Context, id int64) (*User, error)

	// List returns a paginated collection of users
	// Default limit is 10 if limit < 1
	// nextToken is used for token-based pagination (empty string for first page)
	List(ctx context.Context, limit int, nextToken string) (*types.Page[*User], error)
}

// service implements the Service interface
type service struct {
	log        logger.Logger
	repository Repository
}

// NewService creates a new user service
func NewService(log logger.Logger, repository Repository) Service {
	return &service{
		log:        log,
		repository: repository,
	}
}

// Create creates a new user
func (s *service) Create(ctx context.Context, email, password string) (*User, error) {
	// Check if email already exists
	existingUser, err := s.repository.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
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
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if email != "" && email != user.Email {
		// Check if new email already exists
		existingUser, err := s.repository.GetByEmail(ctx, email)
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
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// List retrieves a paginated list of users
func (s *service) List(ctx context.Context, limit int, nextToken string) (*types.Page[*User], error) {
	// Set default limit
	if limit <= 0 {
		limit = 10
	}

	// Retrieve users with token-based pagination
	return s.repository.List(ctx, limit, nextToken)
}
