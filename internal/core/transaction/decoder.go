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
	// DecodeERC20Transfer attempts to convert a Transaction into an ERC20Transfer.
	// Returns an error if the transaction is not a valid ERC20 transfer call.
	DecodeERC20Transfer(ctx context.Context, tx *types.Transaction) (*types.ERC20Transfer, error)

	// DecodeMultiSigWithdrawalRequest attempts to convert a Transaction into a MultiSigWithdrawalRequest.
	// Returns an error if the transaction is not a valid MultiSig withdrawal request call.
	DecodeMultiSigWithdrawalRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigWithdrawalRequest, error)

	// DecodeMultiSigSignWithdrawal attempts to convert a Transaction into a MultiSigSignWithdrawal.
	// Returns an error if the transaction is not a valid MultiSig sign withdrawal call.
	DecodeMultiSigSignWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigSignWithdrawal, error)

	// DecodeMultiSigExecuteWithdrawal attempts to convert a Transaction into a MultiSigExecuteWithdrawal.
	// Returns an error if the transaction is not a valid MultiSig execute withdrawal call.
	DecodeMultiSigExecuteWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigExecuteWithdrawal, error)

	// DecodeMultiSigAddSupportedToken attempts to convert a Transaction into a MultiSigAddSupportedToken.
	// Returns an error if the transaction is not a valid MultiSig add token call.
	DecodeMultiSigAddSupportedToken(ctx context.Context, tx *types.Transaction) (*types.MultiSigAddSupportedToken, error)

	// DecodeMultiSigRecoveryRequest attempts to convert a Transaction into a MultiSigRecoveryRequest.
	// Returns an error if the transaction is not a valid MultiSig recovery request call.
	DecodeMultiSigRecoveryRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigRecoveryRequest, error)

	// DecodeTransaction acts as the primary dispatcher.
	// It attempts to identify the contract method from the transaction data
	// and calls the appropriate specific conversion method (e.g., ToERC20Transfer).
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
