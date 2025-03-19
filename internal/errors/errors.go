package errors

import (
	"encoding/json"
	"fmt"
)

// Vault0Error represents a structured application error
type Vault0Error struct {
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
func (e *Vault0Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the underlying error
func (e *Vault0Error) Unwrap() error {
	return e.Err
}

// MarshalJSON implements json.Marshaler
func (e *Vault0Error) MarshalJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Is implements error matching for errors.Is
func (e *Vault0Error) Is(target error) bool {
	t, ok := target.(*Vault0Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}
