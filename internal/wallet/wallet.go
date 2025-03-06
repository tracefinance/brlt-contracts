package wallet

import (
	"context"
	"math/big"
	"vault0/internal/common"
)

// Common errors
var (
	ErrUnsupportedChain    = common.ErrUnsupportedChain
	ErrInvalidAddress      = common.ErrInvalidAddress
	ErrInvalidAmount       = common.ErrInvalidAmount
	ErrTransactionFailed   = common.ErrTransactionFailed
	ErrInsufficientBalance = common.ErrInsufficientBalance
)

// Wallet defines the interface for wallet operations
type Wallet interface {
	// ChainType returns the blockchain type
	ChainType() common.ChainType

	// DeriveAddress derives a wallet address from a public key
	DeriveAddress(ctx context.Context, publicKey []byte) (string, error)

	// CreateNativeTransaction creates a native currency transaction without broadcasting
	CreateNativeTransaction(ctx context.Context, fromAddress, toAddress string, amount *big.Int, options common.TransactionOptions) (*common.Transaction, error)

	// CreateTokenTransaction creates an ERC20 token transaction without broadcasting
	CreateTokenTransaction(ctx context.Context, fromAddress, tokenAddress, toAddress string, amount *big.Int, options common.TransactionOptions) (*common.Transaction, error)

	// SignTransaction signs a transaction
	SignTransaction(ctx context.Context, keyID string, tx *common.Transaction) ([]byte, error)
}
