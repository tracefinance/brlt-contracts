package middleares

import (
	"net/http"
	"vault0/internal/errors"
)

// DefaultErrorMapper provides fallback behavior for mapping errors to HTTP responses
func DefaultErrorMapper(err error) (int, any) {
	if appErr, ok := err.(*errors.Vault0Error); ok {
		switch appErr.Code {
		// Input validation errors - 400 Bad Request
		case errors.ErrCodeInvalidInput,
			errors.ErrCodeValidationError,
			errors.ErrCodeInvalidRequest,
			errors.ErrCodeMissingParameter,
			errors.ErrCodeInvalidParameter,
			errors.ErrCodeInvalidWallet,
			errors.ErrCodeInvalidWalletConfig,
			errors.ErrCodeInvalidKey,
			errors.ErrCodeInvalidKeyType,
			errors.ErrCodeInvalidCurve,
			errors.ErrCodeInvalidSignature,
			errors.ErrCodeInvalidAddress,
			errors.ErrCodeInvalidTransaction,
			errors.ErrCodeInvalidNonce,
			errors.ErrCodeInvalidGasPrice,
			errors.ErrCodeInvalidGasLimit,
			errors.ErrCodeInvalidContractCall,
			errors.ErrCodeInvalidAmount,
			errors.ErrCodeInvalidContract,
			errors.ErrCodeInvalidBlockchainConfig,
			errors.ErrCodeInvalidAPIKey,
			errors.ErrCodeInvalidExplorerResponse,
			errors.ErrCodeInvalidEncryptionKey,
			errors.ErrCodeInvalidKeystore,
			errors.ErrCodeInvalidToken,
			errors.ErrCodeMissingKeyID,
			errors.ErrCodeMissingWalletAddress,
			errors.ErrCodeMissingAPIKey,
			// OAuth2 validation errors
			errors.ErrCodeInvalidScope,
			errors.ErrCodeUnsupportedGrantType,
			// New Validation Errors
			errors.ErrCodeInvalidBlockIdentifier,
			errors.ErrCodeInvalidEventSignature,
			errors.ErrCodeInvalidEventArgs,
			errors.ErrCodeUnsupportedEventArgType:
			return http.StatusBadRequest, appErr

		// Authentication errors - 401 Unauthorized
		case errors.ErrCodeUnauthorized,
			errors.ErrCodeInvalidAccessToken,
			errors.ErrCodeAccessTokenExpired,
			errors.ErrCodeInvalidCredentials,
			// OAuth2 authentication errors
			errors.ErrCodeInvalidClient,
			errors.ErrCodeInvalidGrant:
			return http.StatusUnauthorized, appErr

		// Permission errors - 403 Forbidden
		case errors.ErrCodeForbidden:
			return http.StatusForbidden, appErr

		// Resource not found errors - 404 Not Found
		case errors.ErrCodeNotFound,
			errors.ErrCodeResourceNotFound,
			errors.ErrCodeDatabaseNotFound,
			errors.ErrCodeWalletNotFound,
			errors.ErrCodeUserNotFound,
			errors.ErrCodeKeyNotFound,
			errors.ErrCodeTransactionNotFound,
			errors.ErrCodeContractNotFound,
			errors.ErrCodeSignerNotFound,
			errors.ErrCodeSignerAddressNotFound,
			// New Not Found Errors
			errors.ErrCodeTokenPriceNotFound,
			errors.ErrCodeBlockNotFound:
			return http.StatusNotFound, appErr

		// Resource already exists errors - 409 Conflict
		case errors.ErrCodeAlreadyExists,
			errors.ErrCodeResourceExists,
			errors.ErrCodeWalletExists,
			errors.ErrCodeUserExists,
			errors.ErrCodeEmailExists,
			errors.ErrCodeKeyExists,
			errors.ErrCodeInsufficientFunds,
			errors.ErrCodeKeyInUseByWallet:
			return http.StatusConflict, appErr

		// Precondition failures - 412 Precondition Failed
		case errors.ErrCodeAddressMismatch,
			errors.ErrCodeSignatureRecovery:
			return http.StatusPreconditionFailed, appErr

		// Rate limit errors - 429 Too Many Requests
		case errors.ErrCodeRateLimitExceeded:
			return http.StatusTooManyRequests, appErr

		// Timeout errors - 408 Request Timeout
		case errors.ErrCodeTimeout:
			return http.StatusRequestTimeout, appErr

		// Other application errors - 500 Internal Server Error
		case errors.ErrCodeInternalError,
			errors.ErrCodeDatabaseError,
			errors.ErrCodeBlockchainError,
			errors.ErrCodeKeystoreError,
			errors.ErrCodeWalletError,
			errors.ErrCodeCryptoError,
			errors.ErrCodeEncryptionError,
			errors.ErrCodeDecryptionError,
			errors.ErrCodeExplorerError,
			errors.ErrCodeSigningError,
			errors.ErrCodeTransactionFailed,
			errors.ErrCodeRPCError,
			errors.ErrCodeWalletOperationFailed,
			errors.ErrCodeOperationFailed,
			errors.ErrCodeTransactionSyncFailed,
			errors.ErrCodeChainNotSupported,
			errors.ErrCodeServiceUnavailable,
			errors.ErrCodeExplorerRequestFailed,
			// OAuth2 server errors
			errors.ErrCodeServerOAuth2Error,
			// New Internal/Server Errors
			errors.ErrCodePriceFeedUpdateFailed,
			errors.ErrCodeDataConversionFailed,
			errors.ErrCodeConfiguration,
			errors.ErrCodeInvalidTokenBalance,
			errors.ErrCodePriceFeedRequestFailed,
			errors.ErrCodeInvalidPriceFeedResponse,
			errors.ErrCodePriceFeedProviderNotSupported,
			errors.ErrCodeLogTopicIndexOutOfBounds,
			errors.ErrCodeLogTopicInvalidFormat:
			return http.StatusInternalServerError, appErr

		default:
			return http.StatusInternalServerError, appErr
		}
	}
	// Fallback for untyped errors
	appErr := &errors.Vault0Error{
		Code:    errors.ErrCodeInternalError,
		Message: err.Error(),
		Err:     err,
	}
	return http.StatusInternalServerError, appErr
}
