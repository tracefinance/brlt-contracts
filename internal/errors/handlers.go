package errors

// Handler module error codes
const (
	// Request validation errors
	ErrCodeValidationError  = "validation_error"
	ErrCodeInvalidRequest   = "invalid_request"
	ErrCodeMissingParameter = "missing_parameter"
	ErrCodeInvalidParameter = "invalid_parameter"

	// Authentication errors
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeForbidden    = "forbidden"
	ErrCodeInvalidToken = "invalid_token"
	ErrCodeTokenExpired = "token_expired"

	// Response errors
	ErrCodeInternalError      = "internal_error"
	ErrCodeServiceUnavailable = "service_unavailable"
	ErrCodeTimeout            = "timeout"
)

// NewValidationError creates an error for request validation failures
func NewValidationError(details map[string]any) *AppError {
	return &AppError{
		Code:    ErrCodeValidationError,
		Message: "Request validation failed",
		Details: details,
	}
}

// NewInvalidRequestError creates an error for malformed requests
func NewInvalidRequestError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidRequest,
		Message: message,
	}
}

// NewMissingParameterError creates an error for missing required parameters
func NewMissingParameterError(param string) *AppError {
	return &AppError{
		Code:    ErrCodeMissingParameter,
		Message: "Missing required parameter: " + param,
	}
}

// NewInvalidParameterError creates an error for invalid parameter values
func NewInvalidParameterError(param string, reason string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidParameter,
		Message: "Invalid parameter: " + param,
		Details: map[string]any{
			"parameter": param,
			"reason":    reason,
		},
	}
}

// NewUnauthorizedError creates an error for unauthorized access
func NewUnauthorizedError() *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: "Authentication required",
	}
}

// NewForbiddenError creates an error for forbidden access
func NewForbiddenError() *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: "Access forbidden",
	}
}

// NewInvalidTokenError creates an error for invalid authentication tokens
func NewInvalidTokenError() *AppError {
	return &AppError{
		Code:    ErrCodeInvalidToken,
		Message: "Invalid authentication token",
	}
}

// NewTokenExpiredError creates an error for expired authentication tokens
func NewTokenExpiredError() *AppError {
	return &AppError{
		Code:    ErrCodeTokenExpired,
		Message: "Authentication token has expired",
	}
}

// NewInternalError creates an error for internal server errors
func NewInternalError(err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternalError,
		Message: "Internal server error",
		Err:     err,
	}
}

// NewServiceUnavailableError creates an error for service unavailability
func NewServiceUnavailableError(service string) *AppError {
	return &AppError{
		Code:    ErrCodeServiceUnavailable,
		Message: service + " service is currently unavailable",
	}
}

// NewTimeoutError creates an error for request timeouts
func NewTimeoutError() *AppError {
	return &AppError{
		Code:    ErrCodeTimeout,
		Message: "Request timed out",
	}
}
