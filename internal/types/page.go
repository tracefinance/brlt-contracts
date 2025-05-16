package types

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"vault0/internal/errors"
)

// NextPageToken represents pagination state for token-based pagination
type NextPageToken struct {
	// Column is the database column used for cursor-based pagination
	Column string `json:"c"`
	// Value is the value to compare against for pagination
	Value any `json:"v"`
}

// GetValueInt64 attempts to assert the Value field to an int64.
// It handles cases where JSON unmarshalling might store numbers as float64.
// It returns the int64 value and true if successful, otherwise zero and false.
func (t *NextPageToken) GetValueInt64() (int64, bool) {
	if t == nil || t.Value == nil {
		return 0, false
	}

	switch v := t.Value.(type) {
	case int64:
		return v, true
	case float64:
		return int64(v), true
	case json.Number:
		i, err := v.Int64()
		if err == nil {
			return i, true
		}
		f, err := v.Float64()
		if err == nil {
			return int64(f), true
		}
		return 0, false
	default:
		return 0, false
	}
}

// EncodeNextPageToken converts a NextPageToken to an encoded string
func EncodeNextPageToken(token NextPageToken) (string, error) {
	data, err := json.Marshal(token)
	if err != nil {
		return "", errors.NewTokenEncodingFailedError(err)
	}
	return base64.URLEncoding.EncodeToString(data), nil
}

// DecodeNextPageToken parses an encoded string into a NextPageToken
// and validates that the token's column matches the expected column
func DecodeNextPageToken(tokenStr string, expectedColumn string) (*NextPageToken, error) {
	if tokenStr == "" {
		return nil, nil
	}

	data, err := base64.URLEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, errors.NewTokenDecodingFailedError(tokenStr, err)
	}

	var token NextPageToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, errors.NewInvalidPaginationTokenError(tokenStr, err)
	}

	// Validate that the token's column matches the expected column
	if token.Column != expectedColumn {
		return nil, errors.NewInvalidPaginationTokenError(
			tokenStr,
			fmt.Errorf("invalid column: expected '%s', got '%s'", expectedColumn, token.Column))
	}

	return &token, nil
}

// Page represents a paginated response
type Page[T any] struct {
	// Items contains the page items
	Items []T
	// NextToken is used for fetching the next page. Empty string for the last page.
	NextToken string
	// Limit is the maximum number of items per page
	Limit int
}

// NewPage creates a new Page with token-based pagination
// items should contain up to limit+1 elements to determine if there are more pages
// generateNextToken is a function that creates a token from the last item
// If limit is 0, all items are returned without pagination (no NextToken)
func NewPage[T any](items []T, limit int, generateNextToken func(T) *NextPageToken) *Page[T] {
	// Special case: when limit is 0, return all items without pagination
	if limit <= 0 {
		return &Page[T]{
			Items:     items,
			NextToken: "",
			Limit:     0,
		}
	}

	var nextToken string
	pageItems := items

	// Check if we have more items than the limit (we fetched limit+1)
	hasMore := len(items) > limit

	if hasMore && len(items) > 0 {
		// Get the last visible item (not the limit+1 item)
		lastItem := items[limit-1]

		// Generate token from the last visible item
		token := generateNextToken(lastItem)
		if token != nil {
			encoded, err := EncodeNextPageToken(*token)
			if err == nil {
				nextToken = encoded
			}
		}

		// Slice back to the requested limit
		pageItems = items[:limit]
	}

	return &Page[T]{
		Items:     pageItems,
		NextToken: nextToken,
		Limit:     limit,
	}
}
