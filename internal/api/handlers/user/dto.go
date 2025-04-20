package user

import (
	"strconv"
	"time"

	"vault0/internal/services/user"
)

// CreateUserRequest represents data needed to create a user
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserRequest represents data for updating a user
type UpdateUserRequest struct {
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Password string `json:"password,omitempty" binding:"omitempty,min=8"`
}

// ListUsersRequest defines the query parameters for listing users
type ListUsersRequest struct {
	NextToken string `form:"next_token"`
	Limit     *int   `form:"limit" binding:"omitempty,min=1"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a user model to a user response
func ToResponse(user *user.User) *UserResponse {
	return &UserResponse{
		ID:        strconv.FormatInt(user.ID, 10),
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// ToResponseList converts a slice of user models to a slice of user responses
func ToResponseList(users []*user.User) []*UserResponse {
	responses := make([]*UserResponse, len(users))
	for i, user := range users {
		responses[i] = ToResponse(user)
	}
	return responses
}
