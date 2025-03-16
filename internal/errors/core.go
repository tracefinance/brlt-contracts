package errors

import (
	"fmt"
	"math/big"
)

// Core module error codes
const (
	// Database errors
	ErrCodeDatabaseError    = "database_error"
	ErrCodeDatabaseNotFound = "database_not_found"

	// Resource errors
	ErrCodeResourceNotFound = "resource_not_found"
	ErrCodeResourceExists   = "resource_exists"

	// Token errors
	ErrCodeInvalidToken = "invalid_token"

	// Blockchain errors
	ErrCodeBlockchainError         = "blockchain_error"
	ErrCodeInvalidBlockchainConfig = "invalid_blockchain_config"
	ErrCodeChainNotSupported       = "chain_not_supported"
	ErrCodeInsufficientFunds       = "insufficient_funds"
	ErrCodeInvalidTransaction      = "invalid_transaction"
	ErrCodeTransactionNotFound     = "transaction_not_found"
	ErrCodeRPCError                = "rpc_error"
	ErrCodeInvalidAddress          = "invalid_address"
	ErrCodeTransactionFailed       = "transaction_failed"
	ErrCodeInvalidContract         = "invalid_contract"

	// Keystore errors
	ErrCodeKeystoreError   = "keystore_error"
	ErrCodeKeyNotFound     = "key_not_found"
	ErrCodeKeyExists       = "key_exists"
	ErrCodeInvalidKey      = "invalid_key"
	ErrCodeSigningError    = "signing_error"
	ErrCodeInvalidKeystore = "invalid_keystore"

	// Wallet errors
	ErrCodeWalletError         = "wallet_error"
	ErrCodeInvalidWalletConfig = "invalid_wallet_config"
	ErrCodeInvalidKeyType      = "invalid_key_type"
	ErrCodeInvalidCurve        = "invalid_curve"
	ErrCodeInvalidSignature    = "invalid_signature"
	ErrCodeSignatureRecovery   = "signature_recovery_failed"
	ErrCodeAddressMismatch     = "address_mismatch"

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

	// Blockchain transaction errors
	ErrCodeInvalidNonce        = "invalid_nonce"
	ErrCodeInvalidGasPrice     = "invalid_gas_price"
	ErrCodeInvalidGasLimit     = "invalid_gas_limit"
	ErrCodeInvalidContractCall = "invalid_contract_call"
	ErrCodeInvalidAmount       = "invalid_amount"
)

// NewInvalidBlockchainConfigError creates an error for invalid blockchain configuration
func NewInvalidBlockchainConfigError(chain string, key string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidBlockchainConfig,
		Message: fmt.Sprintf("Invalid blockchain configuration for %s: missing %s", chain, key),
		Details: map[string]any{
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
		Details: map[string]any{
			"error": err.Error(),
		},
	}
}

// NewChainNotSupportedError creates a new error for unsupported chains
func NewChainNotSupportedError(chain string) *AppError {
	return &AppError{
		Code:    ErrCodeChainNotSupported,
		Message: fmt.Sprintf("Chain not supported: %s", chain),
		Details: map[string]any{
			"chain": chain,
		},
	}
}

// NewInsufficientFundsError creates a new error for insufficient funds
func NewInsufficientFundsError(balance string, required string) *AppError {
	return &AppError{
		Code:    ErrCodeInsufficientFunds,
		Message: fmt.Sprintf("Insufficient funds: have %s, need %s", balance, required),
		Details: map[string]any{
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

// NewInvalidKeyError creates an error for invalid keys
func NewInvalidKeyError(msg string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidKey,
		Message: msg,
		Err:     err,
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
		Err:     err,
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
		Details: map[string]any{
			"address": address,
		},
	}
}

// NewInvalidAmountError creates a new error for invalid amounts
func NewInvalidAmountError(amount string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidAmount,
		Message: fmt.Sprintf("Invalid amount: %s", amount),
		Details: map[string]any{
			"amount": amount,
		},
	}
}

// NewRPCError creates a new error for RPC-related issues
func NewRPCError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeRPCError,
		Message: fmt.Sprintf("RPC error: %v", err),
		Err:     err,
	}
}

// NewInvalidTransactionError creates a new error for invalid transactions
func NewInvalidTransactionError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidTransaction,
		Message: fmt.Sprintf("Invalid transaction: %v", err),
		Err:     err,
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
		Err:     err,
	}
}

// NewExplorerRequestFailedError creates a new error for failed block explorer requests
func NewExplorerRequestFailedError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeExplorerRequestFailed,
		Message: fmt.Sprintf("Block explorer request failed: %v", err),
		Err:     err,
	}
}

// NewMissingAPIKeyError creates a new error for missing API key
func NewMissingAPIKeyError() *AppError {
	return &AppError{
		Code:    ErrCodeMissingAPIKey,
		Message: "Missing API key for block explorer",
	}
}

// NewInvalidNonceError creates a new error for invalid nonce
func NewInvalidNonceError(address string, nonce uint64) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidNonce,
		Message: fmt.Sprintf("Invalid nonce for address %s: %d", address, nonce),
		Details: map[string]any{
			"address": address,
			"nonce":   nonce,
		},
	}
}

