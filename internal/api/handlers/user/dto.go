package user

import (
	"time"
	"vault0/internal/services/user"
)

// CreateUserRequest represents data needed to create a user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=8"`
}

// UpdateUserRequest represents data for updating a user
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Password string `json:"password,omitempty" binding:"omitempty,min=8"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts a user model to a user response
func ToResponse(user *user.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID,
		Username:  user.Username,
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
