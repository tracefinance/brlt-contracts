package errors

import (
	"encoding/json"
	"fmt"
)

// AppError represents a structured application error
type AppError struct {
	// Code is a unique identifier for the error type
	Code string `json:"code"`
	// Message is a human-readable error message
	Message string `json:"message"`
	// Details contains additional error context (optional)
	Details map[string]any `json:"details,omitempty"`
	// Err is the underlying error (not exposed in JSON)
	Err error `json:"-"`
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *AppError) Unwrap() error {
	return e.Err
}

// MarshalJSON implements json.Marshaler
func (e *AppError) MarshalJSON() ([]byte, error) {
	type Alias AppError
	return json.Marshal(&struct {
		*Alias
		Error string `json:"error"`
	}{
		Alias: (*Alias)(e),
		Error: e.Error(),
	})
}

// Is implements error matching for errors.Is
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}
