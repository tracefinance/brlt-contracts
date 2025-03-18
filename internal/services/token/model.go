package token

import "vault0/internal/types"

// TokenFilter defines filtering options for token listing
type TokenFilter struct {
	ChainType *types.ChainType
	TokenType *types.TokenType
}
