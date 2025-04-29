package transaction

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto" // Need crypto for Keccak256

	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Precompute ERC20 transfer method ID locally for dispatcher logic
var erc20TransferMethodID = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

// evmMapper implements the Mapper interface for EVM-based transactions.
type evmMapper struct {
	tokenStore tokenstore.TokenStore
	logger     logger.Logger
	abiUtils   ABIUtils
}

// NewMapper creates a new instance of the EVM transaction mapper.
func NewMapper(tokenStore tokenstore.TokenStore, log logger.Logger, abiUtils ABIUtils) Mapper { // Added abiUtils param
	return &evmMapper{
		tokenStore: tokenStore,
		logger:     log.With(logger.String("component", "transaction_mapper")),
		abiUtils:   abiUtils,
	}
}

// --- Interface Implementation ---

// ToERC20Transfer implements Mapper.
func (m *evmMapper) ToERC20Transfer(ctx context.Context, tx *types.Transaction) (*types.ERC20Transfer, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}
	if tx.Type != types.TransactionTypeContractCall {
		return nil, errors.NewMappingError(tx.Hash, "transaction type is not contract_call")
	}
	if tx.BaseTransaction.To == "" {
		return nil, errors.NewMappingError(tx.Hash, "contract address (tx.To) is empty for ERC20 transfer")
	}
	if len(tx.Data) < 4 {
		return nil, errors.NewMappingError(tx.Hash, "transaction data too short for method ID")
	}

	methodID := m.abiUtils.ExtractMethodID(tx.Data)

	// Load ABI by address for this specific contract
	contractAddr := common.HexToAddress(tx.BaseTransaction.To)
	erc20ABI, err := m.abiUtils.LoadABIByAddress(ctx, contractAddr)
	if err != nil {
		// Log the error but potentially create a generic mapping error
		m.logger.Error("Failed to load ABI by address for potential ERC20 transfer",
			logger.String("address", contractAddr.Hex()),
			logger.String("tx_hash", tx.Hash),
			logger.Error(err),
		)
		// Return a mapping error indicating ABI load failure
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to load ABI for address %s: %v", contractAddr.Hex(), err))
	}

	// Check if methodID matches ERC20 transfer in the loaded ABI
	transferMethod, err := m.abiUtils.GetMethodFromABI(erc20ABI, methodID)
	if err != nil || transferMethod.Name != "transfer" { // Also check name for certainty
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("method ID %x does not match ERC20 transfer in ABI for %s", methodID, contractAddr.Hex()))
	}

	// Parse input using abiUtils (data without selector)
	parsedArgs, err := m.abiUtils.ParseContractInput(transferMethod, tx.Data[4:])
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to parse ERC20 transfer input: %v", err))
	}

	// Extract arguments using abiUtils helpers
	recipientAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "to")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get recipient address: %v", err))
	}

	amountBigInt, err := m.abiUtils.GetBigIntFromArgs(parsedArgs, "value")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get transfer amount: %v", err))
	}

	// Resolve token details remains the same
	tokenInfo, err := m.tokenStore.GetToken(ctx, tx.BaseTransaction.To)
	if err != nil {
		m.logger.Warn("Failed to resolve token details for ERC20 transfer",
			logger.String("chain", string(tx.BaseTransaction.Chain)),
			logger.String("token_address", tx.BaseTransaction.To),
			logger.String("tx_hash", tx.Hash),
			logger.Error(err),
		)
		tokenInfo = &types.Token{Address: tx.BaseTransaction.To, Symbol: "UNKNOWN", Decimals: 0}
	}

	// Create the specific type remains the same
	erc20Tx := &types.ERC20Transfer{
		Transaction:  *tx,
		TokenAddress: tokenInfo.Address,
		TokenSymbol:  tokenInfo.Symbol,
		Recipient:    recipientAddr.String(),
		Amount:       amountBigInt.ToBigInt(),
	}
	erc20Tx.Type = types.TransactionTypeERC20Transfer

	return erc20Tx, nil
}

