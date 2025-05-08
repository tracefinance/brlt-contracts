package transaction

import (
	"context"

	"vault0/internal/services/token"
	"vault0/internal/types"
)

// GetTokenAddressFromTransaction extracts the token address from a transaction based on its type.
// For ERC20 transfers, it returns the token address.
// For MultiSig transactions, it returns the token address if available.
// For other transaction types, it attempts to return the native token address.
// Returns empty string if no token address can be determined.
func GetTokenAddressFromTransaction(tx types.CoreTransaction) string {
	if tx == nil {
		return ""
	}

	// Extract token addresses based on transaction type
	switch typedTx := tx.(type) {
	case *types.ERC20Transfer:
		return typedTx.TokenAddress
	case *types.MultiSigWithdrawalRequest:
		return typedTx.TokenAddress
	case *types.MultiSigExecuteWithdrawal:
		// If we can extract token from execute withdrawal, add it here
		return ""
	default:
		// For other transaction types, try native token
		nativeToken, err := types.NewNativeToken(tx.GetChainType())
		if err == nil && nativeToken != nil {
			return nativeToken.Address
		}
	}

	return ""
}

// GetTokenForTransaction returns the appropriate token for a transaction.
// It first attempts to get the token from the provided address (if not empty).
// If no address is provided, it extracts the token address from the transaction.
// If token lookup fails or the address is empty/zero, falls back to native token.
// If native token lookup fails, returns a default token with 18 decimals.
//
// Parameters:
// - ctx: Context for token service calls
// - tx: Transaction to get token for
// - tokenService: Service for looking up tokens
// - tokenAddress: Optional token address (if already known)
//
// Returns a *types.Token that is guaranteed to be non-nil.
func GetTokenForTransaction(ctx context.Context, tx types.CoreTransaction, tokenService token.Service, tokenAddress string) *types.Token {
	if tx == nil {
		return &types.Token{Decimals: 18}
	}

	// If token address wasn't provided, try to extract it from the transaction
	if tokenAddress == "" {
		tokenAddress = GetTokenAddressFromTransaction(tx)
	}

	// If we have a non-zero address, try to get the token
	if tokenAddress != "" && tokenAddress != types.ZeroAddress {
		token, err := tokenService.GetToken(ctx, tokenAddress)
		if err == nil && token != nil {
			return token
		}
	}

	// Fall back to native token if no token address or lookup failed
	nativeToken, err := types.NewNativeToken(tx.GetChainType())
	if err != nil {
		// Fall back to default token if native token lookup fails
		return &types.Token{Decimals: 18}
	}

	return nativeToken
}
