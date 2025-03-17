package middleares

import (
	"vault0/internal/errors"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ErrorMapper is a function type that maps an error to an HTTP status code and response body
type ErrorMapper func(err error) (int, any)

// ErrorHandler manages error mapping with an optional custom mapper
type ErrorHandler struct {
	mapper        ErrorMapper // Optional custom mapper
	defaultMapper ErrorMapper // Fallback default mapper
}

// NewErrorHandler creates a new ErrorHandler with an optional custom mapper
func NewErrorHandler(mapper ErrorMapper) *ErrorHandler {
	return &ErrorHandler{
		mapper:        mapper,
		defaultMapper: DefaultErrorMapper,
	}
}

// Middleware returns a Gin middleware function that handles errors
func (h *ErrorHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Process the request

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Handle binding errors from c.ShouldBindJSON
			if bindErr, ok := err.(*gin.Error); ok && bindErr.Type == gin.ErrorTypeBind {
				err = &errors.AppError{
					Code:    errors.ErrCodeInvalidRequest,
					Message: "Invalid request format",
					Err:     bindErr,
				}
				// Handle validation errors from c.ShouldBindJSON
			} else if validationErrors, ok := err.(validator.ValidationErrors); ok {
				details := make(map[string]any)
				for _, verr := range validationErrors {
					details[verr.Field()] = verr.Tag()
				}
				err = &errors.AppError{
					Code:    errors.ErrCodeValidationError,
					Message: "Invalid request format",
					Details: details,
				}
				return
			}
			// Try the custom mapper if provided
			if h.mapper != nil {
				if status, body := h.mapper(err); body != nil {
					c.JSON(status, body)
					return
				}
			}
			// Fall back to the default mapper
			status, body := h.defaultMapper(err)
			c.JSON(status, body)
		}
	}
}
