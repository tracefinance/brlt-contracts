package errors

import "fmt"

// Service module error codes
const (
	// Common service errors
	ErrCodeInvalidInput    = "invalid_input"
	ErrCodeNotFound        = "not_found"
	ErrCodeAlreadyExists   = "already_exists"
	ErrCodeOperationFailed = "operation_failed"

	// Wallet service errors
	ErrCodeWalletNotFound        = "wallet_not_found"
	ErrCodeWalletExists          = "wallet_exists"
	ErrCodeInvalidWallet         = "invalid_wallet"
	ErrCodeWalletOperationFailed = "wallet_operation_failed"

	// User service errors
	ErrCodeUserNotFound       = "user_not_found"
	ErrCodeUserExists         = "user_exists"
	ErrCodeInvalidCredentials = "invalid_credentials"
	ErrCodeEmailExists        = "email_exists"
)

// NewInvalidInputError creates an error for invalid input data
func NewInvalidInputError(details map[string]any) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidInput,
		Message: "Invalid input data",
		Details: details,
	}
}

// NewNotFoundError creates a generic not found error
func NewNotFoundError(entity string) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", entity),
	}
}

// NewAlreadyExistsError creates a generic already exists error
func NewAlreadyExistsError(entity string) *AppError {
	return &AppError{
		Code:    ErrCodeAlreadyExists,
		Message: fmt.Sprintf("%s already exists", entity),
	}
}

// NewOperationFailedError creates a generic operation failed error
func NewOperationFailedError(operation string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeOperationFailed,
		Message: fmt.Sprintf("%s operation failed", operation),
		Err:     err,
	}
}

// NewWalletNotFoundError creates an error for missing wallet
func NewWalletNotFoundError() *AppError {
	return &AppError{
		Code:    ErrCodeWalletNotFound,
		Message: "Wallet not found",
	}
}

// NewWalletExistsError creates an error for duplicate wallet
func NewWalletExistsError(address string) *AppError {
	return &AppError{
		Code:    ErrCodeWalletExists,
		Message: fmt.Sprintf("Wallet already exists with address: %s", address),
	}
}

// NewInvalidWalletError creates an error for invalid wallet data
func NewInvalidWalletError(details map[string]any) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidWallet,
		Message: "Invalid wallet data",
		Details: details,
	}
}

// NewUserNotFoundError creates an error for missing user
func NewUserNotFoundError() *AppError {
	return &AppError{
		Code:    ErrCodeUserNotFound,
		Message: "User not found",
	}
}

// NewUserExistsError creates an error for duplicate user
func NewUserExistsError(email string) *AppError {
	return &AppError{
		Code:    ErrCodeUserExists,
		Message: fmt.Sprintf("User already exists with email: %s", email),
	}
}

// NewInvalidCredentialsError creates an error for invalid login credentials
func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Code:    ErrCodeInvalidCredentials,
		Message: "Invalid email or password",
	}
}

// NewEmailExistsError creates an error for when an email is already registered
func NewEmailExistsError(email string) *AppError {
	return &AppError{
		Code:    ErrCodeEmailExists,
		Message: fmt.Sprintf("Email already exists: %s", email),
		Details: map[string]any{
			"email": email,
		},
	}
}