// Shared logic for mapping MultiSig transactions
func (m *evmMapper) mapMultiSigTx(ctx context.Context, tx *types.Transaction, expectedMethodName string, argNames []string) (map[string]interface{}, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}
	if tx.Type != types.TransactionTypeContractCall {
		return nil, errors.NewMappingError(tx.Hash, "transaction type is not contract_call")
	}
	if len(tx.Data) < 4 {
		// Special case for recovery which might have short data
		if expectedMethodName == "requestRecovery" {
			return make(map[string]interface{}), nil // Return empty map, validation happens later
		}
		return nil, errors.NewMappingError(tx.Hash, "transaction data too short for method ID")
	}

	methodID := m.abiUtils.ExtractMethodID(tx.Data)

	// Use LoadABIByName
	multiSigABI, err := m.abiUtils.LoadABIByName(ctx, ABITypeMultiSig) // Corrected call
	if err != nil {
		return nil, fmt.Errorf("failed to load MultiSig ABI: %w", err)
	}

	method, err := m.abiUtils.GetMethodFromABI(multiSigABI, methodID)
	if err != nil || method.Name != expectedMethodName {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("method ID %x does not match MultiSig %s", methodID, expectedMethodName))
	}

	parsedArgs, err := m.abiUtils.ParseContractInput(method, tx.Data[4:])
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to parse MultiSig %s input: %v", expectedMethodName, err))
	}

	// Optional: Validate expected args exist (could be done in specific funcs too)
	for _, name := range argNames {
		if _, ok := parsedArgs[name]; !ok {
			return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("missing expected argument '%s' for %s", name, expectedMethodName))
		}
	}

	return parsedArgs, nil
}

// ToMultiSigWithdrawalRequest implements Mapper.
func (m *evmMapper) ToMultiSigWithdrawalRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigWithdrawalRequest, error) {
	parsedArgs, err := m.mapMultiSigTx(ctx, tx, "requestWithdrawal", []string{"token", "recipient", "amount", "withdrawalNonce"})
	if err != nil {
		return nil, err
	}

	tokenAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "token")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get token address: %v", err))
	}
	recipientAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "recipient")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get recipient address: %v", err))
	}
	amountBigInt, err := m.abiUtils.GetBigIntFromArgs(parsedArgs, "amount")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get withdrawal amount: %v", err))
	}
	withdrawalNonce, err := m.abiUtils.GetUint64FromArgs(parsedArgs, "withdrawalNonce")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get withdrawal nonce: %v", err))
	}

	multiSigTx := &types.MultiSigWithdrawalRequest{
		Transaction:     *tx,
		Token:           tokenAddr.String(),
		Amount:          amountBigInt.ToBigInt(),
		Recipient:       recipientAddr.String(),
		WithdrawalNonce: withdrawalNonce,
	}
	multiSigTx.Type = types.TransactionTypeMultiSigWithdrawalRequest

	return multiSigTx, nil
}

// ToMultiSigSignWithdrawal implements Mapper.
func (m *evmMapper) ToMultiSigSignWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigSignWithdrawal, error) {
	parsedArgs, err := m.mapMultiSigTx(ctx, tx, "signWithdrawal", []string{"requestID"})
	if err != nil {
		return nil, err
	}

	requestIDBytes32, err := m.abiUtils.GetBytes32FromArgs(parsedArgs, "requestID")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get request ID: %v", err))
	}

	multiSigTx := &types.MultiSigSignWithdrawal{
		Transaction: *tx,
		RequestID:   requestIDBytes32,
	}
	multiSigTx.Type = types.TransactionTypeMultiSigSignWithdrawal

	return multiSigTx, nil
}

// ToMultiSigExecuteWithdrawal implements Mapper.
func (m *evmMapper) ToMultiSigExecuteWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigExecuteWithdrawal, error) {
	parsedArgs, err := m.mapMultiSigTx(ctx, tx, "executeWithdrawal", []string{"requestID"})
	if err != nil {
		return nil, err
	}

	requestIDBytes32, err := m.abiUtils.GetBytes32FromArgs(parsedArgs, "requestID")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get request ID: %v", err))
	}

	multiSigTx := &types.MultiSigExecuteWithdrawal{
		Transaction: *tx,
		RequestID:   requestIDBytes32,
	}
	multiSigTx.Type = types.TransactionTypeMultiSigExecuteWithdrawal

	return multiSigTx, nil
}

