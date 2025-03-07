package wallet

import (
	"context"
	"math/big"
	"vault0/internal/types"
)

// Wallet is the interface for interacting with blockchain wallets
type Wallet interface {
	// ChainType returns the blockchain type
	ChainType() types.ChainType

	// DeriveAddress derives a wallet address from a public key
	DeriveAddress(ctx context.Context, publicKey []byte) (string, error)

	// CreateNativeTransaction creates a new native currency transaction
	CreateNativeTransaction(ctx context.Context, fromAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error)

	// CreateTokenTransaction creates a new token transaction
	CreateTokenTransaction(ctx context.Context, fromAddress, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error)

	// SignTransaction signs a transaction with the specified key
	SignTransaction(ctx context.Context, keyID string, tx *types.Transaction) ([]byte, error)
}
