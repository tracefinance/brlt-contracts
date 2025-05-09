package transaction

import (
	"context"
	"vault0/internal/core/abiutils"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Decoder defines the interface for converting generic transactions
// into specific, type-safe transaction structs by parsing ABI data.
type Decoder interface {
	// DecodeTransaction acts as the primary dispatcher.
	// It attempts to identify the contract method from the transaction data
	// and calls the appropriate specific conversion method (e.g., DecodeERC20Transfer).
	// It returns the specific transaction struct (as `any`) or an error if identification
	// or parsing fails, or if the method is unknown.
	DecodeTransaction(ctx context.Context, tx *types.Transaction) (types.CoreTransaction, error)
}

// NewDecoder creates a new instance of the EVM transaction mapper.
func NewDecoder(chainType types.ChainType, tokenStore tokenstore.TokenStore, log logger.Logger, abiUtils abiutils.ABIUtils) (Decoder, error) {
	switch chainType {
	case types.ChainTypeEthereum, types.ChainTypePolygon, types.ChainTypeBase:
		return NewEvmDecoder(tokenStore, log, abiUtils), nil
	default:
		return nil, errors.NewChainNotSupportedError(string(chainType))
	}
}
