package types

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// Address and chain errors
	ErrInvalidAddress = errors.New("invalid address")
	ErrMissingRPCURL  = errors.New("missing RPC URL")

	// Transaction errors
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrTransactionFailed   = errors.New("transaction failed")
	ErrInsufficientBalance = errors.New("insufficient balance")
)

// UnsupportedChainError represents an error when an operation is attempted with an unsupported blockchain
type UnsupportedChainError struct {
	ChainType ChainType
}

// Error implements the error interface
func (e *UnsupportedChainError) Error() string {
	return fmt.Sprintf("unsupported blockchain: %s", e.ChainType)
}

// IsUnsupportedChainError checks if an error is an UnsupportedChainError
func IsUnsupportedChainError(err error) bool {
	_, ok := err.(*UnsupportedChainError)
	return ok
}
