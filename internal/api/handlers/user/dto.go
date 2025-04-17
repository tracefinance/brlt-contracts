package user

import (
	"strconv"
	"time"
	"vault0/internal/services/user"
	"vault0/internal/types"
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

// UserResponse represents a user response
type UserResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PagedUsersResponse represents a paginated list of users
type PagedUsersResponse struct {
	Items   []*UserResponse `json:"items"`
	Limit   int             `json:"limit"`
	Offset  int             `json:"offset"`
	HasMore bool            `json:"has_more"`
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

// ToPagedResponse converts a Page of user models to a PagedUserResponse
func ToPagedResponse(page *types.Page[*user.User]) *PagedUsersResponse {
	return &PagedUsersResponse{
		Items:   ToResponseList(page.Items),
		Limit:   page.Limit,
		Offset:  page.Offset,
		HasMore: page.HasMore,
	}
}
