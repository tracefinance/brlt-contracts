package docs

// This file contains type definitions specifically for Swagger documentation
// These allow us to work around the fact that swag doesn't support Go generics

// KeyPagedResponse is a non-generic version of PagedResponse[KeyResponse]
// swagger:model KeyPagedResponse
type KeyPagedResponse struct {
	// The list of keys
	Items []KeyResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// TokenPagedResponse is a non-generic version of PagedResponse[TokenResponse]
// swagger:model TokenPagedResponse
type TokenPagedResponse struct {
	// The list of tokens
	Items []TokenResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// TransactionPagedResponse is a non-generic version of PagedResponse[TransactionResponse]
// swagger:model TransactionPagedResponse
type TransactionPagedResponse struct {
	// The list of transactions
	Items []TransactionResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// UserPagedResponse is a non-generic version of PagedResponse[UserResponse]
// swagger:model UserPagedResponse
type UserPagedResponse struct {
	// The list of users
	Items []UserResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// WalletPagedResponse is a non-generic version of PagedResponse[WalletResponse]
// swagger:model WalletPagedResponse
type WalletPagedResponse struct {
	// The list of wallets
	Items []WalletResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// VaultPagedResponse is a non-generic version of PagedResponse[VaultResponse]
// swagger:model VaultPagedResponse
type VaultPagedResponse struct {
	// The list of vaults
	Items []VaultResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// TokenPricePagedResponse is a non-generic version of PagedResponse[TokenPriceResponse]
// swagger:model TokenPricePagedResponse
type TokenPricePagedResponse struct {
	// The list of token prices
	Items []TokenPriceResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// SignerPagedResponse is a non-generic version of PagedResponse[SignerResponse]
// swagger:model SignerPagedResponse
type SignerPagedResponse struct {
	// The list of signers
	Items []SignerResponse `json:"items"`
	// Token for the next page
	NextToken string `json:"next_token,omitempty" example:"eyJjIjoiaWQiLCJ2IjoxMDAwfQ=="`
	// The limit used for the page
	Limit int `json:"limit" example:"10"`
}

// These are placeholders to make the file compile
// The actual implementations are in their respective handler packages

type KeyResponse struct{}
type TokenResponse struct{}
type TransactionResponse struct{}
type UserResponse struct{}
type WalletResponse struct{}
type VaultResponse struct{}
type TokenPriceResponse struct{}
type SignerResponse struct{}
