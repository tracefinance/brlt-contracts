package blockchain

import (
	"time"
	"vault0/internal/types"
)

// Blockchain represents a blockchain record in the database
type Blockchain struct {
	ChainType     types.ChainType `db:"chain_type"`     // Primary key
	ChainID       int64           `db:"chain_id"`       // Chain ID from the network
	WalletID      string          `db:"wallet_id"`      // Associated wallet ID
	DeactivatedAt *time.Time      `db:"deactivated_at"` // Deactivation timestamp
	CreatedAt     time.Time       `db:"created_at"`     // Creation timestamp
}

// IsActive returns true if the blockchain is active (not deactivated)
func (b *Blockchain) IsActive() bool {
	return b.DeactivatedAt == nil
}
