package transaction

import (
	"context"
	"vault0/internal/types"
)

// Mapper defines the interface for converting generic transactions
// into specific, type-safe transaction structs by parsing ABI data.
type Mapper interface {
	// ToERC20Transfer attempts to convert a Transaction into an ERC20Transfer.
	// Returns an error if the transaction is not a valid ERC20 transfer call.
	ToERC20Transfer(ctx context.Context, tx *types.Transaction) (*types.ERC20Transfer, error)

	// ToMultiSigWithdrawalRequest attempts to convert a Transaction into a MultiSigWithdrawalRequest.
	// Returns an error if the transaction is not a valid MultiSig withdrawal request call.
	ToMultiSigWithdrawalRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigWithdrawalRequest, error)

	// ToMultiSigSignWithdrawal attempts to convert a Transaction into a MultiSigSignWithdrawal.
	// Returns an error if the transaction is not a valid MultiSig sign withdrawal call.
	ToMultiSigSignWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigSignWithdrawal, error)

	// ToMultiSigExecuteWithdrawal attempts to convert a Transaction into a MultiSigExecuteWithdrawal.
	// Returns an error if the transaction is not a valid MultiSig execute withdrawal call.
	ToMultiSigExecuteWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigExecuteWithdrawal, error)

	// ToMultiSigAddSupportedToken attempts to convert a Transaction into a MultiSigAddSupportedToken.
	// Returns an error if the transaction is not a valid MultiSig add token call.
	ToMultiSigAddSupportedToken(ctx context.Context, tx *types.Transaction) (*types.MultiSigAddSupportedToken, error)

	// ToMultiSigRecoveryRequest attempts to convert a Transaction into a MultiSigRecoveryRequest.
	// Returns an error if the transaction is not a valid MultiSig recovery request call.
	ToMultiSigRecoveryRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigRecoveryRequest, error)

	// ToTypedTransaction acts as the primary dispatcher.
	// It attempts to identify the contract method from the transaction data
	// and calls the appropriate specific conversion method (e.g., ToERC20Transfer).
	// It returns the specific transaction struct (as `any`) or an error if identification
	// or parsing fails, or if the method is unknown.
	ToTypedTransaction(ctx context.Context, tx *types.Transaction) (any, error)
}
