package errors

// Chain and address error codes
const (
	ErrCodeInvalidAddress = "invalid_address"
	ErrCodeMissingRPCURL  = "missing_rpc_url"
)

// Transaction error codes
const (
	ErrCodeInvalidAmount = "invalid_amount"
)

// NewInvalidAddressError creates an error for invalid blockchain addresses
func NewInvalidAddressError(address string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidAddress,
		Message: "Invalid address",
		Details: map[string]any{
			"address": address,
		},
	}
}

// NewMissingRPCURLError creates an error for missing RPC URL configuration
func NewMissingRPCURLError(chainType string) *AppError {
	return &AppError{
		Code:    ErrCodeMissingRPCURL,
		Message: "Missing RPC URL",
		Details: map[string]any{
			"chain_type": chainType,
		},
	}
}

// NewInvalidAmountError creates an error for invalid transaction amounts
func NewInvalidAmountError(amount string) *AppError {
	return &AppError{
		Code:    ErrCodeInvalidAmount,
		Message: "Invalid amount",
		Details: map[string]any{
			"amount": amount,
		},
	}
}
