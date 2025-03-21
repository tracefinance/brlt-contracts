package wallet

import (
	"math/big"
	"time"

	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// CreateWalletRequest represents a request to create a wallet
type CreateWalletRequest struct {
	ChainType types.ChainType   `json:"chain_type" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	Tags      map[string]string `json:"tags,omitempty"`
}

// UpdateWalletRequest represents a request to update a wallet
type UpdateWalletRequest struct {
	Name string            `json:"name"`
	Tags map[string]string `json:"tags,omitempty"`
}

// WalletResponse represents a wallet response
type WalletResponse struct {
	ID        int64             `json:"id"`
	KeyID     string            `json:"key_id"`
	ChainType types.ChainType   `json:"chain_type"`
	Address   string            `json:"address"`
	Name      string            `json:"name"`
	Tags      map[string]string `json:"tags,omitempty"`
	Balance   *big.Float        `json:"balance"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

// TokenBalanceResponse represents a token balance response
type TokenBalanceResponse struct {
	Token     TokenResponse `json:"token"`
	Balance   *big.Float    `json:"balance"`
	UpdatedAt time.Time     `json:"updated_at"`
}

// TokenResponse represents a token response
type TokenResponse struct {
	Address   string          `json:"address"`
	ChainType types.ChainType `json:"chain_type"`
	Symbol    string          `json:"symbol"`
	Decimals  uint8           `json:"decimals"`
	Type      string          `json:"type"`
}

// PagedWalletsResponse represents a response with a list of wallets
type PagedWalletsResponse struct {
	Items   []*WalletResponse `json:"items"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasMore bool              `json:"has_more"`
}

// ToResponse converts a wallet model to a wallet response
func ToResponse(wallet *wallet.Wallet) *WalletResponse {
	// Get the native token for this wallet
	nativeToken, err := wallet.GetToken()
	if err != nil {
		// If there's an error, use a default token with 18 decimals
		nativeToken = &types.Token{Decimals: 18}
	}

	// Convert big.Int to big.Float using the token's decimal places
	balanceFloat := nativeToken.ToBigFloat(wallet.Balance)

	return &WalletResponse{
		ID:        wallet.ID,
		KeyID:     wallet.KeyID,
		ChainType: wallet.ChainType,
		Address:   wallet.Address,
		Name:      wallet.Name,
		Tags:      wallet.Tags,
		Balance:   balanceFloat,
		CreatedAt: wallet.CreatedAt,
		UpdatedAt: wallet.UpdatedAt,
	}
}

// ToResponseList converts a slice of wallet models to a slice of wallet responses
func ToResponseList(wallets []*wallet.Wallet) []*WalletResponse {
	responses := make([]*WalletResponse, len(wallets))
	for i, w := range wallets {
		responses[i] = ToResponse(w)
	}
	return responses
}

// ToPagedResponse converts a Page of wallet models to a PagedWalletsResponse
func ToPagedResponse(page *types.Page[*wallet.Wallet]) *PagedWalletsResponse {
	return &PagedWalletsResponse{
		Items:   ToResponseList(page.Items),
		Limit:   page.Limit,
		Offset:  page.Offset,
		HasMore: page.HasMore,
	}
}

// ToTokenResponse converts a Token to a TokenResponse
func ToTokenResponse(token *types.Token) TokenResponse {
	return TokenResponse{
		Address:   token.Address,
		ChainType: token.ChainType,
		Symbol:    token.Symbol,
		Decimals:  token.Decimals,
		Type:      string(token.Type),
	}
}

// ToTokenBalanceResponse converts a TokenBalanceData model to a TokenBalanceResponse
func ToTokenBalanceResponse(tokenBalance *wallet.TokenBalanceData) *TokenBalanceResponse {
	// Convert big.Int to big.Float using the token's decimal places
	balanceFloat := tokenBalance.Token.ToBigFloat(tokenBalance.Balance)

	return &TokenBalanceResponse{
		Token:     ToTokenResponse(tokenBalance.Token),
		Balance:   balanceFloat,
		UpdatedAt: tokenBalance.UpdatedAt,
	}
}

// ToTokenBalanceResponseList converts a slice of token balance data to a slice of token balance responses
func ToTokenBalanceResponseList(tokenBalances []*wallet.TokenBalanceData) []*TokenBalanceResponse {
	responses := make([]*TokenBalanceResponse, len(tokenBalances))
	for i, tb := range tokenBalances {
		responses[i] = ToTokenBalanceResponse(tb)
	}
	return responses
}
