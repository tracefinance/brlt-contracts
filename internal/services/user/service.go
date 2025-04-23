package user

import (
	"context"

	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/signer"
	"vault0/internal/types"

	"golang.org/x/crypto/bcrypt"
)

// Service defines the user management operations interface
type Service interface {
	// CreateUser registers a new user with the given email and password
	// Returns ErrEmailExists if the email is already registered
	CreateUser(ctx context.Context, email, password string) (*User, error)

	// UpdateUser modifies an existing user's email and/or password
	// Empty parameters are ignored. Returns ErrEmailExists if the new email is already in use
	UpdateUser(ctx context.Context, id int64, email, password string) (*User, error)

	// DeleteUser removes a user by ID
	// Returns ErrResourceNotFound if the user doesn't exist
	DeleteUser(ctx context.Context, id int64) error

	// GetUserByID retrieves a user by ID
	// Returns ErrResourceNotFound if the user doesn't exist
	GetUserByID(ctx context.Context, id int64) (*User, error)

	// ListUsers returns a paginated collection of users
	// Default limit is 10 if limit < 1
	// nextToken is used for token-based pagination (empty string for first page)
	ListUsers(ctx context.Context, limit int, nextToken string) (*types.Page[*User], error)
}

// service implements the Service interface
type service struct {
	log           logger.Logger
	repository    Repository
	signerService signer.Service
}

// NewService creates a new user service
func NewService(log logger.Logger, repository Repository, signerSvc signer.Service) Service {
	return &service{
		log:           log,
		repository:    repository,
		signerService: signerSvc,
	}
}

// CreateUser creates a new user
func (s *service) CreateUser(ctx context.Context, email, password string) (*User, error) {
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

// UpdateUser updates an existing user
func (s *service) UpdateUser(ctx context.Context, id int64, email, password string) (*User, error) {
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

// DeleteUser removes a user
func (s *service) DeleteUser(ctx context.Context, id int64) error {
	// Check if user is associated with any signers before deleting
	signers, err := s.signerService.FindSignersByUserID(ctx, id)
	if err != nil {
		s.log.Error("Failed to check signer association for user",
			logger.Int64("user_id", id),
			logger.Error(err))
		return err
	}

	// If signers exist for the user, prevent deletion
	if len(signers) > 0 {
		s.log.Warn("Attempted to delete user associated with signers",
			logger.Int64("user_id", id),
			logger.Int("signer_count", len(signers)))
		return errors.NewUserAssociatedWithSignerError(id)
	}

	// Proceed with deletion if no signers are associated
	if err := s.repository.Delete(ctx, id); err != nil {
		s.log.Error("Failed to delete user from repository",
			logger.Int64("user_id", id),
			logger.Error(err))
		return err
	}

	s.log.Info("User deleted successfully", logger.Int64("user_id", id))
	return nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id int64) (*User, error) {
	user, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// ListUsers retrieves a paginated list of users
func (s *service) ListUsers(ctx context.Context, limit int, nextToken string) (*types.Page[*User], error) {
	// Set default limit
	if limit <= 0 {
		limit = 10
	}

	// Retrieve users with token-based pagination
	return s.repository.List(ctx, limit, nextToken)
}
