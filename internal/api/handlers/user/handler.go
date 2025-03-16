package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/errors"
	"vault0/internal/services/user"
)

// Handler handles user-related HTTP requests
type Handler struct {
	userService user.Service
}

// NewHandler creates a new user handler
func NewHandler(userService user.Service) *Handler {
	return &Handler{
		userService: userService,
	}
}

// CreateUser handles POST /users
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewInvalidRequestError("Invalid request body"))
		return
	}

	createdUser, err := h.userService.Create(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Return the service error directly without wrapping
		var appErr *errors.AppError
		if e, ok := err.(*errors.AppError); ok {
			appErr = e
		} else {
			// If it's not an AppError, wrap it as an internal error
			appErr = errors.NewInternalError(err)
		}

		// Map error codes to HTTP status codes
		status := http.StatusInternalServerError
		if appErr.Code == errors.ErrCodeInvalidInput {
			status = http.StatusBadRequest
		} else if appErr.Code == errors.ErrCodeEmailExists {
			status = http.StatusConflict
		}

		c.JSON(status, appErr)
		return
	}

	c.JSON(http.StatusCreated, ToResponse(createdUser))
}

// UpdateUser handles PUT /users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, errors.NewInvalidRequestError("Invalid request body"))
		return
	}

	updatedUser, err := h.userService.Update(c.Request.Context(), id, req.Email, req.Password)
	if err != nil {
		// Return the service error directly without wrapping
		var appErr *errors.AppError
		if e, ok := err.(*errors.AppError); ok {
			appErr = e
		} else {
			// If it's not an AppError, wrap it as an internal error
			appErr = errors.NewInternalError(err)
		}

		// Map error codes to HTTP status codes
		status := http.StatusInternalServerError
		if appErr.Code == errors.ErrCodeUserNotFound {
			status = http.StatusNotFound
		} else if appErr.Code == errors.ErrCodeInvalidInput {
			status = http.StatusBadRequest
		} else if appErr.Code == errors.ErrCodeEmailExists {
			status = http.StatusConflict
		}

		c.JSON(status, appErr)
		return
	}

	c.JSON(http.StatusOK, ToResponse(updatedUser))
}

// DeleteUser handles DELETE /users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		// Return the service error directly without wrapping
		var appErr *errors.AppError
		if e, ok := err.(*errors.AppError); ok {
			appErr = e
		} else {
			// If it's not an AppError, wrap it as an internal error
			appErr = errors.NewInternalError(err)
		}

		// Map error codes to HTTP status codes
		status := http.StatusInternalServerError
		if appErr.Code == errors.ErrCodeUserNotFound {
			status = http.StatusNotFound
		}

		c.JSON(status, appErr)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUser handles GET /users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	foundUser, err := h.userService.Get(c.Request.Context(), id)
	if err != nil {
		// Return the service error directly without wrapping
		var appErr *errors.AppError
		if e, ok := err.(*errors.AppError); ok {
			appErr = e
		} else {
			// If it's not an AppError, wrap it as an internal error
			appErr = errors.NewInternalError(err)
		}

		// Map error codes to HTTP status codes
		status := http.StatusInternalServerError
		if appErr.Code == errors.ErrCodeUserNotFound {
			status = http.StatusNotFound
		}

		c.JSON(status, appErr)
		return
	}

	c.JSON(http.StatusOK, ToResponse(foundUser))
}

// ListUsers handles GET /users
func (h *Handler) ListUsers(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil || pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	users, total, err := h.userService.List(c.Request.Context(), page, pageSize)
	if err != nil {
		// Return the service error directly without wrapping
		var appErr *errors.AppError
		if e, ok := err.(*errors.AppError); ok {
			appErr = e
		} else {
			// If it's not an AppError, wrap it as an internal error
			appErr = errors.NewInternalError(err)
		}

		c.JSON(http.StatusInternalServerError, appErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"page":  page,
		"size":  pageSize,
		"users": ToResponseList(users),
	})
}

func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	userRoutes := router.Group("/users")
	userRoutes.POST("", h.CreateUser)
	userRoutes.PUT("/:id", h.UpdateUser)
	userRoutes.DELETE("/:id", h.DeleteUser)
	userRoutes.GET("/:id", h.GetUser)
	userRoutes.GET("", h.ListUsers)
}
