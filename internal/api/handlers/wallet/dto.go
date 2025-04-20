package wallet

import (
	"strconv"
	"time"

	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// @Description Request model for creating a new wallet
type CreateWalletRequest struct {
	ChainType types.ChainType   `json:"chain_type" binding:"required" example:"ethereum"`
	Name      string            `json:"name" binding:"required" example:"My ETH Wallet"`
	Tags      map[string]string `json:"tags,omitempty" example:"{\"purpose\":\"defi\",\"environment\":\"production\"}"`
}

// @Description Request model for updating an existing wallet
type UpdateWalletRequest struct {
	Name string            `json:"name" example:"My Updated ETH Wallet"`
	Tags map[string]string `json:"tags,omitempty" example:"{\"purpose\":\"defi\",\"environment\":\"production\"}"`
}

// ListWalletsRequest defines the query parameters for listing wallets
type ListWalletsRequest struct {
	NextToken string `form:"next_token"`
	Limit     *int   `form:"limit" binding:"omitempty,min=0"`
}

// @Description Response model containing wallet details
type WalletResponse struct {
	ID              string            `json:"id" example:"1"`
	KeyID           string            `json:"key_id" example:"wallet_key_e8a1b8f7"`
	ChainType       types.ChainType   `json:"chain_type" example:"ethereum"`
	Address         string            `json:"address" example:"0x71C7656EC7ab88b098defB751B7401B5f6d8976F"`
	Name            string            `json:"name" example:"My ETH Wallet"`
	Tags            map[string]string `json:"tags,omitempty" example:"{\"purpose\":\"defi\",\"environment\":\"production\"}"`
	Balance         string            `json:"balance" example:"1.234567890000000000"`
	LastBlockNumber int64             `json:"last_block_number,omitempty" example:"12345678"`
	CreatedAt       time.Time         `json:"created_at" example:"2023-01-01T12:00:00Z"`
	UpdatedAt       time.Time         `json:"updated_at" example:"2023-01-02T12:00:00Z"`
}

// @Description Response model containing token balance details
type TokenBalanceResponse struct {
	Token     TokenResponse `json:"token"`
	Balance   string        `json:"balance" example:"100.000000"`
	UpdatedAt time.Time     `json:"updated_at" example:"2023-01-02T12:00:00Z"`
}

// @Description Response model containing token details
type TokenResponse struct {
	Address   string          `json:"address" example:"0xdAC17F958D2ee523a2206206994597C13D831ec7"`
	ChainType types.ChainType `json:"chain_type" example:"ethereum"`
	Symbol    string          `json:"symbol" example:"USDT"`
	Decimals  uint8           `json:"decimals" example:"6"`
	Type      string          `json:"type" example:"erc20"`
}

func ToResponse(wallet *wallet.Wallet) *WalletResponse {
	nativeToken, err := wallet.GetToken()
	if err != nil {
		nativeToken = &types.Token{Decimals: 18}
	}

	balanceFloat := nativeToken.ToBigFloat(wallet.Balance.ToBigInt())

	return &WalletResponse{
		ID:              strconv.FormatInt(wallet.ID, 10),
		KeyID:           wallet.KeyID,
		ChainType:       wallet.ChainType,
		Address:         wallet.Address,
		Name:            wallet.Name,
		Tags:            wallet.Tags,
		Balance:         balanceFloat.Text('f', int(nativeToken.Decimals)),
		LastBlockNumber: wallet.LastBlockNumber,
		CreatedAt:       wallet.CreatedAt,
		UpdatedAt:       wallet.UpdatedAt,
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
