package errors

// Handler module error codes
const (
	// Request validation errors
	ErrCodeValidationError  = "validation_error"
	ErrCodeInvalidRequest   = "invalid_request"
	ErrCodeMissingParameter = "missing_parameter"
	ErrCodeInvalidParameter = "invalid_parameter"

	// Authentication errors
	ErrCodeUnauthorized       = "unauthorized"
	ErrCodeForbidden          = "forbidden"
	ErrCodeInvalidAccessToken = "invalid_access_token"
	ErrCodeAccessTokenExpired = "access_token_expired"

	// Response errors
	ErrCodeInternalError      = "internal_error"
	ErrCodeServiceUnavailable = "service_unavailable"
	ErrCodeTimeout            = "timeout"
)

// NewValidationError creates an error for request validation failures
func NewValidationError(details map[string]any) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeValidationError,
		Message: "Request validation failed",
		Details: details,
	}
}

// NewInvalidRequestError creates an error for malformed requests
func NewInvalidRequestError(message string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeInvalidRequest,
		Message: message,
	}
}

// NewMissingParameterError creates an error for missing required parameters
func NewMissingParameterError(param string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeMissingParameter,
		Message: "Missing required parameter: " + param,
	}
}

// NewInvalidParameterError creates an error for invalid parameter values
func NewInvalidParameterError(param string, reason string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeInvalidParameter,
		Message: "Invalid parameter: " + param,
		Details: map[string]any{
			"parameter": param,
			"reason":    reason,
		},
	}
}

// NewUnauthorizedError creates an error for unauthorized access
func NewUnauthorizedError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeUnauthorized,
		Message: "Authentication required",
	}
}

// NewForbiddenError creates an error for forbidden access
func NewForbiddenError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeForbidden,
		Message: "Access forbidden",
	}
}

// NewInvalidAccessTokenError creates an error for invalid authentication tokens
func NewInvalidAccessTokenError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeInvalidAccessToken,
		Message: "Invalid authentication token",
	}
}

// NewAccessTokenExpiredError creates an error for expired authentication tokens
func NewAccessTokenExpiredError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeAccessTokenExpired,
		Message: "Authentication token has expired",
	}
}

// NewInternalError creates an error for internal server errors
func NewInternalError(err error) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeInternalError,
		Message: "Internal server error",
		Err:     err,
	}
}

// NewServiceUnavailableError creates an error for service unavailability
func NewServiceUnavailableError(service string) *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeServiceUnavailable,
		Message: service + " service is currently unavailable",
	}
}

// NewTimeoutError creates an error for request timeouts
func NewTimeoutError() *Vault0Error {
	return &Vault0Error{
		Code:    ErrCodeTimeout,
		Message: "Request timed out",
	}
}
