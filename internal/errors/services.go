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
	ErrCodeMissingKeyID          = "missing_key_id"
	ErrCodeMissingWalletAddress  = "missing_wallet_address"

	// User service errors
	ErrCodeUserNotFound       = "user_not_found"
	ErrCodeUserExists         = "user_exists"
	ErrCodeInvalidCredentials = "invalid_credentials"
	ErrCodeEmailExists        = "email_exists"

	// Transaction service errors
	ErrCodeTransactionSyncFailed = "transaction_sync_failed"
)

// NewInvalidInputError creates an error for invalid input data with a custom message
func NewInvalidInputError(message string, field string, value any) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidInput,
		Message: message,
		Details: map[string]any{
			"field":   field,
			"value":   value,
			"message": message,
		},
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
func NewWalletNotFoundError(address string) *AppError {
	return &AppError{
		Code:    ErrCodeWalletNotFound,
		Message: fmt.Sprintf("Wallet not found for address: %s", address),
		Details: map[string]any{
			"address": address,
		},
	}
}

// NewWalletExistsError creates an error for duplicate wallet
func NewWalletExistsError(address string) *AppError {
	return &AppError{
		Code:    ErrCodeWalletExists,
		Message: fmt.Sprintf("Wallet already exists for address: %s", address),
		Details: map[string]any{
			"address": address,
		},
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

// NewMissingKeyIDError creates an error for when a wallet is missing a key ID
func NewMissingKeyIDError() *AppError {
	return &AppError{
		Code:    ErrCodeMissingKeyID,
		Message: "Internal wallet requires a key ID",
	}
}

// NewMissingWalletAddressError creates an error for when a wallet is missing an address
func NewMissingWalletAddressError() *AppError {
	return &AppError{
		Code:    ErrCodeMissingWalletAddress,
		Message: "External wallet requires an address",
	}
}

// NewWalletOperationFailedError creates an error for wallet operation failures
func NewWalletOperationFailedError(operation string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeWalletOperationFailed,
		Message: fmt.Sprintf("Wallet %s operation failed", operation),
		Err:     err,
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

// NewTransactionSyncFailedError creates an error for transaction sync failure
func NewTransactionSyncFailedError(operation string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeTransactionSyncFailed,
		Message: fmt.Sprintf("Transaction sync failed: %s", operation),
		Err:     err,
	}
}
