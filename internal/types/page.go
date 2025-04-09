package types

// Page represents a paginated response
type Page[T any] struct {
	// Items contains the page items
	Items []T `json:"items"`
	// Offset is the starting index of the page
	Offset int `json:"offset"`
	// Limit is the maximum number of items per page
	Limit int `json:"limit"`
	// Whether there are more pages available
	HasMore bool `json:"has_more"`
}

// NewPage creates a new Page
func NewPage[T any](items []T, offset, limit int) *Page[T] {
	hasMore := limit > 0 && len(items) > limit
	pageItems := items

	if hasMore {
		pageItems = items[:limit] // Slice back to the requested limit
	}

	return &Page[T]{
		Items:   pageItems,
		Offset:  offset,
		Limit:   limit,
		HasMore: hasMore,
	}
}
