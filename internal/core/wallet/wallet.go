package wallet

import (
	"context"
	"fmt"
	"math/big"
	"vault0/internal/config"
	"vault0/internal/core/keystore"
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
	Chain() types.Chain

	// DeriveAddress derives a wallet address
	DeriveAddress(ctx context.Context) (string, error)

	// CreateNativeTransaction creates a new native currency transaction
	CreateNativeTransaction(ctx context.Context, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error)

	// CreateTokenTransaction creates a new token transaction
	CreateTokenTransaction(ctx context.Context, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error)

	// SignTransaction signs a transaction with the wallet's key
	SignTransaction(ctx context.Context, tx *types.Transaction) ([]byte, error)
}

// NewWallet creates a new wallet instance for the specified chain type and key ID.
func NewWallet(ctx context.Context, keystore keystore.KeyStore, chains types.Chains, cfg *config.Config, chainType types.ChainType, keyID string) (Wallet, error) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		// Get chain struct from blockchain package
		chain := chains[chainType]

		// All EVM-compatible chains use the same implementation
		return NewEVMWallet(keystore, chain, keyID)
	default:
		return nil, fmt.Errorf("%w: %s", types.ErrUnsupportedChain, chainType)
	}
}