// ToMultiSigAddSupportedToken implements Mapper.
func (m *evmMapper) ToMultiSigAddSupportedToken(ctx context.Context, tx *types.Transaction) (*types.MultiSigAddSupportedToken, error) {
	parsedArgs, err := m.mapMultiSigTx(ctx, tx, "addSupportedToken", []string{"token"})
	if err != nil {
		return nil, err
	}

	tokenAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "token")
	if err != nil {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("failed to get token address: %v", err))
	}

	multiSigTx := &types.MultiSigAddSupportedToken{
		Transaction: *tx,
		Token:       tokenAddr.String(),
	}
	multiSigTx.Type = types.TransactionTypeMultiSigAddSupportedToken

	return multiSigTx, nil
}

// ToMultiSigRecoveryRequest implements Mapper.
func (m *evmMapper) ToMultiSigRecoveryRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigRecoveryRequest, error) {
	_, err := m.mapMultiSigTx(ctx, tx, "requestRecovery", []string{}) // No arguments expected
	if err != nil {
		// Check if it was the specific error for short data but okay for recovery
		if errors.IsError(err, errors.ErrCodeMappingError) && len(tx.Data) < 4 {
			// This case is handled within mapMultiSigTx, allows proceeding
			m.logger.Warn("MultiSig recovery request transaction has data shorter than 4 bytes", logger.String("tx_hash", tx.Hash))
		} else {
			return nil, err // Return other mapping errors
		}
	}
	// No arguments to extract

	multiSigTx := &types.MultiSigRecoveryRequest{
		Transaction: *tx,
	}
	multiSigTx.Type = types.TransactionTypeMultiSigRecoveryRequest

	return multiSigTx, nil
}

// ToTypedTransaction implements Mapper.
func (m *evmMapper) ToTypedTransaction(ctx context.Context, tx *types.Transaction) (any, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction or base transaction is nil", "tx")
	}

	if tx.Type != types.TransactionTypeContractCall || len(tx.Data) < 4 {
		return tx, nil // Not a mappable contract call
	}

	methodID := m.abiUtils.ExtractMethodID(tx.Data)

	// Check for ERC20 transfer *first* by specific method ID
	if bytes.Equal(methodID, erc20TransferMethodID) {
		// If it matches the signature, attempt the specific mapping which loads ABI by address
		mappedTx, err := m.ToERC20Transfer(ctx, tx)
		if err == nil {
			return mappedTx, nil
		} else {
			// Log the mapping error but fall through to check MultiSig
			m.logger.Warn("Method ID matched ERC20 transfer, but mapping failed",
				logger.String("tx_hash", tx.Hash),
				logger.String("address", tx.BaseTransaction.To),
				logger.Error(err),
			)
		}
	}

	// Try mapping to MultiSig methods (this uses LoadABIByName internally now)
	multiSigABI, err := m.abiUtils.LoadABIByName(ctx, ABITypeMultiSig) // Use LoadABIByName with type
	if err == nil {
		if method, err := m.abiUtils.GetMethodFromABI(multiSigABI, methodID); err == nil {
			switch method.Name {
			case "requestWithdrawal":
				return m.ToMultiSigWithdrawalRequest(ctx, tx)
			case "signWithdrawal":
				return m.ToMultiSigSignWithdrawal(ctx, tx)
			case "executeWithdrawal":
				return m.ToMultiSigExecuteWithdrawal(ctx, tx)
			case "addSupportedToken":
				return m.ToMultiSigAddSupportedToken(ctx, tx)
			case "requestRecovery":
				return m.ToMultiSigRecoveryRequest(ctx, tx)
			}
		}
	} else {
		m.logger.Error("Failed to load MultiSig ABI for mapping check", logger.Error(err))
	}

	// If no known method matched
	m.logger.Debug("Transaction method not recognized by mapper",
		logger.String("tx_hash", tx.Hash),
		logger.String("method_id", common.Bytes2Hex(methodID)),
	)
	return tx, nil // Return original tx if no specific type is identified
}
