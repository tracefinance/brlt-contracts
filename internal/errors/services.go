package errors

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

	// Transaction service errors
	ErrCodeTransactionNotFound = "transaction_not_found"
	ErrCodeTransactionFailed   = "transaction_failed"
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
		Message: entity + " not found",
	}
}

// NewAlreadyExistsError creates a generic already exists error
func NewAlreadyExistsError(entity string) *AppError {
	return &AppError{
		Code:    ErrCodeAlreadyExists,
		Message: entity + " already exists",
	}
}

// NewOperationFailedError creates a generic operation failed error
func NewOperationFailedError(operation string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeOperationFailed,
		Message: operation + " operation failed",
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
		Message: "Wallet already exists with address: " + address,
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
		Message: "User already exists with email: " + email,
	}
}

// NewInvalidCredentialsError creates an error for invalid login credentials
func NewInvalidCredentialsError() *AppError {
	return &AppError{
		Code:    ErrCodeInvalidCredentials,
		Message: "Invalid email or password",
	}
}

// NewTransactionNotFoundError creates an error for missing transaction
func NewTransactionNotFoundError() *AppError {
	return &AppError{
		Code:    ErrCodeTransactionNotFound,
		Message: "Transaction not found",
	}
}

// NewTransactionFailedError creates an error for failed transaction
func NewTransactionFailedError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeTransactionFailed,
		Message: "Transaction failed",
		Err:     err,
	}
}
