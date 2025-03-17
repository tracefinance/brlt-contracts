package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
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
		c.Error(errors.NewInvalidRequestError("Invalid request body"))
		return
	}

	createdUser, err := h.userService.Create(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ToResponse(createdUser))
}

// UpdateUser handles PUT /users/:id
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	updatedUser, err := h.userService.Update(c.Request.Context(), id, req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToResponse(updatedUser))
}

// DeleteUser handles DELETE /users/:id
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUser handles GET /users/:id
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	foundUser, err := h.userService.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToResponse(foundUser))
}

// ListUsers handles GET /users
func (h *Handler) ListUsers(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	page, err := h.userService.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToPagedResponse(page))
}

func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Create error handler middleware
	errorHandler := middleares.NewErrorHandler(nil)

	// Apply middleware to user routes group
	userRoutes := router.Group("/users")
	userRoutes.Use(errorHandler.Middleware())

	// Setup routes
	userRoutes.POST("", h.CreateUser)
	userRoutes.PUT("/:id", h.UpdateUser)
	userRoutes.DELETE("/:id", h.DeleteUser)
	userRoutes.GET("/:id", h.GetUser)
	userRoutes.GET("", h.ListUsers)
}
