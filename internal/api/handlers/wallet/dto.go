package wallet

import (
	"time"

	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// @Description Request model for creating a new wallet
type CreateWalletRequest struct {
	// The blockchain network type (e.g., ethereum, bitcoin)
	ChainType types.ChainType `json:"chain_type" binding:"required" example:"ethereum"`
	// A user-friendly name for the wallet
	Name string `json:"name" binding:"required" example:"My ETH Wallet"`
	// Optional key-value pairs for categorizing or adding metadata to the wallet
	Tags map[string]string `json:"tags,omitempty" example:"{\"purpose\":\"defi\",\"environment\":\"production\"}"`
}

// @Description Request model for updating an existing wallet
type UpdateWalletRequest struct {
	// Updated user-friendly name for the wallet
	Name string `json:"name" example:"My Updated ETH Wallet"`
	// Updated key-value pairs for categorizing or adding metadata to the wallet
	Tags map[string]string `json:"tags,omitempty" example:"{\"purpose\":\"defi\",\"environment\":\"production\"}"`
}

// @Description Response model containing wallet details
type WalletResponse struct {
	// Unique identifier for the wallet
	ID int64 `json:"id" example:"1"`
	// Unique identifier for the associated key in the keystore
	KeyID string `json:"key_id" example:"wallet_key_e8a1b8f7"`
	// The blockchain network type
	ChainType types.ChainType `json:"chain_type" example:"ethereum"`
	// The wallet's blockchain address
	Address string `json:"address" example:"0x71C7656EC7ab88b098defB751B7401B5f6d8976F"`
	// User-friendly name for the wallet
	Name string `json:"name" example:"My ETH Wallet"`
	// Key-value pairs for categorizing or adding metadata to the wallet
	Tags map[string]string `json:"tags,omitempty" example:"{\"purpose\":\"defi\",\"environment\":\"production\"}"`
	// Current wallet balance in the native currency, formatted as a string with appropriate decimals
	Balance string `json:"balance" example:"1.234567890000000000"`
	// Timestamp when the wallet was created
	CreatedAt time.Time `json:"created_at" example:"2023-01-01T12:00:00Z"`
	// Timestamp of the last wallet update
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-02T12:00:00Z"`
}

// @Description Response model containing token balance details
type TokenBalanceResponse struct {
	// Information about the token
	Token TokenResponse `json:"token"`
	// Current token balance, formatted as a string with appropriate decimals
	Balance string `json:"balance" example:"100.000000"`
	// Timestamp when the balance was last updated
	UpdatedAt time.Time `json:"updated_at" example:"2023-01-02T12:00:00Z"`
}

// @Description Response model containing token details
type TokenResponse struct {
	// The token's contract address (empty for native tokens)
	Address string `json:"address" example:"0xdAC17F958D2ee523a2206206994597C13D831ec7"`
	// The blockchain network type where the token exists
	ChainType types.ChainType `json:"chain_type" example:"ethereum"`
	// The token's symbol
	Symbol string `json:"symbol" example:"USDT"`
	// Number of decimal places for the token
	Decimals uint8 `json:"decimals" example:"6"`
	// Type of token (e.g., native, erc20)
	Type string `json:"type" example:"erc20"`
}

// @Description Paginated response model containing a list of wallets
type PagedWalletsResponse struct {
	// List of wallet objects
	Items []*WalletResponse `json:"items"`
	// Maximum number of items requested
	Limit int `json:"limit" example:"10"`
	// Number of items skipped for pagination
	Offset int `json:"offset" example:"0"`
	// Indicates if there are more items available beyond this page
	HasMore bool `json:"has_more" example:"true"`
}

func ToResponse(wallet *wallet.Wallet) *WalletResponse {
	nativeToken, err := wallet.GetToken()
	if err != nil {
		nativeToken = &types.Token{Decimals: 18}
	}

	balanceFloat := nativeToken.ToBigFloat(wallet.Balance)

	return &WalletResponse{
		ID:        wallet.ID,
		KeyID:     wallet.KeyID,
		ChainType: wallet.ChainType,
		Address:   wallet.Address,
		Name:      wallet.Name,
		Tags:      wallet.Tags,
		Balance:   balanceFloat.Text('f', int(nativeToken.Decimals)),
		CreatedAt: wallet.CreatedAt,
		UpdatedAt: wallet.UpdatedAt,
	}
}

func ToResponseList(wallets []*wallet.Wallet) []*WalletResponse {
	responses := make([]*WalletResponse, len(wallets))
	for i, w := range wallets {
		responses[i] = ToResponse(w)
	}
	return responses
}

func ToPagedResponse(page *types.Page[*wallet.Wallet]) *PagedWalletsResponse {
	return &PagedWalletsResponse{
		Items:   ToResponseList(page.Items),
		Limit:   page.Limit,
		Offset:  page.Offset,
		HasMore: page.HasMore,
	}
}

func ToTokenResponse(token *types.Token) TokenResponse {
	return TokenResponse{
		Address:   token.Address,
		ChainType: token.ChainType,
		Symbol:    token.Symbol,
		Decimals:  token.Decimals,
		Type:      string(token.Type),
	}
}

func ToTokenBalanceResponse(tokenBalance *wallet.TokenBalanceData) *TokenBalanceResponse {
	balanceFloat := tokenBalance.Token.ToBigFloat(tokenBalance.Balance)

	return &TokenBalanceResponse{
		Token:     ToTokenResponse(tokenBalance.Token),
		Balance:   balanceFloat.Text('f', int(tokenBalance.Token.Decimals)),
		UpdatedAt: tokenBalance.UpdatedAt,
	}
}

func ToTokenBalanceResponseList(tokenBalances []*wallet.TokenBalanceData) []*TokenBalanceResponse {
	responses := make([]*TokenBalanceResponse, len(tokenBalances))
	for i, tb := range tokenBalances {
		responses[i] = ToTokenBalanceResponse(tb)
	}
	return responses
}
