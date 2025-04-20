package utils

import (
	"vault0/internal/types"
)

// PagedResponse is a generic paginated response for API endpoints
// It wraps any type of items with pagination metadata
type PagedResponse[T any] struct {
	// Items contains the page items
	Items []T `json:"items"`
	// NextToken is used for fetching the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// Limit is the maximum number of items per page
	Limit int `json:"limit" example:"10"`
}

// NewPagedResponse creates a new paged response from a Page
func NewPagedResponse[T any, R any](page *types.Page[T], transformFunc func(T) R) *PagedResponse[R] {
	items := make([]R, len(page.Items))
	for i, item := range page.Items {
		items[i] = transformFunc(item)
	}

	return &PagedResponse[R]{
		Items:     items,
		NextToken: page.NextToken,
		Limit:     page.Limit,
	}
}
