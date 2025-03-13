package blockchain

import (
	"time"
	"vault0/internal/services/blockchain"
	"vault0/internal/types"
)

// BlockchainResponse represents a blockchain in API responses
type BlockchainResponse struct {
	ChainType     types.ChainType `json:"chain_type"`
	ChainID       int64           `json:"chain_id"`
	WalletID      string          `json:"wallet_id"`
	IsActive      bool            `json:"is_active"`
	DeactivatedAt *time.Time      `json:"deactivated_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
}

// ToResponse converts a blockchain model to a response DTO
func ToResponse(b *blockchain.Blockchain) *BlockchainResponse {
	return &BlockchainResponse{
		ChainType:     b.ChainType,
		ChainID:       b.ChainID,
		WalletID:      b.WalletID,
		IsActive:      b.IsActive(),
		DeactivatedAt: b.DeactivatedAt,
		CreatedAt:     b.CreatedAt,
	}
}

// ToResponseList converts a list of blockchain models to response DTOs
func ToResponseList(blockchains []*blockchain.Blockchain) []*BlockchainResponse {
	responses := make([]*BlockchainResponse, len(blockchains))
	for i, b := range blockchains {
		responses[i] = ToResponse(b)
	}
	return responses
}
