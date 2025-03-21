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
	ID        int64           `json:"id"`
	Address   string          `json:"address"`
	ChainType types.ChainType `json:"chain_type"`
	Symbol    string          `json:"symbol"`
	Decimals  uint8           `json:"decimals"`
	Type      string          `json:"type"`
}

// WalletBalancesResponse represents the balance information for a wallet
type WalletBalancesResponse struct {
	Wallet   *WalletResponse         `json:"wallet"`
	Balances []*TokenBalanceResponse `json:"balances"`
}

// PagedWalletsResponse represents a response with a list of wallets
type PagedWalletsResponse struct {
	Items   []*WalletResponse `json:"items"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasMore bool              `json:"has_more"`
}

// ToResponse converts a wallet model to a wallet response
func ToResponse(wallet *wallet.Wallet, nativeToken *types.Token) *WalletResponse {
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
func ToResponseList(wallets []*wallet.Wallet, tokensMap map[types.ChainType]*types.Token) []*WalletResponse {
	responses := make([]*WalletResponse, len(wallets))
	for i, w := range wallets {
		// Get native token for this wallet's chain
		nativeToken, ok := tokensMap[w.ChainType]
		if !ok {
			// Fallback to direct conversion if token not found
			nativeToken = &types.Token{Decimals: 18} // Default to 18 decimals
		}
		responses[i] = ToResponse(w, nativeToken)
	}
	return responses
}

// ToPagedResponse converts a Page of wallet models to a PagedWalletsResponse
func ToPagedResponse(page *types.Page[*wallet.Wallet], tokensMap map[types.ChainType]*types.Token) *PagedWalletsResponse {
	return &PagedWalletsResponse{
		Items:   ToResponseList(page.Items, tokensMap),
		Limit:   page.Limit,
		Offset:  page.Offset,
		HasMore: page.HasMore,
	}
}

// ToTokenResponse converts a Token to a TokenResponse
func ToTokenResponse(token *types.Token) TokenResponse {
	return TokenResponse{
		ID:        token.ID,
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

// ToWalletBalancesResponse converts a wallet and token balances to a wallet balances response
func ToWalletBalancesResponse(wallet *wallet.Wallet, balances []*wallet.TokenBalanceData) *WalletBalancesResponse {
	// Find the native token in the balances
	var nativeToken *types.Token
	for _, balance := range balances {
		if balance.Token.IsNative() {
			nativeToken = balance.Token
			break
		}
	}

	// If no native token found, use a default
	if nativeToken == nil {
		nativeToken = &types.Token{Decimals: 18} // Default to 18 decimals as fallback
	}

	return &WalletBalancesResponse{
		Wallet:   ToResponse(wallet, nativeToken),
		Balances: ToTokenBalanceResponseList(balances),
	}
}
