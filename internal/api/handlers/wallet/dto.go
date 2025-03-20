package wallet

import (
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
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
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
	return &WalletResponse{
		ID:        wallet.ID,
		KeyID:     wallet.KeyID,
		ChainType: wallet.ChainType,
		Address:   wallet.Address,
		Name:      wallet.Name,
		Tags:      wallet.Tags,
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
