package tokenprice

import "time"

// TokenPrice represents the stored price data for a token in the database.
// Note the use of `float64` for numerical fields which might require careful
// handling regarding precision, especially for financial data.
type TokenPrice struct {
	Symbol       string    `db:"symbol"`         // Token symbol (e.g., "BTC") - Primary Key
	Rank         int       `db:"rank"`           // Market cap rank
	PriceUSD     float64   `db:"price_usd"`      // Price in USD
	Supply       float64   `db:"supply"`         // Circulating supply
	MarketCapUSD float64   `db:"market_cap_usd"` // Market cap in USD
	VolumeUSD24h float64   `db:"volume_usd_24h"` // Trading volume in last 24h USD
	UpdatedAt    time.Time `db:"updated_at"`     // Timestamp when the data was last updated in the DB
}

// TokenPriceFilter defines criteria for filtering token prices
type TokenPriceFilter struct {
	// Symbols is an optional list of token symbols to filter by
	Symbols []string
}
