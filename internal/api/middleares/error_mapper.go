package middleares

import (
	"net/http"
	"vault0/internal/errors"
)

// DefaultErrorMapper provides fallback behavior for mapping errors to HTTP responses
func DefaultErrorMapper(err error) (int, any) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Code {
		// Input validation errors - 400 Bad Request
		case errors.ErrCodeInvalidInput:
		case errors.ErrCodeValidationError:
		case errors.ErrCodeInvalidRequest:
		case errors.ErrCodeMissingParameter:
		case errors.ErrCodeInvalidParameter:
		case errors.ErrCodeInvalidWallet:
		case errors.ErrCodeInvalidWalletConfig:
		case errors.ErrCodeInvalidKey:
		case errors.ErrCodeInvalidKeyType:
		case errors.ErrCodeInvalidCurve:
		case errors.ErrCodeInvalidSignature:
		case errors.ErrCodeInvalidAddress:
		case errors.ErrCodeInvalidTransaction:
		case errors.ErrCodeInvalidNonce:
		case errors.ErrCodeInvalidGasPrice:
		case errors.ErrCodeInvalidGasLimit:
		case errors.ErrCodeInvalidContractCall:
		case errors.ErrCodeInvalidAmount:
		case errors.ErrCodeInvalidContract:
		case errors.ErrCodeInvalidBlockchainConfig:
		case errors.ErrCodeInvalidAPIKey:
		case errors.ErrCodeInvalidExplorerResponse:
		case errors.ErrCodeInvalidEncryptionKey:
		case errors.ErrCodeInvalidKeystore:
		case errors.ErrCodeInvalidToken:
		case errors.ErrCodeMissingKeyID:
		case errors.ErrCodeMissingWalletAddress:
		case errors.ErrCodeMissingAPIKey:
			return http.StatusBadRequest, appErr

		// Authentication errors - 401 Unauthorized
		case errors.ErrCodeUnauthorized:
		case errors.ErrCodeInvalidAccessToken:
		case errors.ErrCodeAccessTokenExpired:
		case errors.ErrCodeInvalidCredentials:
			return http.StatusUnauthorized, appErr

		// Permission errors - 403 Forbidden
		case errors.ErrCodeForbidden:
			return http.StatusForbidden, appErr

		// Resource not found errors - 404 Not Found
		case errors.ErrCodeNotFound:
		case errors.ErrCodeResourceNotFound:
		case errors.ErrCodeDatabaseNotFound:
		case errors.ErrCodeWalletNotFound:
		case errors.ErrCodeUserNotFound:
		case errors.ErrCodeKeyNotFound:
		case errors.ErrCodeTransactionNotFound:
		case errors.ErrCodeContractNotFound:
			return http.StatusNotFound, appErr

		// Resource already exists errors - 409 Conflict
		case errors.ErrCodeAlreadyExists:
		case errors.ErrCodeResourceExists:
		case errors.ErrCodeWalletExists:
		case errors.ErrCodeUserExists:
		case errors.ErrCodeEmailExists:
		case errors.ErrCodeKeyExists:
		case errors.ErrCodeInsufficientFunds:
			return http.StatusConflict, appErr

		// Precondition failures - 412 Precondition Failed
		case errors.ErrCodeAddressMismatch:
		case errors.ErrCodeSignatureRecovery:
			return http.StatusPreconditionFailed, appErr

		// Rate limit errors - 429 Too Many Requests
		case errors.ErrCodeRateLimitExceeded:
			return http.StatusTooManyRequests, appErr

		// Timeout errors - 408 Request Timeout
		case errors.ErrCodeTimeout:
			return http.StatusRequestTimeout, appErr

		// Other application errors - 500 Internal Server Error
		case errors.ErrCodeInternalError:
		case errors.ErrCodeDatabaseError:
		case errors.ErrCodeBlockchainError:
		case errors.ErrCodeKeystoreError:
		case errors.ErrCodeWalletError:
		case errors.ErrCodeCryptoError:
		case errors.ErrCodeEncryptionError:
		case errors.ErrCodeDecryptionError:
		case errors.ErrCodeExplorerError:
		case errors.ErrCodeSigningError:
		case errors.ErrCodeTransactionFailed:
		case errors.ErrCodeRPCError:
		case errors.ErrCodeWalletOperationFailed:
		case errors.ErrCodeOperationFailed:
		case errors.ErrCodeTransactionSyncFailed:
		case errors.ErrCodeChainNotSupported:
		case errors.ErrCodeServiceUnavailable:
		case errors.ErrCodeExplorerRequestFailed:
		default:
			return http.StatusInternalServerError, appErr
		}
	}
	// Fallback for untyped errors
	appErr := &errors.AppError{
		Code:    errors.ErrCodeInternalError,
		Message: err.Error(),
		Err:     err,
	}
	return http.StatusInternalServerError, appErr
}
