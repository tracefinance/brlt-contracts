package wallet

import (
	"context"
	"math/big"
	"vault0/internal/core/blockchain"
	"vault0/internal/types"
)

// WalletInfo contains the wallet identification and address information
type WalletInfo struct {
	// KeyID is the ID of the key in the keystore
	KeyID string
	// Address is the derived blockchain address
	Address string
	// ChainType is the blockchain type
	ChainType types.ChainType
}

// Wallet is the interface for interacting with blockchain wallets
type Wallet interface {
	// Chain returns the blockchain chain information
	Chain() blockchain.Chain

	// DeriveAddress derives a wallet address
	DeriveAddress(ctx context.Context) (string, error)

	// CreateNativeTransaction creates a new native currency transaction
	CreateNativeTransaction(ctx context.Context, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error)

	// CreateTokenTransaction creates a new token transaction
	CreateTokenTransaction(ctx context.Context, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error)

	// SignTransaction signs a transaction with the wallet's key
	SignTransaction(ctx context.Context, tx *types.Transaction) ([]byte, error)
}
