package tokenprice

import "time"

// TokenPriceResponse represents a single token price entry in API responses.
type TokenPriceResponse struct {
	Symbol       string    `json:"symbol"`
	Rank         int       `json:"rank"`
	PriceUSD     float64   `json:"price_usd"`
	Supply       float64   `json:"supply"`
	MarketCapUSD float64   `json:"market_cap_usd"`
	VolumeUSD24h float64   `json:"volume_usd_24h"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ListTokenPricesRequest defines query parameters for the list endpoint.
// We use pointers to distinguish between default values and not provided.
type ListTokenPricesRequest struct {
	Limit  *int `form:"limit" binding:"omitempty,min=1,max=100"`
	Offset *int `form:"offset" binding:"omitempty,min=0"`
}

// GetTokenPriceRequest defines path parameters for getting a single token.
type GetTokenPriceRequest struct {
	Symbol string `uri:"symbol" binding:"required"`
}
