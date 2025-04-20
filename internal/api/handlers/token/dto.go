package token

import (
	"vault0/internal/types"
)

// TokenResponse is the token data returned in responses
type TokenResponse struct {
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

// ListTokensRequest defines the query parameters for listing tokens
type ListTokensRequest struct {
	ChainType string `form:"chain_type"`
	TokenType string `form:"token_type"`
	NextToken string `form:"next_token"`
	Limit     *int   `form:"limit" binding:"omitempty,min=1"`
}

// TokenToResponse converts a token to a token response
func TokenToResponse(token types.Token) TokenResponse {
	return TokenResponse{
		Address:   token.Address,
		ChainType: token.ChainType,
		Symbol:    token.Symbol,
		Decimals:  token.Decimals,
		Type:      token.Type,
	}
}
