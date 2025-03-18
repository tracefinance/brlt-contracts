package token

import "vault0/internal/types"

// TokenResponse is the token data returned in responses
type TokenResponse struct {
	ID        int64           `json:"id"`
	Address   string          `json:"address"`
	ChainType types.ChainType `json:"chain_type"`
	Symbol    string          `json:"symbol"`
	Decimals  uint8           `json:"decimals"`
	Type      types.TokenType `json:"type"`
}

// AddTokenRequest is the request body for adding a token
type AddTokenRequest struct {
	Address   string          `json:"address" binding:"required"`
	ChainType types.ChainType `json:"chain_type" binding:"required"`
	Symbol    string          `json:"symbol" binding:"required"`
	Decimals  uint8           `json:"decimals" binding:"required"`
	Type      types.TokenType `json:"type" binding:"required"`
}

// TokenListResponse is the paginated response for token listing
type TokenListResponse struct {
	Items []TokenResponse `json:"items"`
	Total int64           `json:"total"`
}
