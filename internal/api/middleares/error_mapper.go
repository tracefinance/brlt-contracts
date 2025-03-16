package middleares

import (
	"net/http"
	"vault0/internal/errors"
)

// DefaultErrorMapper provides fallback behavior for mapping errors to HTTP responses
func DefaultErrorMapper(err error) (int, any) {
	if appErr, ok := err.(*errors.AppError); ok {
		switch appErr.Code {
		case "validation_error":
			return http.StatusBadRequest, appErr
		case "unauthorized":
			return http.StatusUnauthorized, appErr
		case "not_found":
			return http.StatusNotFound, appErr
		default:
			return http.StatusInternalServerError, appErr
		}
	}
	// Fallback for untyped errors
	return http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"}
}