// NewInvalidGasPriceError creates a new error for invalid gas price
func NewInvalidGasPriceError(gasPrice *big.Int) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidGasPrice,
		Message: fmt.Sprintf("Invalid gas price: %s", gasPrice.String()),
		Details: map[string]any{
			"gas_price": gasPrice.String(),
		},
	}
}

// NewInvalidGasLimitError creates a new error for invalid gas limit
func NewInvalidGasLimitError(gasLimit uint64) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidGasLimit,
		Message: fmt.Sprintf("Invalid gas limit: %d", gasLimit),
		Details: map[string]any{
			"gas_limit": gasLimit,
		},
	}
}

// NewInvalidContractCallError creates a new error for invalid contract calls
func NewInvalidContractCallError(contract string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidContractCall,
		Message: fmt.Sprintf("Invalid contract call to %s: %v", contract, err),
		Err:     err,
		Details: map[string]any{
			"contract": contract,
			"error":    err.Error(),
		},
	}
}

// NewInvalidContractError creates a new error for invalid contracts
func NewInvalidContractError(contract string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidContract,
		Message: fmt.Sprintf("Invalid contract: %s", contract),
		Err:     err,
		Details: map[string]any{
			"contract": contract,
			"error":    err.Error(),
		},
	}
}

// NewWalletError creates a new error for general wallet operations
func NewWalletError(msg string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeWalletError,
		Message: fmt.Sprintf("Wallet error: %s", msg),
		Err:     err,
		Details: map[string]any{
			"error": err.Error(),
		},
	}
}

// NewInvalidWalletConfigError creates a new error for invalid wallet configuration
func NewInvalidWalletConfigError(msg string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidWalletConfig,
		Message: fmt.Sprintf("Invalid wallet configuration: %s", msg),
	}
}

// NewInvalidKeyTypeError creates a new error for invalid key types
func NewInvalidKeyTypeError(expected, got string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidKeyType,
		Message: fmt.Sprintf("Invalid key type: expected %s, got %s", expected, got),
		Details: map[string]any{
			"expected": expected,
			"got":      got,
		},
	}
}

// NewInvalidCurveError creates a new error for invalid curves
func NewInvalidCurveError(expected, got string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidCurve,
		Message: fmt.Sprintf("Invalid curve: expected %s, got %s", expected, got),
		Details: map[string]any{
			"expected": expected,
			"got":      got,
		},
	}
}

// NewInvalidSignatureError creates a new error for invalid signatures
func NewInvalidSignatureError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidSignature,
		Message: "Invalid signature",
		Err:     err,
		Details: map[string]any{
			"error": err.Error(),
		},
	}
}

// NewSignatureRecoveryError creates a new error for signature recovery failures
func NewSignatureRecoveryError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeSignatureRecovery,
		Message: "Failed to recover signature",
		Err:     err,
	}
}

// NewAddressMismatchError creates a new error for address mismatches
func NewAddressMismatchError(expected, got string) *AppError {
	return &AppError{
		Code:    ErrCodeAddressMismatch,
		Message: fmt.Sprintf("Address mismatch: expected %s, got %s", expected, got),
		Details: map[string]any{
			"expected": expected,
			"got":      got,
		},
	}
}

// NewTransactionNotFoundError creates a new error for transaction not found
func NewTransactionNotFoundError(hash string) *AppError {
	return &AppError{
		Code:    ErrCodeTransactionNotFound,
		Message: fmt.Sprintf("Transaction not found: %s", hash),
		Details: map[string]any{
			"hash": hash,
		},
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

// NewInvalidEncryptionKeyError creates a new error for invalid encryption keys
func NewInvalidEncryptionKeyError(key string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidEncryptionKey,
		Message: fmt.Sprintf("Invalid encryption key: %s", key),
	}
}

// NewResourceNotFoundError creates an error for when a resource is not found
func NewResourceNotFoundError(resource, id string) *AppError {
	return &AppError{
		Code:    ErrCodeResourceNotFound,
		Message: fmt.Sprintf("%s not found: %s", resource, id),
		Details: map[string]any{
			"resource": resource,
			"id":       id,
		},
	}
}

// NewResourceAlreadyExistsError creates an error for when a resource already exists
func NewResourceAlreadyExistsError(resource, attribute, value string) *AppError {
	return &AppError{
		Code:    ErrCodeResourceExists,
		Message: fmt.Sprintf("%s with %s '%s' already exists", resource, attribute, value),
		Details: map[string]any{
			"resource":  resource,
			"attribute": attribute,
			"value":     value,
		},
	}
}

// NewSigningError creates an error for signing operations
func NewSigningError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeSigningError,
		Message: "Signing operation failed",
		Err:     err,
		Details: map[string]any{
			"error": err.Error(),
		},
	}
}

// NewInvalidKeystoreError creates an error for invalid or uninitialized keystore
func NewInvalidKeystoreError(keystore string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidKeystore,
		Message: fmt.Sprintf("Invalid keystore: %s", keystore),
	}
}

// NewInvalidTokenError creates an error for invalid token data
func NewInvalidTokenError(msg string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidToken,
		Message: fmt.Sprintf("Invalid token: %s", msg),
		Err:     err,
	}
}
