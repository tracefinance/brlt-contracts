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

	// Signer service errors
	ErrCodeSignerNotFound        = "signer_not_found"
	ErrCodeSignerAddressNotFound = "signer_address_not_found"

	// Price Feed Service Errors
	ErrCodeTokenPriceNotFound    = "token_price_not_found"
	ErrCodePriceFeedUpdateFailed = "price_feed_update_failed"
	ErrCodeDataConversionFailed  = "data_conversion_failed"
)

// NewInvalidInputError creates an error for invalid input data with a custom message
func NewInvalidInputError(message string, field string, value any) *Vault0Error {
	return &Vault0Error{
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
func NewNotFoundError(entity string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", entity),
	}
}

// NewAlreadyExistsError creates a generic already exists error
func NewAlreadyExistsError(entity string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeAlreadyExists,
		Message: fmt.Sprintf("%s already exists", entity),
	}
}

// NewOperationFailedError creates a generic operation failed error
func NewOperationFailedError(operation string, err error) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeOperationFailed,
		Message: fmt.Sprintf("%s operation failed", operation),
		Err:     err,
	}
}

// NewWalletNotFoundError creates an error for missing wallet
func NewWalletNotFoundError(address string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeWalletNotFound,
		Message: fmt.Sprintf("Wallet not found for address: %s", address),
		Details: map[string]any{
			"address": address,
		},
	}
}

// NewWalletExistsError creates an error for duplicate wallet
func NewWalletExistsError(address string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeWalletExists,
		Message: fmt.Sprintf("Wallet already exists for address: %s", address),
		Details: map[string]any{
			"address": address,
		},
	}
}

// NewInvalidWalletError creates an error for invalid wallet data
func NewInvalidWalletError(details map[string]any) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeInvalidWallet,
		Message: "Invalid wallet data",
		Details: details,
	}
}

// NewMissingKeyIDError creates an error for when a wallet is missing a key ID
func NewMissingKeyIDError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeMissingKeyID,
		Message: "Internal wallet requires a key ID",
	}
}

// NewMissingWalletAddressError creates an error for when a wallet is missing an address
func NewMissingWalletAddressError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeMissingWalletAddress,
		Message: "External wallet requires an address",
	}
}

// NewWalletOperationFailedError creates an error for wallet operation failures
func NewWalletOperationFailedError(operation string, err error) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeWalletOperationFailed,
		Message: fmt.Sprintf("Wallet %s operation failed", operation),
		Err:     err,
	}
}

// NewUserNotFoundError creates an error for missing user
func NewUserNotFoundError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeUserNotFound,
		Message: "User not found",
	}
}

// NewUserExistsError creates an error for duplicate user
func NewUserExistsError(email string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeUserExists,
		Message: fmt.Sprintf("User already exists with email: %s", email),
	}
}

// NewInvalidCredentialsError creates an error for invalid login credentials
func NewInvalidCredentialsError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeInvalidCredentials,
		Message: "Invalid email or password",
	}
}

// NewEmailExistsError creates an error for when an email is already registered
func NewEmailExistsError(email string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeEmailExists,
		Message: fmt.Sprintf("Email already exists: %s", email),
		Details: map[string]any{
			"email": email,
		},
	}
}

// NewTransactionSyncFailedError creates an error for transaction sync failure
func NewTransactionSyncFailedError(operation string, err error) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeTransactionSyncFailed,
		Message: fmt.Sprintf("Transaction sync failed: %s", operation),
		Err:     err,
	}
}

// NewSignerNotFoundError creates an error for missing signer
func NewSignerNotFoundError(id int64) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeSignerNotFound,
		Message: "Signer not found",
		Details: map[string]any{
			"signer_id": id,
		},
	}
}

// NewSignerAddressNotFoundError creates an error for missing signer address
func NewSignerAddressNotFoundError(id int64) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeSignerAddressNotFound,
		Message: "Signer address not found",
		Details: map[string]any{
			"address_id": id,
		},
	}
}

// --- Token Price Service Errors ---

// NewTokenPriceNotFoundError creates an error for when token price data is not found.
func NewTokenPriceNotFoundError(symbol string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeTokenPriceNotFound,
		Message: fmt.Sprintf("Token price data not found for symbol: %s", symbol),
		Err:     nil,
	}
}

// NewPriceFeedUpdateFailed creates an error for failures during the price feed update process.
func NewPriceFeedUpdateFailed(err error, reason string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodePriceFeedUpdateFailed,
		Message: fmt.Sprintf("Failed to update token prices from feed: %s", reason),
		Err:     err,
	}
}

// NewDataConversionFailed creates an error for failures during data conversion.
func NewDataConversionFailed(err error, context string, details map[string]any) *Vault0Error {
	msg := "Data conversion failed"
	if context != "" {
		msg = fmt.Sprintf("Data conversion failed: %s", context)
	}
	return &Vault0Error{
		Code:    ErrCodeDataConversionFailed,
		Message: msg,
		Details: details,
		Err:     err,
	}
}
