package errors

import "fmt"

// Core module error codes
const (
	// Configuration errors
	ErrCodeInvalidBlockchainConfig = "invalid_blockchain_config"

	// Database errors
	ErrCodeDatabaseError    = "database_error"
	ErrCodeDatabaseNotFound = "database_not_found"

	// Blockchain errors
	ErrCodeBlockchainError    = "blockchain_error"
	ErrCodeChainNotSupported  = "chain_not_supported"
	ErrCodeInsufficientFunds  = "insufficient_funds"
	ErrCodeInvalidTransaction = "invalid_transaction"
	ErrCodeRPCError           = "rpc_error"
	ErrCodeInvalidAddress     = "invalid_address"

	// Keystore errors
	ErrCodeKeystoreError = "keystore_error"
	ErrCodeKeyNotFound   = "key_not_found"
	ErrCodeKeyExists     = "key_exists"
	ErrCodeInvalidKey    = "invalid_key"

	// Crypto errors
	ErrCodeCryptoError          = "crypto_error"
	ErrCodeEncryptionError      = "encryption_error"
	ErrCodeDecryptionError      = "decryption_error"
	ErrCodeInvalidEncryptionKey = "invalid_encryption_key"

	// Block explorer errors
	ErrCodeExplorerError           = "explorer_error"
	ErrCodeRateLimitExceeded       = "rate_limit_exceeded"
	ErrCodeInvalidAPIKey           = "invalid_api_key"
	ErrCodeInvalidExplorerResponse = "invalid_explorer_response"
	ErrCodeExplorerRequestFailed   = "explorer_request_failed"
	ErrCodeMissingAPIKey           = "missing_api_key"

	// Transaction error codes
	ErrCodeInvalidAmount = "invalid_amount"

	// New error code
	ErrCodeRPC = "RPC_ERROR"
)

// NewInvalidBlockchainConfigError creates an error for invalid blockchain configuration
func NewInvalidBlockchainConfigError(chain string, key string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidBlockchainConfig,
		Message: fmt.Sprintf("Invalid blockchain configuration for %s: missing %s", chain, key),
		Details: map[string]interface{}{
			"chain": chain,
			"key":   key,
		},
	}
}

// NewDatabaseError creates an error for database failures
func NewDatabaseError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeDatabaseError,
		Message: "Database operation failed",
		Err:     err,
	}
}

// NewDatabaseNotFoundError creates an error for database record not found
func NewDatabaseNotFoundError(entity string) *AppError {
	return &AppError{
		Code:    ErrCodeDatabaseNotFound,
		Message: entity + " not found",
	}
}

// NewBlockchainError creates a new error for blockchain-related issues
func NewBlockchainError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeBlockchainError,
		Message: fmt.Sprintf("Blockchain error: %v", err),
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// NewChainNotSupportedError creates a new error for unsupported chains
func NewChainNotSupportedError(chain string) *AppError {
	return &AppError{
		Code:    ErrCodeChainNotSupported,
		Message: fmt.Sprintf("Chain not supported: %s", chain),
		Details: map[string]interface{}{
			"chain": chain,
		},
	}
}

// NewInsufficientFundsError creates a new error for insufficient funds
func NewInsufficientFundsError(balance string, required string) *AppError {
	return &AppError{
		Code:    ErrCodeInsufficientFunds,
		Message: fmt.Sprintf("Insufficient funds: have %s, need %s", balance, required),
		Details: map[string]interface{}{
			"balance":  balance,
			"required": required,
		},
	}
}

// NewKeystoreError creates an error for keystore operations
func NewKeystoreError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeKeystoreError,
		Message: "Keystore operation failed",
		Err:     err,
	}
}

// NewKeyNotFoundError creates an error for missing keys
func NewKeyNotFoundError(keyID string) *AppError {
	return &AppError{
		Code:    ErrCodeKeyNotFound,
		Message: "Key not found: " + keyID,
	}
}

// NewCryptoError creates an error for cryptographic operations
func NewCryptoError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeCryptoError,
		Message: "Cryptographic operation failed",
		Err:     err,
	}
}

// NewEncryptionError creates an error for encryption failures
func NewEncryptionError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeEncryptionError,
		Message: "Encryption failed",
		Err:     err,
	}
}

// NewDecryptionError creates an error for decryption failures
func NewDecryptionError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeDecryptionError,
		Message: "Decryption failed",
		Err:     err,
	}
}

// NewExplorerError creates a new error for block explorer issues
func NewExplorerError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeExplorerError,
		Message: fmt.Sprintf("Block explorer error: %v", err),
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// NewRateLimitExceededError creates a new error for rate limit issues
func NewRateLimitExceededError() *AppError {
	return &AppError{
		Code:    ErrCodeRateLimitExceeded,
		Message: "Rate limit exceeded",
	}
}

// NewInvalidAddressError creates a new error for invalid addresses
func NewInvalidAddressError(address string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidAddress,
		Message: fmt.Sprintf("Invalid address: %s", address),
		Details: map[string]interface{}{
			"address": address,
		},
	}
}

// NewInvalidAmountError creates a new error for invalid amounts
func NewInvalidAmountError(amount string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidAmount,
		Message: fmt.Sprintf("Invalid amount: %s", amount),
		Details: map[string]interface{}{
			"amount": amount,
		},
	}
}

// NewRPCError creates a new error for RPC-related issues
func NewRPCError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeRPCError,
		Message: fmt.Sprintf("RPC error: %v", err),
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// NewInvalidTransactionError creates a new error for invalid transactions
func NewInvalidTransactionError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidTransaction,
		Message: fmt.Sprintf("Invalid transaction: %v", err),
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// NewInvalidAPIKeyError creates a new error for invalid API keys
func NewInvalidAPIKeyError() *AppError {
	return &AppError{
		Code:    ErrCodeInvalidAPIKey,
		Message: "Invalid API key",
	}
}

// NewInvalidExplorerResponseError creates a new error for invalid responses from block explorer
func NewInvalidExplorerResponseError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidExplorerResponse,
		Message: fmt.Sprintf("Invalid response from block explorer: %v", err),
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// NewExplorerRequestFailedError creates a new error for failed block explorer requests
func NewExplorerRequestFailedError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeExplorerRequestFailed,
		Message: fmt.Sprintf("Block explorer request failed: %v", err),
		Details: map[string]interface{}{
			"error": err.Error(),
		},
	}
}

// NewMissingAPIKeyError creates a new error for missing API key
func NewMissingAPIKeyError() *AppError {
	return &AppError{
		Code:    ErrCodeMissingAPIKey,
		Message: "Missing API key for block explorer",
	}
}
