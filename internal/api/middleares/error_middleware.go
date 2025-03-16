package middleares

import (
	"github.com/gin-gonic/gin"
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
