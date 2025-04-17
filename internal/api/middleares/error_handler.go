package middleares

import (
	"regexp"
	"strings"
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

// toSnakeCase converts a camelCase or PascalCase string to snake_case
func toSnakeCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// Middleware returns a Gin middleware function that handles errors
func (h *ErrorHandler) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() // Process the request

		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err

			// Handle binding errors from c.ShouldBindJSON
			if bindErr, ok := err.(*gin.Error); ok && bindErr.Type == gin.ErrorTypeBind {
				err = &errors.Vault0Error{
					Code:    errors.ErrCodeInvalidRequest,
					Message: "Invalid request format",
					Err:     bindErr,
				}
			} else if validationErrors, ok := err.(validator.ValidationErrors); ok {
				details := make(map[string]any)
				for _, verr := range validationErrors {
					// Convert field name to snake_case
					fieldName := toSnakeCase(verr.Field())
					details[fieldName] = verr.Tag()
				}
				err = &errors.Vault0Error{
					Code:    errors.ErrCodeValidationError,
					Message: "Invalid request format",
					Details: details,
				}
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
