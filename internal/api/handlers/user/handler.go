package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "vault0/internal/api/docs" // Required for Swagger documentation
	"vault0/internal/api/middleares"
	"vault0/internal/api/utils"
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

// CreateUser handles POST /users
// @Summary Create a new user
// @Description Create a new user with the given email and password
// @Tags users
// @Accept json
// @Produce json
// @Param user body CreateUserRequest true "User data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 409 {object} errors.Vault0Error "User already exists"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	createdUser, err := h.userService.CreateUser(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ToResponse(createdUser))
}

// UpdateUser handles PUT /users/:id
// @Summary Update a user
// @Description Update a user's information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body UpdateUserRequest true "User data to update"
// @Success 200 {object} UserResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "User not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /users/{id} [put]
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

	updatedUser, err := h.userService.UpdateUser(c.Request.Context(), id, req.Email, req.Password)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToResponse(updatedUser))
}

// DeleteUser handles DELETE /users/:id
// @Summary Delete a user
// @Description Delete a user by ID
// @Tags users
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "User not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	if err := h.userService.DeleteUser(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetUser handles GET /users/:id
// @Summary Get a user
// @Description Get a user by ID
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "User not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /users/{id} [get]
func (h *Handler) GetUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("id", "must be a valid integer"))
		return
	}

	foundUser, err := h.userService.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToResponse(foundUser))
}

// ListUsers handles GET /users
// @Summary List users
// @Description Get a paginated list of users
// @Tags users
// @Produce json
// @Param limit query int false "Number of items to return (default: 10, max: 100)" default(10)
// @Param next_token query string false "Token for fetching the next page"
// @Success 200 {object} docs.UserPagedResponse
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	var req ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errors.NewInvalidParameterError("query", "invalid query parameters format or value"))
		return
	}

	// Set default limit if not provided
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	page, err := h.userService.ListUsers(c.Request.Context(), limit, req.NextToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, utils.NewPagedResponse(page, ToResponse))
}
